package sqlstore

import (
	"context"
	"errors"
	"time"

	"github.com/GersonLujan97/danber/backend/internal/identidad/internal/app"
	"github.com/GersonLujan97/danber/backend/internal/identidad/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Beginner interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}
type Store struct{ DB Beginner }

// Retry the WHOLE transaction only for a definite serialization failure (40001).
// In particular, a lost connection during COMMIT is not treated as a rollback.
func (s Store) CreateInitial(ctx context.Context, account app.InitialAccount) (string, error) {
	for attempt := 0; attempt < 5; attempt++ {
		id, err := s.create(ctx, account)
		if err == nil {
			return id, nil
		}
		var sqlErr *pgconn.PgError
		if errors.As(err, &sqlErr) && sqlErr.Code == "40001" {
			timer := time.NewTimer(time.Duration(attempt+1) * 20 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return "", domain.ErrPersistence
			case <-timer.C:
			}
			continue
		}
		if errors.As(err, &sqlErr) {
			if sqlErr.Code == "42P01" {
				return "", domain.ErrSchema
			}
			if sqlErr.Code == "23505" {
				return "", domain.ErrUsernameTaken
			}
		}
		for _, known := range []error{domain.ErrInitialized, domain.ErrSchema, domain.ErrRoles, domain.ErrUsernameTaken} {
			if errors.Is(err, known) {
				return "", known
			}
		}
		return "", domain.ErrPersistence
	}
	return "", domain.ErrPersistence
}

func (s Store) create(ctx context.Context, account app.InitialAccount) (string, error) {
	tx, err := s.DB.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return "", err
	}
	defer func() {
		rollbackCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(rollbackCtx)
	}()
	var previous *string
	err = tx.QueryRow(ctx, `UPDATE identidad.aprovisionamiento_inicial SET revision=revision+1
 WHERE clave='maestro_inicial' RETURNING usuario_id::STRING`).Scan(&previous)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrSchema
	}
	if err != nil {
		return "", err
	}
	if previous != nil {
		return "", domain.ErrInitialized
	}
	var exists bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM identidad.usuario_rol ur
 JOIN identidad.rol r ON r.id=ur.rol_id WHERE r.codigo='maestro')`).Scan(&exists)
	if err != nil {
		return "", err
	}
	if exists {
		return "", domain.ErrInitialized
	}
	var roles int
	err = tx.QueryRow(ctx, `SELECT count(*) FROM identidad.rol WHERE codigo IN ('maestro','recepcion')`).Scan(&roles)
	if err != nil {
		return "", err
	}
	if roles != 2 {
		return "", domain.ErrRoles
	}
	var id string
	err = tx.QueryRow(ctx, `INSERT INTO identidad.usuario(nombre_acceso) VALUES($1) RETURNING id::STRING`, account.Username).Scan(&id)
	if err != nil {
		return "", err
	}
	_, err = tx.Exec(ctx, `INSERT INTO identidad.credencial(usuario_id,hash_clave) VALUES($1,$2)`, id, account.PasswordHash)
	if err != nil {
		return "", err
	}
	_, err = tx.Exec(ctx, `INSERT INTO identidad.usuario_rol(usuario_id,rol_id,asignado_por)
 SELECT $1,id,$1 FROM identidad.rol WHERE codigo IN ('maestro','recepcion')`, id)
	if err != nil {
		return "", err
	}
	// actor_id refers to the new account as technical bootstrap actor, NOT a prior login.
	_, err = tx.Exec(ctx, `INSERT INTO identidad.cambio_acceso(usuario_id,actor_id,operacion,detalle)
 VALUES($1,$1,'cuenta_inicial_creada',jsonb_build_object('origen','consola_local',
 'actor_tipo','aprovisionamiento_sin_sesion','operador_sistema',$2::STRING,
 'conexion_sql',current_user,'roles',jsonb_build_array('maestro','recepcion')))`, id, account.Operator)
	if err != nil {
		return "", err
	}
	_, err = tx.Exec(ctx, `UPDATE identidad.aprovisionamiento_inicial
 SET usuario_id=$1,completado_en=now() WHERE clave='maestro_inicial'`, id)
	if err != nil {
		return "", err
	}
	if err = tx.Commit(ctx); err != nil {
		return "", err
	}
	return id, nil
}
