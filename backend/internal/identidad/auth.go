package identidad

import (
	"context"
	"errors"
	"net/http"

	"github.com/GersonLujan97/danber/backend/internal/identidad/internal/app"
	"github.com/GersonLujan97/danber/backend/internal/identidad/internal/password"
	"github.com/GersonLujan97/danber/backend/internal/identidad/internal/sqlstore"
	"github.com/GersonLujan97/danber/backend/internal/identidad/internal/transport"
	"github.com/GersonLujan97/danber/backend/internal/platform/httpsecurity"
)

type SessionPolicy = app.Policy
type Principal = app.Principal
type HTTPOptions = transport.HTTPOptions
type Authentication struct {
	service *app.Auth
	store   sqlstore.Store
}

func NewAuthentication(db Database, policy SessionPolicy) (*Authentication, error) {
	hasher := password.Argon2id{}
	dummy, err := hasher.Hash([]byte("verificacion ficticia de cuenta inexistente"))
	if err != nil {
		return nil, errors.New("no se pudo preparar autenticación")
	}
	service, err := app.NewAuth(sqlstore.Store{DB: db}, hasher, dummy, policy)
	if err != nil {
		return nil, err
	}
	return &Authentication{service: service, store: sqlstore.Store{DB: db}}, nil
}

func (a *Authentication) CheckSchema(ctx context.Context) error {
	if err := a.store.CheckAuthSchema(ctx); err != nil {
		return errors.New("no se pudo comprobar el esquema de autenticación; verificar conexión y migración 010")
	}
	return nil
}
func (a *Authentication) Handler(options HTTPOptions) (http.Handler, error) {
	h, err := transport.NewHandler(a.service, options)
	if err != nil {
		return nil, err
	}
	return httpsecurity.Headers(h), nil
}

// Authenticate is a narrow contract. Each business operation must additionally
// authorize its permission AND resource; a role or a hidden button is insufficient.
func (a *Authentication) Authenticate(ctx context.Context, token string) (Principal, error) {
	s, err := a.service.Current(ctx, token)
	return s.Principal, err
}
