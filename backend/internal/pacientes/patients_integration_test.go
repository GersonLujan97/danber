//go:build integration

package pacientes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	"github.com/GersonLujan97/danber/backend/internal/pacientes"
	"github.com/GersonLujan97/danber/backend/internal/pacientes/internal/app"
	"github.com/GersonLujan97/danber/backend/internal/pacientes/internal/domain"
	"github.com/GersonLujan97/danber/backend/internal/pacientes/internal/sqlstore"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type browser struct {
	server *httptest.Server
	cookie *http.Cookie
	csrf   string
}

func (b browser) call(method, path string, body any, key string) (int, []byte) {
	raw, _ := json.Marshal(body)
	r, _ := http.NewRequest(method, b.server.URL+path, bytes.NewReader(raw))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-Danber-Client", "1")
	r.Header.Set("Origin", b.server.URL)
	r.Header.Set("X-Danber-CSRF", b.csrf)
	r.Header.Set("Idempotency-Key", key)
	if b.cookie != nil {
		r.AddCookie(b.cookie)
	}
	resp, e := b.server.Client().Do(r)
	if e != nil {
		return 0, nil
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, data
}
func input(doc string) domain.Registration {
	return domain.Registration{Names: "Lucía", Surnames: "Ficticia", BirthDate: "1990-01-01", Address: "Dirección ficticia", Phone: "+51 111111111", Document: domain.Document{Type: "DNI", Number: doc}}
}
func key(n int) string { return fmt.Sprintf("00000000-0000-4000-8000-%012d", n) }
func guardian() *domain.Guardian {
	return &domain.Guardian{Names: "María", Surnames: "Ficticia", Document: domain.Document{Type: "DNI", Number: "90000001"}, Relationship: "Madre"}
}

func TestPatientsHTTPAndPersistence(t *testing.T) {
	ctx, admin, pool := fixture(t, 11)
	actor, e := identidad.NewInitialAccount(pool).Create(ctx, "Ficticio", []byte("frase ficticia para probar"), "operador_pruebas")
	if e != nil {
		t.Fatal(e)
	}
	auth, e := identidad.NewAuthentication(pool, identidad.SessionPolicy{Idle: 15 * time.Minute, Absolute: 8 * time.Hour})
	if e != nil {
		t.Fatal(e)
	}
	m, e := pacientes.New(pool)
	if e != nil {
		t.Fatal(e)
	}
	if e = m.CheckSchema(ctx); e != nil {
		t.Fatal(e)
	}
	server := httptest.NewUnstartedServer(nil)
	options := identidad.HTTPOptions{Origin: "http://" + server.Listener.Addr().String(), AllowInsecureLocal: true}
	authHandler, e := auth.Handler(options)
	if e != nil {
		t.Fatal(e)
	}
	patientHandler, e := m.Handler(auth, options)
	if e != nil {
		t.Fatal(e)
	}
	mux := http.NewServeMux()
	mux.Handle("/auth/", authHandler)
	mux.Handle("/pacientes", patientHandler)
	mux.Handle("/pacientes/", patientHandler)
	server.Config.Handler = mux
	server.Start()
	t.Cleanup(server.Close)
	b := browser{server: server}
	login, _ := http.NewRequest("POST", server.URL+"/auth/login", strings.NewReader(`{"usuario":"Ficticio","contrasena":"frase ficticia para probar"}`))
	login.Header.Set("Content-Type", "application/json")
	login.Header.Set("X-Danber-Client", "1")
	resp, e := server.Client().Do(login)
	if e != nil {
		t.Fatal(e)
	}
	var session struct {
		CSRF string `json:"csrf_token"`
	}
	e = json.NewDecoder(resp.Body).Decode(&session)
	resp.Body.Close()
	if e != nil || resp.StatusCode != 200 || len(resp.Cookies()) != 1 {
		t.Fatal("login failed")
	}
	b.cookie = resp.Cookies()[0]
	b.csrf = session.CSRF
	var created string
	t.Run("anonymous CSRF origin and rate controls", func(t *testing.T) {
		anon := b
		anon.cookie = nil
		if status, _ := anon.call("GET", "/pacientes/tipos-documento", nil, ""); status != 401 {
			t.Fatalf("anonymous=%d", status)
		}
		bad := b
		bad.csrf = "bad"
		if status, _ := bad.call("POST", "/pacientes", input("10000001"), key(1)); status != 403 {
			t.Fatalf("csrf=%d", status)
		}
		r, _ := http.NewRequest("GET", server.URL+"/pacientes/tipos-documento", nil)
		r.AddCookie(b.cookie)
		r.Header.Set("Origin", "https://malicioso.invalid")
		resp, e := server.Client().Do(r)
		if e != nil {
			t.Fatal(e)
		}
		resp.Body.Close()
		if resp.StatusCode != 403 {
			t.Fatal("cross origin accepted")
		}
		if status, _ := b.call("GET", "/pacientes/tipos-documento", nil, ""); status != 200 {
			t.Fatal("catalog denied")
		}
	})
	t.Run("atomic registration detail and exact lookup", func(t *testing.T) {
		status, raw := b.call("POST", "/pacientes", input("10000002"), key(2))
		if status != 201 {
			t.Fatalf("register status=%d error=%s", status, raw)
		}
		var result app.Result
		if json.Unmarshal(raw, &result) != nil || !domain.UUID(result.ID) || result.Version != 1 {
			t.Fatal("invalid result")
		}
		created = result.ID
		status, raw = b.call("GET", "/pacientes/"+created, nil, "")
		var p domain.Patient
		if status != 200 || json.Unmarshal(raw, &p) != nil || p.Document.Number != "10000002" || p.Guardian != nil {
			t.Fatal("detail incorrect")
		}
		status, raw = b.call("POST", "/pacientes/buscar", domain.Document{Type: "dni", Number: "10000002"}, "")
		if status != 200 || strings.Contains(string(raw), "direccion") || strings.Contains(string(raw), "telefono") || strings.Contains(string(raw), "fecha_nacimiento") {
			t.Fatal("lookup leaks unnecessary fields")
		}
		if status, _ = b.call("POST", "/pacientes/buscar", domain.Document{Type: "DNI", Number: "99999999"}, ""); status != 404 {
			t.Fatal("absent not 404")
		}
		var ok bool
		e := admin.QueryRow(ctx, `SELECT p.creado_por=$1 AND p.actualizado_por=$1 AND c.actor_id=$1 AND c.antes IS NULL AND c.despues->'documento'->>'numero'='10000002' FROM pacientes.paciente p JOIN pacientes.cambio_paciente c ON c.paciente_id=p.id WHERE p.id=$2`, actor, created).Scan(&ok)
		if e != nil || !ok {
			t.Fatalf("audit incorrect %v", e)
		}
	})
	t.Run("concurrent equal key and incompatible payload", func(t *testing.T) {
		var wg sync.WaitGroup
		out := make(chan string, 4)
		for range 4 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				status, raw := b.call("POST", "/pacientes", input("10000003"), key(3))
				if status != 201 {
					out <- fmt.Sprintf("status:%d:%s", status, raw)
					return
				}
				var r app.Result
				_ = json.Unmarshal(raw, &r)
				out <- r.ID
			}()
		}
		wg.Wait()
		close(out)
		id := ""
		for got := range out {
			if !domain.UUID(got) {
				t.Fatal(got)
			}
			if id != "" && id != got {
				t.Fatal("retry changed id")
			}
			id = got
		}
		var n int
		admin.QueryRow(ctx, `SELECT count(*) FROM pacientes.cambio_paciente WHERE paciente_id=$1`, id).Scan(&n)
		if n != 1 {
			t.Fatal("duplicate audit")
		}
		r := input("10000003")
		r.Phone = "222"
		if status, _ := b.call("POST", "/pacientes", r, key(3)); status != 409 {
			t.Fatal("changed payload not conflict")
		}
	})
	t.Run("concurrent different keys same document", func(t *testing.T) {
		statuses := make(chan int, 2)
		for i := 4; i <= 5; i++ {
			go func(n int) { s, _ := b.call("POST", "/pacientes", input("10000004"), key(n)); statuses <- s }(i)
		}
		x, y := <-statuses, <-statuses
		if !(x == 201 && y == 409 || x == 409 && y == 201) {
			t.Fatalf("statuses=%d,%d", x, y)
		}
	})
	t.Run("guardian required reused and never overwritten", func(t *testing.T) {
		r := input("10000006")
		r.BirthDate = "2020-01-01"
		if status, _ := b.call("POST", "/pacientes", r, key(6)); status != 422 {
			t.Fatal("minor without guardian accepted")
		}
		r.Guardian = guardian()
		if status, raw := b.call("POST", "/pacientes", r, key(6)); status != 201 {
			t.Fatalf("minor=%d %s", status, raw)
		}
		r.Document.Number = "10000007"
		if status, _ := b.call("POST", "/pacientes", r, key(7)); status != 201 {
			t.Fatal("guardian reuse failed")
		}
		r.Document.Number = "10000008"
		r.Guardian.Names = "Otro nombre"
		if status, _ := b.call("POST", "/pacientes", r, key(8)); status != 409 {
			t.Fatal("guardian overwrite accepted")
		}
		var n int
		admin.QueryRow(ctx, `SELECT count(*) FROM pacientes.apoderado WHERE numero='90000001' AND nombres='María'`).Scan(&n)
		if n != 1 {
			t.Fatal("guardian duplicated or changed")
		}
		admin.QueryRow(ctx, `SELECT count(*) FROM pacientes.paciente_apoderado`).Scan(&n)
		if n != 2 {
			t.Fatal("links missing")
		}
		r = input("10000009")
		r.BirthDate = "1900-01-01"
		if status, _ := b.call("POST", "/pacientes", r, key(9)); status != 422 {
			t.Fatal("elder without guardian accepted")
		}
	})
	t.Run("catalog active validation and non DNI format", func(t *testing.T) {
		r := input("X123")
		r.Document.Type = "PRUEBA"
		if status, _ := b.call("POST", "/pacientes", r, key(10)); status != 422 {
			t.Fatal("unknown type accepted")
		}
		if _, e := admin.Exec(ctx, `INSERT INTO pacientes.tipo_documento(codigo,nombre) VALUES('PRUEBA','Solo fixture; no tipo oficial')`); e != nil {
			t.Fatal(e)
		}
		if status, raw := b.call("POST", "/pacientes", r, key(10)); status != 201 {
			t.Fatalf("nonDNI=%d %s", status, raw)
		}
		if _, e := admin.Exec(ctx, `UPDATE pacientes.tipo_documento SET activo=false WHERE codigo='PRUEBA'`); e != nil {
			t.Fatal(e)
		}
		r.Document.Number = "Y123"
		if status, _ := b.call("POST", "/pacientes", r, key(11)); status != 422 {
			t.Fatal("inactive type accepted")
		}
		if status, _ := b.call("POST", "/pacientes/buscar", domain.Document{Type: "PRUEBA", Number: "x123"}, ""); status != 200 {
			t.Fatal("existing document became unreadable")
		}
	})
	t.Run("concurrent children share one guardian", func(t *testing.T) {
		statuses := make(chan int, 2)
		for i := 20; i <= 21; i++ {
			go func(n int) {
				r := input(fmt.Sprintf("100000%02d", n))
				r.BirthDate = "2020-01-01"
				r.Guardian = guardian()
				r.Guardian.Document.Number = "90000020"
				status, _ := b.call("POST", "/pacientes", r, key(n))
				statuses <- status
			}(i)
		}
		if a, b := <-statuses, <-statuses; a != 201 || b != 201 {
			t.Fatalf("concurrent guardian statuses=%d,%d", a, b)
		}
		var count int
		if e := admin.QueryRow(ctx, `SELECT count(*) FROM pacientes.apoderado WHERE numero='90000020'`).Scan(&count); e != nil || count != 1 {
			t.Fatal("concurrent guardian duplicated")
		}
	})
	t.Run("audit failure rolls back entire aggregate and key", func(t *testing.T) {
		var before int
		if e := admin.QueryRow(ctx, `SELECT count(*) FROM pacientes.paciente`).Scan(&before); e != nil {
			t.Fatal(e)
		}
		if _, e := admin.Exec(ctx, `ALTER TABLE pacientes.cambio_paciente ADD CONSTRAINT rechazo_fixture CHECK(operacion <> 'paciente_registrado') NOT VALID`); e != nil {
			t.Fatal(e)
		}
		r := input("10000012")
		r.BirthDate = "2020-01-01"
		r.Guardian = guardian()
		r.Guardian.Document.Number = "90000012"
		status, raw := b.call("POST", "/pacientes", r, key(12))
		if status != 503 || strings.Contains(string(raw), "rechazo_fixture") {
			t.Fatal("audit failure not sanitized")
		}
		var n int
		admin.QueryRow(ctx, `SELECT count(*) FROM pacientes.documento_paciente WHERE numero='10000012'`).Scan(&n)
		if n != 0 {
			t.Fatal("orphan document")
		}
		admin.QueryRow(ctx, `SELECT count(*) FROM pacientes.apoderado WHERE numero='90000012'`).Scan(&n)
		if n != 0 {
			t.Fatal("orphan guardian")
		}
		admin.QueryRow(ctx, `SELECT count(*) FROM pacientes.registro_idempotente WHERE clave=$1`, key(12)).Scan(&n)
		if n != 0 {
			t.Fatal("failed key consumed")
		}
		if e := admin.QueryRow(ctx, `SELECT count(*) FROM pacientes.paciente`).Scan(&n); e != nil || n != before {
			t.Fatal("partial patient survived")
		}
		if _, e := admin.Exec(ctx, `ALTER TABLE pacientes.cambio_paciente DROP CONSTRAINT rechazo_fixture`); e != nil {
			t.Fatal(e)
		}
		if status, _ := b.call("POST", "/pacientes", r, key(12)); status != 201 {
			t.Fatal("retry after rollback failed")
		}
	})
	t.Run("birthday replay preserves original result", func(t *testing.T) {
		zone, _ := time.LoadLocation("America/Lima")
		now := time.Date(2026, 10, 18, 12, 0, 0, 0, zone)
		s := app.Service{Store: sqlstore.Store{DB: pool}, Location: zone, Now: func() time.Time { return now }}
		a := app.Actor{ID: actor, Permissions: []string{"pacientes.crear"}}
		r := input("10000013")
		r.BirthDate = "1960-10-19"
		first, e := s.Register(ctx, a, key(13), r)
		if e != nil {
			t.Fatal(e)
		}
		now = now.AddDate(0, 0, 1)
		replay, e := s.Register(ctx, a, key(13), r)
		if e != nil || replay != first {
			t.Fatal("birthday broke idempotency")
		}
		r.Document.Number = "10000014"
		_, e = s.Register(ctx, a, key(14), r)
		var v *domain.Validation
		if !errors.As(e, &v) || v.Field != "apoderado" {
			t.Fatal("66 boundary not enforced")
		}
	})
	t.Run("permissions reloaded and enforced before resource lookup", func(t *testing.T) {
		if _, e := admin.Exec(ctx, `DELETE FROM identidad.usuario_rol WHERE usuario_id=$1`, actor); e != nil {
			t.Fatal(e)
		}
		if status, _ := b.call("GET", "/pacientes/"+created, nil, ""); status != 403 {
			t.Fatal("roles cached or resource exposed")
		}
		if status, _ := b.call("GET", "/pacientes/"+key(999), nil, ""); status != 403 {
			t.Fatal("absent resource exposed before permission")
		}
		if status, _ := b.call("POST", "/pacientes", input("10000015"), key(15)); status != 403 {
			t.Fatal("create permission bypass")
		}
		if _, e := admin.Exec(ctx, `INSERT INTO identidad.usuario_rol(usuario_id,rol_id) SELECT $1,id FROM identidad.rol WHERE codigo='recepcion'`, actor); e != nil {
			t.Fatal(e)
		}
		if status, _ := b.call("GET", "/pacientes/"+created, nil, ""); status != 200 {
			t.Fatal("reception access failed")
		}
		if status, _ := b.call("POST", "/auth/logout-all", nil, ""); status != 204 {
			t.Fatal("logout failed")
		}
		if status, _ := b.call("GET", "/pacientes/"+created, nil, ""); status != 401 {
			t.Fatal("revoked cookie accepted")
		}
	})
	t.Run("runtime cannot erase audit or idempotency", func(t *testing.T) {
		for _, table := range []string{"pacientes.registro_idempotente", "pacientes.cambio_paciente"} {
			conn, e := pool.BeginTx(ctx, pgx.TxOptions{})
			if e != nil {
				t.Fatal(e)
			}
			_, e = conn.Exec(ctx, `DELETE FROM `+table)
			_ = conn.Rollback(context.Background())
			var sqlError *pgconn.PgError
			if !errors.As(e, &sqlError) || sqlError.Code != "42501" {
				t.Fatal("runtime delete was not denied by privileges")
			}
		}
	})
}

func TestUpgradeToPatientRegistration(t *testing.T) {
	ctx, admin, pool := fixture(t, 10)
	id, e := identidad.NewInitialAccount(pool).Create(ctx, "Ficticio", []byte("frase ficticia para probar"), "operador_pruebas")
	if e != nil {
		t.Fatal(e)
	}
	m, e := pacientes.New(pool)
	if e != nil {
		t.Fatal(e)
	}
	if e = m.CheckSchema(ctx); e == nil {
		t.Fatal("missing 011 accepted")
	}
	_, source, _, _ := runtime.Caller(0)
	raw, e := os.ReadFile(filepath.Join(filepath.Dir(source), "../../../db/migrations/011_registro_paciente_idempotente.sql"))
	if e != nil {
		t.Fatal(e)
	}
	applyStatements(t, ctx, admin, string(raw))
	if e = m.CheckSchema(ctx); e != nil {
		t.Fatal(e)
	}
	var preserved bool
	if e = admin.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM identidad.usuario WHERE id=$1)`, id).Scan(&preserved); e != nil || !preserved {
		t.Fatal("account lost")
	}
}
