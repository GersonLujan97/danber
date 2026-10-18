package httpsecurity

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSensitiveResponsesAndErrorsCannotBeCachedOrFramed(t *testing.T) {
	w := httptest.NewRecorder()
	Headers(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(401) })).ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != 401 || w.Header().Get("Cache-Control") != "no-store" || w.Header().Get("X-Frame-Options") != "DENY" || w.Header().Get("X-Content-Type-Options") != "nosniff" || w.Header().Get("Content-Security-Policy") == "" {
		t.Fatal("response safeguards missing")
	}
}
