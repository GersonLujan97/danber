// Package pacientes owns identification, contact, documents and guardians.
package pacientes

import (
	"context"
	"errors"
	"net/http"
	"time"
	_ "time/tzdata"

	"github.com/GersonLujan97/danber/backend/internal/identidad"
	"github.com/GersonLujan97/danber/backend/internal/pacientes/internal/app"
	"github.com/GersonLujan97/danber/backend/internal/pacientes/internal/sqlstore"
	"github.com/GersonLujan97/danber/backend/internal/pacientes/internal/transport"
	"github.com/jackc/pgx/v5"
)

type Module struct {
	store   sqlstore.Store
	service app.Service
}

type Database interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}

func New(db Database) (*Module, error) {
	location, e := time.LoadLocation("America/Lima")
	if e != nil {
		return nil, errors.New("no se pudo cargar zona horaria del consultorio")
	}
	store := sqlstore.Store{DB: db}
	return &Module{store: store, service: app.Service{Store: store, Now: time.Now, Location: location}}, nil
}
func (m *Module) CheckSchema(ctx context.Context) error {
	if e := m.store.CheckSchema(ctx); e != nil {
		return errors.New("no se pudo comprobar el esquema de pacientes; verificar conexión y migración 011")
	}
	return nil
}
func (m *Module) Handler(identity *identidad.Authentication, options identidad.HTTPOptions) (http.Handler, error) {
	return identity.ProtectHTTP(options, transport.Handler(m.service))
}
