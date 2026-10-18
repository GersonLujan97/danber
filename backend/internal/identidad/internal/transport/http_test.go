package transport

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/GersonLujan97/danber/backend/internal/identidad/internal/app"
)

type ops struct {
	calls int
	err   error
	last  []byte
}

func (o *ops) Login(_ context.Context, _ string, password []byte, _ string) (app.Session, error) {
	o.calls++
	o.last = append([]byte(nil), password...)
	return app.Session{Token: strings.Repeat("A", 43), Expires: time.Now().Add(time.Hour), Principal: app.Principal{ID: "synthetic", Roles: []string{}, Permissions: []string{}}}, o.err
}
func (o *ops) Current(context.Context, string) (app.Session, error) {
	o.calls++
	return app.Session{}, o.err
}
func (o *ops) Logout(context.Context, string, bool) error { o.calls++; return o.err }
func fixture(t *testing.T, o *ops, allow bool) http.Handler {
	t.Helper()
	h, err := NewHandler(o, HTTPOptions{Origin: "http://127.0.0.1:8090", AllowInsecureLocal: allow})
	if err != nil {
		t.Fatal(err)
	}
	return h
}
func request(method, path, body string) *http.Request {
	r := httptest.NewRequest(method, "http://127.0.0.1:8090"+path, strings.NewReader(body))
	r.RemoteAddr = "127.0.0.1:3456"
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-Danber-Client", "1")
	return r
}

const payload = `{"usuario":"Isai","contrasena":"frase ficticia para pruebas"}`

func TestUnsafeInputsAndOriginsNeverReachApplication(t *testing.T) {
	for _, tt := range []struct {
		name   string
		edit   func(*http.Request)
		body   string
		status int
	}{
		{"cross origin", func(r *http.Request) { r.Header.Set("Origin", "https://malicioso.invalid") }, payload, 403},
		{"host rebinding", func(r *http.Request) { r.Host = "malicioso.invalid" }, payload, 403},
		{"cross site", func(r *http.Request) { r.Header.Set("Sec-Fetch-Site", "cross-site") }, payload, 403},
		{"same site sibling", func(r *http.Request) { r.Header.Set("Sec-Fetch-Site", "same-site") }, payload, 403},
		{"no custom header", func(r *http.Request) { r.Header.Del("X-Danber-Client") }, payload, 403},
		{"form", func(r *http.Request) { r.Header.Set("Content-Type", "application/x-www-form-urlencoded") }, "usuario=Isai", 415},
		{"unknown field", func(*http.Request) {}, `{"usuario":"Isai","contrasena":"x","rol":"maestro"}`, 400},
		{"duplicate", func(*http.Request) {}, `{"usuario":"A","usuario":"B","contrasena":"x"}`, 400},
		{"trailing", func(*http.Request) {}, payload + ` {}`, 400},
		{"oversized", func(*http.Request) {}, strings.Repeat("x", 4097), 413},
		{"oversized chunked", func(r *http.Request) { r.ContentLength = -1 }, strings.Repeat("x", 4097), 413},
	} {
		t.Run(tt.name, func(t *testing.T) {
			o := &ops{}
			h := fixture(t, o, true)
			r := request("POST", "/auth/login", tt.body)
			tt.edit(r)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != tt.status || o.calls != 0 {
				t.Fatalf("status=%d calls=%d", w.Code, o.calls)
			}
			if strings.Contains(w.Body.String(), "frase ficticia") {
				t.Fatal("secret reflected")
			}
		})
	}
}
func TestHTTPRequiresExplicitLocalExceptionAndRealLoopbackPeer(t *testing.T) {
	o := &ops{}
	w := httptest.NewRecorder()
	fixture(t, o, false).ServeHTTP(w, request("POST", "/auth/login", payload))
	if w.Code != 426 || o.calls != 0 {
		t.Fatal("HTTP silently accepted")
	}
	r := request("POST", "/auth/login", payload)
	r.RemoteAddr = "203.0.113.1:3456"
	r.Header.Set("X-Forwarded-For", "127.0.0.1")
	w = httptest.NewRecorder()
	fixture(t, o, true).ServeHTTP(w, r)
	if w.Code != 403 || o.calls != 0 {
		t.Fatal("forwarding header bypassed peer check")
	}
}
func TestHTTPSCookiesAndJSONNeverExposeBearer(t *testing.T) {
	o := &ops{}
	h, err := NewHandler(o, HTTPOptions{Origin: "https://127.0.0.1:8090"})
	if err != nil {
		t.Fatal(err)
	}
	r := request("POST", "/auth/login", payload)
	r.TLS = &tls.ConnectionState{}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].Secure || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteStrictMode || cookies[0].Domain != "" || cookies[0].Path != "/" {
		t.Fatal("cookie safeguards missing")
	}
	if strings.Contains(w.Body.String(), cookies[0].Value) || strings.Contains(w.Body.String(), "frase ficticia") {
		t.Fatal("bearer/password exposed in JSON")
	}
	var body struct {
		CSRF string `json:"csrf_token"`
	}
	if json.Unmarshal(w.Body.Bytes(), &body) != nil || body.CSRF == "" {
		t.Fatal("missing CSRF token")
	}
}
func TestLogoutRequiresCSRFBoundToCookie(t *testing.T) {
	for _, header := range []string{"", "invalid", csrf(strings.Repeat("B", 43)), csrf(strings.Repeat("A", 43))} {
		o := &ops{}
		h := fixture(t, o, true)
		r := request("POST", "/auth/logout", "")
		r.AddCookie(&http.Cookie{Name: cookieName, Value: strings.Repeat("A", 43)})
		r.Header.Set("X-Danber-CSRF", header)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if header == csrf(strings.Repeat("A", 43)) {
			if w.Code != 204 || o.calls != 1 || w.Result().Cookies()[0].MaxAge != -1 {
				t.Fatal("logout failed")
			}
		} else if w.Code != 403 || o.calls != 0 {
			t.Fatal("CSRF bypass")
		}
	}
}
func TestRateLimitUsesPeerAndCanonicalUsernameAndResets(t *testing.T) {
	o := &ops{err: app.ErrCredentials}
	h := fixture(t, o, true)
	for i := range 11 {
		r := request("POST", "/auth/login", strings.ReplaceAll(payload, "Isai", strings.Repeat("x", i+1)))
		r.Header.Set("X-Forwarded-For", string(rune('a'+i)))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if i < 10 && w.Code != 401 {
			t.Fatal(w.Code)
		}
		if i == 10 && (w.Code != 429 || w.Header().Get("Retry-After") == "") {
			t.Fatal("spoofed forwarding header bypassed rate limit")
		}
	}
	if o.calls != 10 {
		t.Fatal("limit did not bound authentication work")
	}
	l := newLimiter(2, 3, 2)
	now := time.Now()
	l.now = func() time.Time { return now }
	for range 2 {
		if ok, _ := l.allow("first"); !ok {
			t.Fatal("unexpected limit")
		}
	}
	if ok, _ := l.allow("first"); ok {
		t.Fatal("per-key limit bypassed")
	}
	if ok, _ := l.allow("second"); !ok {
		t.Fatal("other key blocked")
	}
	if ok, _ := l.allow("third"); ok {
		t.Fatal("global/capacity limit bypassed")
	}
	now = now.Add(time.Minute)
	if ok, _ := l.allow("third"); !ok {
		t.Fatal("expired entries did not recover")
	}
}
func TestLimiterConcurrentBudgetAndBoundedMemory(t *testing.T) {
	l := newLimiter(5, 100, 2)
	var wg sync.WaitGroup
	var mu sync.Mutex
	success := 0
	for range 30 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if ok, _ := l.allow("one"); ok {
				mu.Lock()
				success++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if success != 5 {
		t.Fatal("concurrent budget exceeded")
	}
	if ok, _ := l.allow("two"); !ok {
		t.Fatal("capacity too small")
	}
	if ok, _ := l.allow("three"); ok || len(l.entries) > 2 {
		t.Fatal("unbounded memory/eviction")
	}
}
func TestErrorsDoNotExposeInfrastructureOrSecrets(t *testing.T) {
	o := &ops{err: errors.New("postgresql://private:SECRET@host/db")}
	w := httptest.NewRecorder()
	fixture(t, o, true).ServeHTTP(w, request("POST", "/auth/login", payload))
	if w.Code != 503 || bytes.Contains(w.Body.Bytes(), []byte("SECRET")) {
		t.Fatal("internal detail leaked")
	}
}

func TestAccountLimitNormalizesCaseAndSpacesAcrossPeers(t *testing.T) {
	o := &ops{err: app.ErrCredentials}
	h := fixture(t, o, true)
	for i, name := range []string{"Isai", " ISAI ", "isai", "Isai", " isai ", "ISAI"} {
		r := request("POST", "/auth/login", strings.Replace(payload, "Isai", name, 1))
		r.RemoteAddr = "127.0.0." + string(rune('1'+i)) + ":1234"
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if i < 5 && w.Code != 401 {
			t.Fatal(w.Code)
		}
		if i == 5 && w.Code != 429 {
			t.Fatal("account limit bypassed across source peers")
		}
	}
	if o.calls != 5 {
		t.Fatal("too many password verifications")
	}
}

func TestAmbiguousSessionCookiesAreRejected(t *testing.T) {
	o := &ops{}
	h := fixture(t, o, true)
	r := request("GET", "/auth/session", "")
	r.AddCookie(&http.Cookie{Name: cookieName, Value: strings.Repeat("A", 43)})
	r.AddCookie(&http.Cookie{Name: cookieName, Value: strings.Repeat("B", 43)})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 401 || o.calls != 0 {
		t.Fatal("ambiguous session accepted")
	}
}
