package transport

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GersonLujan97/danber/backend/internal/pacientes/internal/domain"
)

func TestStrictJSON(t *testing.T) {
	for _, c := range []struct {
		body, kind string
		status     int
	}{
		{`{"documento":{"tipo":"DNI","numero":"00123456","numero":"00999999"}}`, "application/json", 400},
		{`{"nombres":"Ana","rol":"maestro"}`, "application/json", 400},
		{`{"Nombres":"Ana"}`, "application/json", 400},
		{`{"nombres":null}`, "application/json", 400},
		{`{"nombres":123}`, "application/json", 400},
		{`{"apoderado":{"documento":{"Numero":"00123456"}}}`, "application/json", 400},
		{`{} {}`, "application/json", 400},
		{`[]`, "application/json", 400},
		{"{\"nombres\":\"\xff\"}", "application/json", 400},
		{strings.Repeat("x", (16<<10)+1), "application/json", 413},
		{`{}`, "text/plain", 415},
	} {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/pacientes", strings.NewReader(c.body))
		r.Header.Set("Content-Type", c.kind)
		r.ContentLength = -1
		var data domain.Registration
		if decode(w, r, &data) || w.Code != c.status {
			t.Fatalf("status %d expected %d", w.Code, c.status)
		}
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/pacientes", strings.NewReader(`{"nombres":"Lucía","apoderado":null}`))
	r.Header.Set("Content-Type", "application/json; charset=utf-8")
	var data domain.Registration
	if !decode(w, r, &data) || data.Names != "Lucía" {
		t.Fatal("valid JSON rejected")
	}
}
