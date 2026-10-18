//go:build integration

package identidad_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/GersonLujan97/danber/backend/internal/identidad"
)

type authResponse struct {
	Principal identidad.Principal `json:"principal"`
	Expires   time.Time           `json:"expira_en"`
	CSRF      string              `json:"csrf_token"`
}
type browser struct {
	server *httptest.Server
	client *http.Client
}

func newBrowser(t *testing.T, auth *identidad.Authentication) *browser {
	t.Helper()
	server := httptest.NewUnstartedServer(nil)
	h, err := auth.Handler(identidad.HTTPOptions{Origin: "http://" + server.Listener.Addr().String(), AllowInsecureLocal: true})
	if err != nil {
		t.Fatal(err)
	}
	server.Config.Handler = h
	server.Start()
	t.Cleanup(server.Close)
	return &browser{server: server, client: &http.Client{Timeout: 8 * time.Second}}
}
func (b *browser) request(method, path, body string, cookie *http.Cookie, csrf string) (*http.Response, []byte, error) {
	r, err := http.NewRequest(method, b.server.URL+path, strings.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-Danber-Client", "1")
	r.Header.Set("Origin", b.server.URL)
	if cookie != nil {
		r.AddCookie(cookie)
	}
	if csrf != "" {
		r.Header.Set("X-Danber-CSRF", csrf)
	}
	response, err := b.client.Do(r)
	if err != nil {
		return nil, nil, err
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(response.Body, 8192))
	return response, raw, err
}

const authCredentials = `{"usuario":"Ficticio","contrasena":"frase ficticia para probar"}`

func (b *browser) login(t *testing.T, previous *http.Cookie) (*http.Cookie, authResponse) {
	t.Helper()
	r, raw, err := b.request("POST", "/auth/login", authCredentials, previous, "")
	if err != nil {
		t.Fatal("HTTP login failed")
	}
	if r.StatusCode != 200 {
		t.Fatalf("login status=%d", r.StatusCode)
	}
	var value authResponse
	if err = json.Unmarshal(raw, &value); err != nil {
		t.Fatal("invalid login response")
	}
	cookies := r.Cookies()
	if len(cookies) != 1 || value.CSRF == "" {
		t.Fatal("missing cookie/CSRF")
	}
	if bytes.Contains(raw, []byte(cookies[0].Value)) || bytes.Contains(raw, []byte("frase ficticia")) {
		t.Fatal("bearer or password leaked")
	}
	return cookies[0], value
}
func (b *browser) status(t *testing.T, method, path string, cookie *http.Cookie, csrf string, want int) []byte {
	t.Helper()
	r, raw, err := b.request(method, path, "", cookie, csrf)
	if err != nil {
		t.Fatal("HTTP operation failed")
	}
	if r.StatusCode != want {
		t.Fatalf("%s %s status=%d want=%d", method, path, r.StatusCode, want)
	}
	if r.Header.Get("Cache-Control") != "no-store" {
		t.Fatal("cache control missing")
	}
	return raw
}

func TestAuthenticationHTTPWithCockroach(t *testing.T) {
	ctx, admin, pool := fixture(t, 10)
	user, err := identidad.NewInitialAccount(pool).Create(ctx, "Ficticio", []byte("frase ficticia para probar"), "operador_pruebas")
	if err != nil {
		t.Fatal(err)
	}
	auth, err := identidad.NewAuthentication(pool, identidad.SessionPolicy{Idle: 15 * time.Minute, Absolute: 8 * time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := admin.Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("generic_failures", func(t *testing.T) {
		b := newBrowser(t, auth)
		var expected []byte
		for _, body := range []string{strings.Replace(authCredentials, "Ficticio", "Inexistente", 1), strings.Replace(authCredentials, "frase ficticia para probar", "incorrecta", 1), authCredentials} {
			if body == authCredentials {
				exec(`UPDATE identidad.usuario SET estado='suspendido',version=version+1 WHERE id=$1`, user)
			}
			r, raw, err := b.request("POST", "/auth/login", body, nil, "")
			if err != nil {
				t.Fatal(err)
			}
			if r.StatusCode != 401 || len(r.Cookies()) != 0 {
				t.Fatal("failure contract differs")
			}
			if expected == nil {
				expected = raw
			} else if !bytes.Equal(expected, raw) {
				t.Fatal("account state disclosed")
			}
		}
		exec(`UPDATE identidad.usuario SET estado='activo',version=version+1 WHERE id=$1`, user)
		count(t, ctx, admin, "identidad.sesion", 0)
	})
	t.Run("hash_cookie_and_live_permissions", func(t *testing.T) {
		b := newBrowser(t, auth)
		cookie, value := b.login(t, nil)
		if cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode || !value.Principal.Allows("identidad.administrar") {
			t.Fatal("local cookie/claims contract failed")
		}
		hash := sha256.Sum256([]byte(cookie.Value))
		var protected bool
		if err := admin.QueryRow(ctx, `SELECT length(token_hash)=32 FROM identidad.sesion WHERE token_hash=$1`, hash[:]).Scan(&protected); err != nil || !protected {
			t.Fatal("SQL did not store opaque-token hash")
		}
		body := b.status(t, "GET", "/auth/session", cookie, "", 200)
		if bytes.Contains(body, []byte(cookie.Value)) {
			t.Fatal("session response leaked bearer")
		}
		var detail string
		if err := admin.QueryRow(ctx, `SELECT detalle::STRING FROM identidad.cambio_acceso WHERE operacion='sesion_abierta' ORDER BY fecha DESC LIMIT 1`).Scan(&detail); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(detail, cookie.Value) || strings.Contains(detail, "frase ficticia") || strings.Contains(detail, "argon2") {
			t.Fatal("audit contains secrets")
		}
	})
	t.Run("rotation_and_logout_csrf", func(t *testing.T) {
		b := newBrowser(t, auth)
		old, _ := b.login(t, nil)
		current, value := b.login(t, old)
		if old.Value == current.Value {
			t.Fatal("session fixation")
		}
		b.status(t, "GET", "/auth/session", old, "", 401)
		b.status(t, "POST", "/auth/logout", current, "invalid", 403)
		b.status(t, "GET", "/auth/session", current, "", 200)
		b.status(t, "POST", "/auth/logout", current, value.CSRF, 204)
		b.status(t, "GET", "/auth/session", current, "", 401)
		b.status(t, "POST", "/auth/logout", current, value.CSRF, 204)
	})
	t.Run("concurrent_logout_records_one_event", func(t *testing.T) {
		b := newBrowser(t, auth)
		cookie, value := b.login(t, nil)
		hash := sha256.Sum256([]byte(cookie.Value))
		var sessionID string
		if err := admin.QueryRow(ctx, `SELECT id::STRING FROM identidad.sesion WHERE token_hash=$1`, hash[:]).Scan(&sessionID); err != nil {
			t.Fatal(err)
		}
		start := make(chan struct{})
		codes := make(chan int, 4)
		var wg sync.WaitGroup
		for range 4 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				r, _, err := b.request("POST", "/auth/logout", "", cookie, value.CSRF)
				if err != nil {
					codes <- 0
				} else {
					codes <- r.StatusCode
				}
			}()
		}
		close(start)
		wg.Wait()
		close(codes)
		for code := range codes {
			if code != 204 {
				t.Fatalf("concurrent logout status=%d", code)
			}
		}
		var events int
		if err := admin.QueryRow(ctx, `SELECT count(*) FROM identidad.cambio_acceso WHERE operacion='sesion_cerrada' AND detalle->>'sesion_id'=$1`, sessionID).Scan(&events); err != nil || events != 1 {
			t.Fatal("logout audit duplicated or missing")
		}
	})
	t.Run("idle_absolute_and_no_sliding_absolute_expiry", func(t *testing.T) {
		b := newBrowser(t, auth)
		cookie, value := b.login(t, nil)
		hash := sha256.Sum256([]byte(cookie.Value))
		b.status(t, "GET", "/auth/session", cookie, "", 200)
		var expiration time.Time
		if err := admin.QueryRow(ctx, `SELECT expira_en FROM identidad.sesion WHERE token_hash=$1`, hash[:]).Scan(&expiration); err != nil || !expiration.Equal(value.Expires) {
			t.Fatal("absolute expiry was extended")
		}
		exec(`UPDATE identidad.sesion SET ultimo_uso_en=now()-INTERVAL '20 minutes' WHERE token_hash=$1`, hash[:])
		b.status(t, "GET", "/auth/session", cookie, "", 401)
		cookie, _ = b.login(t, nil)
		hash = sha256.Sum256([]byte(cookie.Value))
		exec(`UPDATE identidad.sesion SET expira_en=now() WHERE token_hash=$1`, hash[:])
		b.status(t, "GET", "/auth/session", cookie, "", 401)
	})
	t.Run("logout_all_revokes_existing_and_allows_new_login", func(t *testing.T) {
		b := newBrowser(t, auth)
		one, v := b.login(t, nil)
		two, _ := b.login(t, nil)
		b.status(t, "POST", "/auth/logout-all", one, v.CSRF, 204)
		b.status(t, "GET", "/auth/session", one, "", 401)
		b.status(t, "GET", "/auth/session", two, "", 401)
		fresh, _ := b.login(t, nil)
		b.status(t, "GET", "/auth/session", fresh, "", 200)
	})
	t.Run("concurrent_login_and_revoke_all_have_consistent_versions", func(t *testing.T) {
		b := newBrowser(t, auth)
		one, v := b.login(t, nil)
		start := make(chan struct{})
		codes := make(chan int, 2)
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			r, _, err := b.request("POST", "/auth/login", authCredentials, nil, "")
			if err != nil {
				codes <- 0
			} else {
				codes <- r.StatusCode
			}
		}()
		go func() {
			defer wg.Done()
			<-start
			r, _, err := b.request("POST", "/auth/logout-all", "", one, v.CSRF)
			if err != nil {
				codes <- 0
			} else {
				codes <- r.StatusCode
			}
		}()
		close(start)
		wg.Wait()
		close(codes)
		for code := range codes {
			if code != 200 && code != 204 {
				t.Fatalf("concurrent operation status=%d", code)
			}
		}
		var mismatches int
		if err := admin.QueryRow(ctx, `SELECT count(*) FROM identidad.sesion s JOIN identidad.usuario u ON u.id=s.usuario_id WHERE u.id=$1 AND s.revocado_en IS NULL AND s.usuario_version<>u.version`, user).Scan(&mismatches); err != nil || mismatches != 0 {
			t.Fatal("revocation lost a concurrent session/version")
		}
	})
	t.Run("suspension_version_and_password_timestamp_invalidate_access", func(t *testing.T) {
		b := newBrowser(t, auth)
		cookie, _ := b.login(t, nil)
		exec(`UPDATE identidad.usuario SET estado='suspendido',version=version+1 WHERE id=$1`, user)
		b.status(t, "GET", "/auth/session", cookie, "", 401)
		exec(`UPDATE identidad.usuario SET estado='activo',version=version+1 WHERE id=$1`, user)
		b.status(t, "GET", "/auth/session", cookie, "", 401)
		cookie, _ = b.login(t, nil)
		exec(`UPDATE identidad.credencial SET cambiado_en=now() WHERE usuario_id=$1`, user)
		b.status(t, "GET", "/auth/session", cookie, "", 401)
	})
	t.Run("failed_audit_rolls_back_session_and_revocation", func(t *testing.T) {
		b := newBrowser(t, auth)
		var before, after int
		admin.QueryRow(ctx, `SELECT count(*) FROM identidad.sesion`).Scan(&before)
		exec(`ALTER TABLE identidad.cambio_acceso ADD CONSTRAINT falla_login_prueba CHECK (operacion<>'sesion_abierta') NOT VALID`)
		r, _, err := b.request("POST", "/auth/login", authCredentials, nil, "")
		if err != nil || r.StatusCode != 503 {
			t.Fatal("audit failure must fail login")
		}
		admin.QueryRow(ctx, `SELECT count(*) FROM identidad.sesion`).Scan(&after)
		if before != after {
			t.Fatal("session left behind after audit failure")
		}
		exec(`ALTER TABLE identidad.cambio_acceso DROP CONSTRAINT falla_login_prueba`)
		cookie, v := b.login(t, nil)
		exec(`ALTER TABLE identidad.cambio_acceso ADD CONSTRAINT falla_revocar_prueba CHECK (operacion<>'sesiones_revocadas') NOT VALID`)
		b.status(t, "POST", "/auth/logout-all", cookie, v.CSRF, 503)
		b.status(t, "GET", "/auth/session", cookie, "", 200)
		exec(`ALTER TABLE identidad.cambio_acceso DROP CONSTRAINT falla_revocar_prueba`)
	})
	t.Run("permissions_are_reloaded_and_role_does_not_bypass_catalog", func(t *testing.T) {
		b := newBrowser(t, auth)
		cookie, _ := b.login(t, nil)
		exec(`DELETE FROM identidad.usuario_rol WHERE usuario_id=$1 AND rol_id IN (SELECT id FROM identidad.rol WHERE codigo='maestro')`, user)
		var value authResponse
		body := b.status(t, "GET", "/auth/session", cookie, "", 200)
		if json.Unmarshal(body, &value) != nil {
			t.Fatal("invalid principal")
		}
		if value.Principal.Allows("identidad.administrar") || !value.Principal.Allows("agenda.reservar") {
			t.Fatal("stale permissions retained")
		}
		exec(`DELETE FROM identidad.usuario_rol WHERE usuario_id=$1`, user)
		body = b.status(t, "GET", "/auth/session", cookie, "", 200)
		if json.Unmarshal(body, &value) != nil || len(value.Principal.Permissions) != 0 || value.Principal.Allows("agenda.reservar") {
			t.Fatal("unassigned permission accepted")
		}
	})
}

func TestSessionUpgradeFromNinePreservesExistingData(t *testing.T) {
	ctx, admin, pool := fixture(t, 9)
	_, err := identidad.NewInitialAccount(pool).Create(ctx, "Ficticio", []byte("frase ficticia para probar"), "operador_pruebas")
	if err != nil {
		t.Fatal(err)
	}
	auth, err := identidad.NewAuthentication(pool, identidad.SessionPolicy{Idle: 15 * time.Minute, Absolute: 8 * time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	if err = auth.CheckSchema(ctx); err == nil {
		t.Fatal("missing 010 not detected")
	}
	// A synthetic pre-upgrade session proves ALTER defaults on populated data.
	token := strings.Repeat("A", 43)
	hash := sha256.Sum256([]byte(token))
	if _, err = admin.Exec(ctx, `INSERT INTO identidad.sesion(usuario_id,token_hash,expira_en)
 SELECT id,$1,now()+INTERVAL '8 hours' FROM identidad.usuario`, hash[:]); err != nil {
		t.Fatal(err)
	}
	var before, after string
	if err = admin.QueryRow(ctx, `SELECT hash_clave FROM identidad.credencial`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	_, source, _, _ := runtime.Caller(0)
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(source), "../../../db/migrations/010_sesiones_y_version_acceso.sql"))
	if err != nil {
		t.Fatal(err)
	}
	applyStatements(t, ctx, admin, string(raw))
	if err = auth.CheckSchema(ctx); err != nil {
		t.Fatal(err)
	}
	if err = admin.QueryRow(ctx, `SELECT hash_clave FROM identidad.credencial`).Scan(&after); err != nil || before != after {
		t.Fatal("upgrade modified credential")
	}
	count(t, ctx, admin, "identidad.usuario", 1)
	count(t, ctx, admin, "identidad.sesion", 1)
	if _, err = auth.Authenticate(ctx, token); err != nil {
		t.Fatal("session defaults not applied to existing data")
	}
}
