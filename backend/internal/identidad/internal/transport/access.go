package transport

import (
	"context"
	"crypto/subtle"
	"net/http"

	"github.com/GersonLujan97/danber/backend/internal/identidad/internal/app"
)

type principalKey struct{}

func Principal(ctx context.Context) (app.Principal, bool) {
	p, ok := ctx.Value(principalKey{}).(app.Principal)
	return p, ok
}

// Protected shares one bounded budget across all routes of a business module.
// Case services still enforce operation permissions, independently of HTTP.
func Protected(ops Operations, options HTTPOptions, next http.Handler) (http.Handler, error) {
	budget := newLimiter(120, 600, 4096)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !admitted(w, budget, peerIP(r)) {
			return
		}
		token, ok := accessToken(r)
		if !ok || app.TokenHash(token) == nil {
			failure(w, 401, "no_autenticado")
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			if subtle.ConstantTimeCompare([]byte(csrf(token)), []byte(r.Header.Get("X-Danber-CSRF"))) != 1 {
				failure(w, 403, "csrf_invalido")
				return
			}
		}
		session, err := ops.Current(r.Context(), token)
		if err != nil {
			serviceError(w, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), principalKey{}, session.Principal)))
	})
	return guard(options, 16<<10, h)
}
