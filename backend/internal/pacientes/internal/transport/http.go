package transport

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/GersonLujan97/danber/backend/internal/identidad"
	"github.com/GersonLujan97/danber/backend/internal/pacientes/internal/app"
	"github.com/GersonLujan97/danber/backend/internal/pacientes/internal/domain"
)

func actor(r *http.Request) app.Actor {
	p, ok := identidad.PrincipalFromContext(r.Context())
	if !ok {
		return app.Actor{}
	}
	return app.Actor{ID: p.ID, Permissions: p.Permissions}
}
func response(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, status int, code string) {
	response(w, status, map[string]string{"error": code})
}
func serviceError(w http.ResponseWriter, e error) {
	var invalid *domain.Validation
	switch {
	case errors.As(e, &invalid):
		response(w, 422, struct {
			Error string `json:"error"`
			Field string `json:"campo"`
			Code  string `json:"codigo"`
		}{"datos_invalidos", invalid.Field, invalid.Code})
	case errors.Is(e, app.ErrForbidden):
		fail(w, 403, "acceso_denegado")
	case errors.Is(e, app.ErrNotFound):
		fail(w, 404, "paciente_no_encontrado")
	case errors.Is(e, app.ErrDuplicate):
		fail(w, 409, "documento_duplicado")
	case errors.Is(e, app.ErrGuardianConflict):
		fail(w, 409, "apoderado_datos_difieren")
	case errors.Is(e, app.ErrKeyConflict):
		fail(w, 409, "clave_incompatible")
	default:
		w.Header().Set("Retry-After", "1")
		fail(w, 503, "servicio_no_disponible")
	}
}

// Scan object keys before decoding: encoding/json otherwise accepts duplicates
// and case-insensitive struct keys. No arrays, numbers or booleans are inputs here.
func scan(d *json.Decoder, depth int) error {
	if depth > 4 {
		return errors.New("json")
	}
	t, e := d.Token()
	if e != nil {
		return e
	}
	if t == nil {
		return nil
	}
	if _, ok := t.(string); ok {
		return nil
	}
	if t != json.Delim('{') {
		return errors.New("json")
	}
	seen := map[string]bool{}
	for d.More() {
		k, e := d.Token()
		if e != nil {
			return e
		}
		key, ok := k.(string)
		if !ok || seen[key] {
			return errors.New("json")
		}
		seen[key] = true
		if e = scan(d, depth+1); e != nil {
			return e
		}
	}
	_, e = d.Token()
	return e
}
func exact(raw json.RawMessage, t reflect.Type) error {
	if t.Kind() == reflect.Pointer {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return nil
		}
		return exact(raw, t.Elem())
	}
	if t.Kind() == reflect.String {
		var s string
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return errors.New("json")
		}
		return json.Unmarshal(raw, &s)
	}
	if t.Kind() != reflect.Struct || !bytes.HasPrefix(bytes.TrimSpace(raw), []byte("{")) {
		return errors.New("json")
	}
	var fields map[string]json.RawMessage
	if e := json.Unmarshal(raw, &fields); e != nil {
		return e
	}
	for name, value := range fields {
		found := false
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if strings.Split(f.Tag.Get("json"), ",")[0] == name {
				found = true
				if e := exact(value, f.Type); e != nil {
					return e
				}
				break
			}
		}
		if !found {
			return errors.New("json")
		}
	}
	return nil
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	kind, _, e := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if e != nil || kind != "application/json" {
		fail(w, 415, "tipo_no_admitido")
		return false
	}
	raw, e := io.ReadAll(http.MaxBytesReader(w, r.Body, 16<<10))
	if e != nil {
		var big *http.MaxBytesError
		if errors.As(e, &big) {
			fail(w, 413, "cuerpo_demasiado_grande")
		} else {
			fail(w, 400, "solicitud_invalida")
		}
		return false
	}
	defer clear(raw)
	d := json.NewDecoder(bytes.NewReader(raw))
	if !utf8.Valid(raw) || scan(d, 0) != nil {
		fail(w, 400, "solicitud_invalida")
		return false
	}
	if _, e = d.Token(); e != io.EOF {
		fail(w, 400, "solicitud_invalida")
		return false
	}
	if exact(raw, reflect.TypeOf(v).Elem()) != nil || json.Unmarshal(raw, v) != nil {
		fail(w, 400, "solicitud_invalida")
		return false
	}
	return true
}
func Handler(service app.Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /pacientes", func(w http.ResponseWriter, r *http.Request) {
		a := actor(r)
		if !a.Allows("pacientes.crear") {
			serviceError(w, app.ErrForbidden)
			return
		}
		var data domain.Registration
		if !decode(w, r, &data) {
			return
		}
		result, e := service.Register(r.Context(), a, r.Header.Get("Idempotency-Key"), data)
		if e != nil {
			serviceError(w, e)
			return
		}
		// The same minimal 201 response is replayed; no current patient details leak.
		w.Header().Set("Location", "/pacientes/"+result.ID)
		response(w, 201, result)
	})
	mux.HandleFunc("POST /pacientes/buscar", func(w http.ResponseWriter, r *http.Request) {
		a := actor(r)
		if !a.Allows("pacientes.leer") {
			serviceError(w, app.ErrForbidden)
			return
		}
		var data domain.Document
		if !decode(w, r, &data) {
			return
		}
		result, e := service.Find(r.Context(), a, data)
		if e != nil {
			serviceError(w, e)
			return
		}
		response(w, 200, result)
	})
	mux.HandleFunc("GET /pacientes/{id}", func(w http.ResponseWriter, r *http.Request) {
		result, e := service.Get(r.Context(), actor(r), r.PathValue("id"))
		if e != nil {
			serviceError(w, e)
			return
		}
		response(w, 200, result)
	})
	mux.HandleFunc("GET /pacientes/tipos-documento", func(w http.ResponseWriter, r *http.Request) {
		result, e := service.Types(r.Context(), actor(r))
		if e != nil {
			serviceError(w, e)
			return
		}
		response(w, 200, map[string]any{"tipos": result})
	})
	return mux
}
