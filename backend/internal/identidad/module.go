// Package identidad exposes the module's entry point, not its persistence tables.
package identidad

import (
	"context"

	"github.com/GersonLujan97/danber/backend/internal/identidad/internal/app"
	"github.com/GersonLujan97/danber/backend/internal/identidad/internal/domain"
	"github.com/GersonLujan97/danber/backend/internal/identidad/internal/password"
	"github.com/GersonLujan97/danber/backend/internal/identidad/internal/sqlstore"
	"github.com/jackc/pgx/v5"
)

type InitialAccount struct{ service app.Bootstrap }
type Database interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}

func NewInitialAccount(db Database) *InitialAccount {
	return &InitialAccount{service: app.Bootstrap{Store: sqlstore.Store{DB: db}, Hasher: password.Argon2id{}}}
}

func (s *InitialAccount) Create(ctx context.Context, username string, password []byte, operator string) (string, error) {
	return s.service.Create(ctx, username, password, operator)
}

var ErrInitialized = domain.ErrInitialized
