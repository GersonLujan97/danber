package transport

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math"
	"mime"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/GersonLujan97/danber/backend/internal/identidad/internal/app"
)

const cookieName = "danber_session"

type Operations interface {
	Login(context.Context, string, []byte, string) (app.Session, error)
	Current(context.Context, string) (app.Session, error)
	Logout(context.Context, string, bool) error
}
type HTTPOptions struct {
	Origin             string
	AllowInsecureLocal bool
	Logger             *slog.Logger
}
type handler struct {
	ops                           Operations
	options                       HTTPOptions
	loginIP, loginNames, requests *limiter
}

func NewHandler(ops Operations, options HTTPOptions) (http.Handler, error) {
	if options.Logger == nil {
		options.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	h := &handler{ops: ops, options: options, loginIP: newLimiter(10, 60, 4096), loginNames: newLimiter(5, 60, 4096), requests: newLimiter(120, 600, 4096)}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/login", h.login)
	mux.HandleFunc("GET /auth/session", h.current)
	mux.HandleFunc("POST /auth/logout", func(w http.ResponseWriter, r *http.Request) { h.logout(w, r, false) })
	mux.HandleFunc("POST /auth/logout-all", func(w http.ResponseWriter, r *http.Request) { h.logout(w, r, true) })
	return guard(options, 4096, mux)
}

func guard(options HTTPOptions, maxBody int64, next http.Handler) (http.Handler, error) {
	u, err := url.Parse(options.Origin)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("origen HTTP de autenticación inválido")
	}
	ip := net.ParseIP(u.Hostname())
	if u.Hostname() != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return nil, errors.New("autenticación de esta entrega solo admite loopback")
	}
	protection := http.NewCrossOriginProtection()
	protection.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { failure(w, 403, "origen_no_permitido") }))
	protected := protection.Handler(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		if r.Host != u.Host || scheme != u.Scheme {
			failure(w, 403, "origen_no_permitido")
			return
		}
		if r.TLS == nil && !options.AllowInsecureLocal {
			failure(w, 426, "https_requerido")
			return
		}
		if r.TLS == nil {
			ip := peerIP(r)
			parsed, e := netip.ParseAddr(ip)
			if e != nil || !parsed.IsLoopback() {
				failure(w, 403, "origen_no_permitido")
				return
			}
		}
		if origin := r.Header.Get("Origin"); origin != "" && origin != options.Origin {
			failure(w, 403, "origen_no_permitido")
			return
		}
		site := r.Header.Get("Sec-Fetch-Site")
		if site != "" && site != "same-origin" && site != "none" {
			failure(w, 403, "origen_no_permitido")
			return
		}
		if r.Method == http.MethodPost && r.Header.Get("X-Danber-Client") != "1" {
			failure(w, 403, "cliente_no_permitido")
			return
		}
		if r.ContentLength > maxBody {
			failure(w, 413, "cuerpo_demasiado_grande")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		protected.ServeHTTP(w, r.WithContext(ctx))
	}), nil
}

func peerIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "invalid"
	}
	ip, err := netip.ParseAddr(host)
	if err != nil {
		return "invalid"
	}
	return ip.Unmap().String() // Never trust client-supplied forwarding headers.
}
func admitted(w http.ResponseWriter, l *limiter, key string) bool {
	allowed, retry := l.allow(key)
	if allowed {
		return true
	}
	w.Header().Set("Retry-After", strconv.Itoa(max(1, int(math.Ceil(retry.Seconds())))))
	failure(w, 429, "demasiados_intentos")
	return false
}
func accessToken(r *http.Request) (string, bool) {
	var value string
	count := 0
	for _, c := range r.Cookies() {
		if c.Name == cookieName {
			count++
			value = c.Value
		}
	}
	return value, count <= 1
}
func csrf(token string) string {
	m := hmac.New(sha256.New, []byte(token))
	m.Write([]byte("danber-csrf-v1"))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}
func send(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func failure(w http.ResponseWriter, status int, code string) {
	send(w, status, map[string]string{"error": code})
}
func serviceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, app.ErrCredentials):
		failure(w, 401, "credenciales_invalidas")
	case errors.Is(err, app.ErrSession):
		failure(w, 401, "no_autenticado")
	case errors.Is(err, app.ErrBusy):
		w.Header().Set("Retry-After", "1")
		failure(w, 429, "demasiados_intentos")
	default:
		w.Header().Set("Retry-After", "1")
		failure(w, 503, "servicio_no_disponible")
	}
}
func sessionResponse(w http.ResponseWriter, s app.Session, token string) {
	send(w, 200, struct {
		Principal app.Principal `json:"principal"`
		Expires   time.Time     `json:"expira_en"`
		CSRF      string        `json:"csrf_token"`
	}{s.Principal, s.Expires, csrf(token)})
}
func setCookie(w http.ResponseWriter, r *http.Request, token string, expires time.Time, remove bool) {
	c := &http.Cookie{Name: cookieName, Value: token, Path: "/", Expires: expires, HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteStrictMode}
	if remove {
		c.MaxAge = -1
		c.Expires = time.Unix(1, 0)
	}
	http.SetCookie(w, c)
}

// Decode exactly two string fields. Reject duplicate/unknown keys, trailing JSON,
// malformed UTF-8, and oversized bodies without reflecting submitted values.
func credentials(w http.ResponseWriter, r *http.Request) (string, []byte, error) {
	kind, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || kind != "application/json" {
		return "", nil, errors.New("media")
	}
	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 4096))
	if err != nil {
		return "", nil, err
	}
	defer clear(raw)
	if !utf8.Valid(raw) {
		return "", nil, errors.New("json")
	}
	d := json.NewDecoder(strings.NewReader(string(raw)))
	token, err := d.Token()
	if err != nil || token != json.Delim('{') {
		return "", nil, errors.New("json")
	}
	fields := make(map[string]string, 2)
	for d.More() {
		key, e := d.Token()
		if e != nil {
			return "", nil, e
		}
		name, ok := key.(string)
		if !ok || (name != "usuario" && name != "contrasena") {
			return "", nil, errors.New("json")
		}
		if _, exists := fields[name]; exists {
			return "", nil, errors.New("json")
		}
		var value string
		if e = d.Decode(&value); e != nil {
			return "", nil, e
		}
		fields[name] = value
	}
	if _, err = d.Token(); err != nil {
		return "", nil, err
	}
	var extra any
	if err = d.Decode(&extra); err != io.EOF {
		return "", nil, errors.New("json")
	}
	if len(fields) != 2 {
		return "", nil, errors.New("json")
	}
	return fields["usuario"], []byte(fields["contrasena"]), nil
}

func (h *handler) login(w http.ResponseWriter, r *http.Request) {
	if !admitted(w, h.loginIP, peerIP(r)) {
		return
	}
	user, password, err := credentials(w, r)
	defer clear(password)
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			failure(w, 413, "cuerpo_demasiado_grande")
		} else if err.Error() == "media" {
			failure(w, 415, "tipo_no_admitido")
		} else {
			failure(w, 400, "solicitud_invalida")
		}
		return
	}
	nameKey := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(user))))
	if !admitted(w, h.loginNames, hex.EncodeToString(nameKey[:])) {
		return
	}
	previous, ok := accessToken(r)
	if !ok {
		failure(w, 400, "solicitud_invalida")
		return
	}
	s, err := h.ops.Login(r.Context(), user, password, previous)
	if err != nil {
		if errors.Is(err, app.ErrCredentials) {
			h.options.Logger.Info("auth_fallida")
		}
		serviceError(w, err)
		return
	}
	setCookie(w, r, s.Token, s.Expires, false)
	sessionResponse(w, s, s.Token)
}
func (h *handler) current(w http.ResponseWriter, r *http.Request) {
	if !admitted(w, h.requests, peerIP(r)) {
		return
	}
	token, ok := accessToken(r)
	if !ok || app.TokenHash(token) == nil {
		failure(w, 401, "no_autenticado")
		return
	}
	s, err := h.ops.Current(r.Context(), token)
	if err != nil {
		if errors.Is(err, app.ErrSession) {
			setCookie(w, r, "", time.Time{}, true)
		}
		serviceError(w, err)
		return
	}
	sessionResponse(w, s, token)
}
func (h *handler) logout(w http.ResponseWriter, r *http.Request, all bool) {
	if !admitted(w, h.requests, peerIP(r)) {
		return
	}
	token, ok := accessToken(r)
	if !ok || app.TokenHash(token) == nil {
		failure(w, 401, "no_autenticado")
		return
	}
	expected := csrf(token)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(r.Header.Get("X-Danber-CSRF"))) != 1 {
		failure(w, 403, "csrf_invalido")
		return
	}
	if err := h.ops.Logout(r.Context(), token, all); err != nil {
		serviceError(w, err)
		return
	}
	setCookie(w, r, "", time.Time{}, true)
	w.WriteHeader(204)
}
