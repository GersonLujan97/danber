package sqlstore

import (
	"context"
	"errors"
	"testing"

	"github.com/GersonLujan97/danber/backend/internal/identidad/internal/app"
	"github.com/GersonLujan97/danber/backend/internal/identidad/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type failingDB struct {
	calls int
	err   error
}

func (d *failingDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	d.calls++
	return nil, d.err
}

func TestOnlySerializationFailuresAreRetriedAndBounded(t *testing.T) {
	for _, tt := range []struct {
		code  string
		calls int
	}{{"40001", 5}, {"08006", 1}, {"42501", 1}} {
		db := &failingDB{err: &pgconn.PgError{Code: tt.code, Message: "PRIVATE_SQL_DETAIL"}}
		_, err := (Store{DB: db}).CreateInitial(context.Background(), app.InitialAccount{})
		if !errors.Is(err, domain.ErrPersistence) || db.calls != tt.calls {
			t.Fatalf("code=%s attempts=%d", tt.code, db.calls)
		}
	}
}

type uncertainDB struct{ calls int }

func (d *uncertainDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	d.calls++
	return &uncertainTx{}, nil
}

type uncertainTx struct {
	pgx.Tx
	reads int
}
type rowFunc func(...any) error

func (r rowFunc) Scan(dest ...any) error { return r(dest...) }
func (t *uncertainTx) QueryRow(context.Context, string, ...any) pgx.Row {
	t.reads++
	return rowFunc(func(dest ...any) error {
		switch t.reads {
		case 1:
			*(dest[0].(**string)) = nil
		case 2:
			*(dest[0].(*bool)) = false
		case 3:
			*(dest[0].(*int)) = 2
		case 4:
			*(dest[0].(*string)) = "synthetic-id"
		}
		return nil
	})
}
func (t *uncertainTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag("INSERT 0 1"), nil
}
func (t *uncertainTx) Rollback(context.Context) error { return nil }
func (t *uncertainTx) Commit(context.Context) error {
	return errors.New("connection lost: commit outcome unknown")
}

func TestUnknownCommitOutcomeIsNotReplayed(t *testing.T) {
	db := &uncertainDB{}
	_, err := (Store{DB: db}).CreateInitial(context.Background(), app.InitialAccount{})
	if !errors.Is(err, domain.ErrPersistence) || db.calls != 1 {
		t.Fatal("uncertain commit must not be retried automatically")
	}
}
