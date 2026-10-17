package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLivenessDoesNotDependOnDatabase(t *testing.T) {
	called := false
	h := Handler(func(context.Context) error { called = true; return errors.New("postgresql://secret") })
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health/live", nil))
	if recorder.Code != 200 || called {
		t.Fatal("liveness touched unavailable dependency")
	}
}

func TestReadiness(t *testing.T) {
	for _, available := range []bool{true, false} {
		t.Run(map[bool]string{true: "available", false: "unavailable"}[available], func(t *testing.T) {
			h := Handler(func(ctx context.Context) error {
				if _, ok := ctx.Deadline(); !ok {
					t.Error("probe has no deadline")
				}
				if !available {
					return errors.New("password=secret-value")
				}
				return nil
			})
			recorder := httptest.NewRecorder()
			h.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
			expected := 200
			if !available {
				expected = 503
			}
			if recorder.Code != expected {
				t.Fatalf("status %d", recorder.Code)
			}
			if strings.Contains(recorder.Body.String(), "secret-value") {
				t.Fatal("leaked dependency error")
			}
			if recorder.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("probe can be cached")
			}
		})
	}
}

func TestUnsupportedRoutesAndMethods(t *testing.T) {
	h := Handler(func(context.Context) error { return nil })
	for _, tt := range []struct {
		method, path string
		status       int
	}{{"POST", "/health/live", 405}, {"GET", "/api/v1/pacientes", 404}} {
		recorder := httptest.NewRecorder()
		h.ServeHTTP(recorder, httptest.NewRequest(tt.method, tt.path, nil))
		if recorder.Code != tt.status {
			t.Fatalf("%s %s: %d", tt.method, tt.path, recorder.Code)
		}
	}
}
