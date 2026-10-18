package app

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"time"

	"github.com/GersonLujan97/danber/backend/internal/pacientes/internal/domain"
)

var (
	ErrForbidden        = errors.New("acceso_denegado")
	ErrDuplicate        = errors.New("documento_duplicado")
	ErrGuardianConflict = errors.New("apoderado_datos_difieren")
	ErrKeyConflict      = errors.New("clave_incompatible")
	ErrNotFound         = errors.New("paciente_no_encontrado")
	ErrUnavailable      = errors.New("servicio_no_disponible")
)

type Actor struct {
	ID          string
	Permissions []string
}

func (a Actor) Allows(permission string) bool {
	if !domain.UUID(a.ID) {
		return false
	}
	for _, p := range a.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

type Result struct {
	ID      string `json:"id"`
	Version int64  `json:"version"`
}
type Store interface {
	Register(context.Context, string, string, []byte, domain.Registration, string) (Result, error)
	Find(context.Context, domain.Document) (domain.Summary, error)
	Get(context.Context, string) (domain.Patient, error)
	Types(context.Context) ([]domain.DocumentType, error)
}
type Service struct {
	Store    Store
	Now      func() time.Time
	Location *time.Location
}

func (s Service) Register(ctx context.Context, a Actor, key string, r domain.Registration) (Result, error) {
	if !a.Allows("pacientes.crear") {
		return Result{}, ErrForbidden
	}
	if !domain.UUID(key) {
		return Result{}, domain.Invalid("Idempotency-Key", "uuid_requerido")
	}
	// Copy optional input so normalization never mutates caller-owned data.
	if r.Guardian != nil {
		g := *r.Guardian
		r.Guardian = &g
	}
	today := s.Now().In(s.Location)
	if err := r.Normalize(); err != nil {
		return Result{}, err
	}
	raw, err := json.Marshal(r)
	if err != nil {
		return Result{}, ErrUnavailable
	}
	fingerprint := sha256.Sum256(raw)
	return s.Store.Register(ctx, a.ID, key, fingerprint[:], r, today.Format("2006-01-02"))
}
func (s Service) Find(ctx context.Context, a Actor, d domain.Document) (domain.Summary, error) {
	if !a.Allows("pacientes.leer") {
		return domain.Summary{}, ErrForbidden
	}
	if err := d.Normalize("documento"); err != nil {
		return domain.Summary{}, err
	}
	return s.Store.Find(ctx, d)
}
func (s Service) Get(ctx context.Context, a Actor, id string) (domain.Patient, error) {
	if !a.Allows("pacientes.leer") {
		return domain.Patient{}, ErrForbidden
	}
	if !domain.UUID(id) {
		return domain.Patient{}, domain.Invalid("id", "uuid_requerido")
	}
	return s.Store.Get(ctx, id)
}
func (s Service) Types(ctx context.Context, a Actor) ([]domain.DocumentType, error) {
	if !a.Allows("pacientes.leer") && !a.Allows("pacientes.crear") {
		return nil, ErrForbidden
	}
	return s.Store.Types(ctx)
}
