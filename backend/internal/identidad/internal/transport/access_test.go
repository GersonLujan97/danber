package transport

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProtectedBusinessRoutesRequireSessionCSRFAndShareBudget(t *testing.T) {
	o := &ops{}
	h, e := Protected(o, HTTPOptions{Origin: "http://127.0.0.1:8090", AllowInsecureLocal: true}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := Principal(r.Context()); !ok {
			t.Error("missing authenticated principal")
		}
		w.WriteHeader(204)
	}))
	if e != nil {
		t.Fatal(e)
	}
	call := func(method, path, token, csrfToken string) *httptest.ResponseRecorder {
		r := request(method, path, `{}`)
		if token != "" {
			r.AddCookie(&http.Cookie{Name: cookieName, Value: token})
		}
		r.Header.Set("X-Danber-CSRF", csrfToken)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	if call("GET", "/pacientes/x", "", "").Code != 401 || o.calls != 0 {
		t.Fatal("anonymous accessed service")
	}
	token := strings.Repeat("A", 43)
	if call("POST", "/pacientes", token, "bad").Code != 403 || o.calls != 0 {
		t.Fatal("CSRF bypass")
	}
	if call("POST", "/pacientes", token, csrf(token)).Code != 204 || o.calls != 1 {
		t.Fatal("valid request rejected")
	}
	for range 117 {
		if call("GET", "/pacientes/x", token, "").Code != 204 {
			t.Fatal("budget rejected early")
		}
	}
	w := call("GET", "/pacientes/tipos-documento", token, "")
	if w.Code != 429 || w.Header().Get("Retry-After") == "" {
		t.Fatal("paths did not share bounded budget")
	}
}
