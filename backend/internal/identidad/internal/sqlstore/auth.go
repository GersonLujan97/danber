package sqlstore

import (
	"context"
	"errors"
	"time"

	"github.com/GersonLujan97/danber/backend/internal/identidad/internal/app"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s Store) CheckAuthSchema(ctx context.Context) error {
	return s.authTx(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `SELECT ultimo_uso_en,usuario_version FROM identidad.sesion LIMIT 0`)
		if err != nil {
			return err
		}
		rows.Close()
		return rows.Err()
	})
}

// No SQL errors or connection strings leave this adapter.
func (s Store) authTx(ctx context.Context, work func(pgx.Tx) error) error {
	for attempt := 0; attempt < 5; attempt++ {
		err := s.authOnce(ctx, work)
		if err == nil {
			return nil
		}
		var sqlErr *pgconn.PgError
		if errors.As(err, &sqlErr) && sqlErr.Code == "40001" {
			timer := time.NewTimer(time.Duration(attempt+1) * 20 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return app.ErrUnavailable
			case <-timer.C:
			}
			continue
		}
		if errors.Is(err, app.ErrCredentials) {
			return app.ErrCredentials
		}
		if errors.Is(err, app.ErrSession) {
			return app.ErrSession
		}
		return app.ErrUnavailable
	}
	return app.ErrUnavailable
}
func (s Store) authOnce(ctx context.Context, work func(pgx.Tx) error) error {
	tx, err := s.DB.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return err
	}
	defer func() {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(c)
	}()
	if err = work(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s Store) FindCredential(ctx context.Context, name string) (app.Credential, error) {
	var result app.Credential
	err := s.authTx(ctx, func(tx pgx.Tx) error {
		result = app.Credential{}
		err := tx.QueryRow(ctx, `SELECT u.id::STRING,c.hash_clave,u.estado='activo'
 FROM identidad.usuario u JOIN identidad.credencial c ON c.usuario_id=u.id
 WHERE u.nombre_acceso_normalizado=lower(btrim($1))`, name).Scan(&result.ID, &result.Hash, &result.Active)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	})
	return result, err
}

func principal(ctx context.Context, tx pgx.Tx, id string) (app.Principal, error) {
	p := app.Principal{ID: id, Roles: []string{}, Permissions: []string{}}
	if err := tx.QueryRow(ctx, `SELECT nombre_acceso FROM identidad.usuario WHERE id=$1`, id).Scan(&p.Username); err != nil {
		return p, err
	}
	rows, err := tx.Query(ctx, `SELECT r.codigo FROM identidad.usuario_rol ur JOIN identidad.rol r ON r.id=ur.rol_id WHERE ur.usuario_id=$1 ORDER BY r.codigo`, id)
	if err != nil {
		return p, err
	}
	for rows.Next() {
		var code string
		if err = rows.Scan(&code); err != nil {
			rows.Close()
			return p, err
		}
		p.Roles = append(p.Roles, code)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return p, err
	}
	rows.Close()
	rows, err = tx.Query(ctx, `SELECT DISTINCT p.codigo FROM identidad.usuario_rol ur
 JOIN identidad.rol_permiso rp ON rp.rol_id=ur.rol_id JOIN identidad.permiso p ON p.id=rp.permiso_id
 WHERE ur.usuario_id=$1 ORDER BY p.codigo`, id)
	if err != nil {
		return p, err
	}
	defer rows.Close()
	for rows.Next() {
		var code string
		if err = rows.Scan(&code); err != nil {
			return p, err
		}
		p.Permissions = append(p.Permissions, code)
	}
	return p, rows.Err()
}

func (s Store) StartSession(ctx context.Context, c app.Credential, hash, previous []byte, ttl time.Duration) (app.Session, error) {
	var result app.Session
	err := s.authTx(ctx, func(tx pgx.Tx) error {
		var version int64
		err := tx.QueryRow(ctx, `SELECT version FROM identidad.usuario WHERE id=$1 AND estado='activo' FOR UPDATE`, c.ID).Scan(&version)
		if errors.Is(err, pgx.ErrNoRows) {
			return app.ErrCredentials
		}
		if err != nil {
			return err
		}
		var unchanged bool
		if err = tx.QueryRow(ctx, `SELECT hash_clave=$2 FROM identidad.credencial WHERE usuario_id=$1 FOR UPDATE`, c.ID, c.Hash).Scan(&unchanged); err != nil {
			return err
		}
		if !unchanged {
			return app.ErrCredentials
		}
		if len(previous) == 32 {
			// Rotation closes a browser's previous session in the SAME account.
			if err = closeOne(ctx, tx, previous, c.ID); err != nil {
				return err
			}
		}
		var sessionID string
		err = tx.QueryRow(ctx, `INSERT INTO identidad.sesion(usuario_id,token_hash,expira_en,usuario_version)
 VALUES($1,$2,now()+$3::INT8*INTERVAL '1 second',$4) RETURNING id::STRING,expira_en`, c.ID, hash, int64(ttl/time.Second), version).Scan(&sessionID, &result.Expires)
		if err != nil {
			return err
		}
		if err = auditSession(ctx, tx, c.ID, sessionID, "sesion_abierta"); err != nil {
			return err
		}
		result.Principal, err = principal(ctx, tx, c.ID)
		return err
	})
	return result, err
}

func touch(ctx context.Context, tx pgx.Tx, hash []byte, idle time.Duration) (string, string, time.Time, error) {
	var id, user string
	var expires time.Time
	err := tx.QueryRow(ctx, `UPDATE identidad.sesion AS s SET ultimo_uso_en=now()
 FROM identidad.usuario u,identidad.credencial c
 WHERE s.token_hash=$1 AND u.id=s.usuario_id AND c.usuario_id=u.id AND u.estado='activo'
 AND s.revocado_en IS NULL AND s.expira_en>now()
 AND s.ultimo_uso_en>now()-$2::INT8*INTERVAL '1 second'
 AND s.usuario_version=u.version AND s.creado_en>=c.cambiado_en
 RETURNING s.id::STRING,s.usuario_id::STRING,s.expira_en`, hash, int64(idle/time.Second)).Scan(&id, &user, &expires)
	if errors.Is(err, pgx.ErrNoRows) {
		err = app.ErrSession
	}
	return id, user, expires, err
}

func (s Store) TouchSession(ctx context.Context, hash []byte, idle time.Duration) (app.Session, error) {
	var result app.Session
	err := s.authTx(ctx, func(tx pgx.Tx) error {
		_, user, expires, err := touch(ctx, tx, hash, idle)
		if err != nil {
			return err
		}
		result.Expires = expires
		result.Principal, err = principal(ctx, tx, user)
		return err
	})
	return result, err
}

func auditSession(ctx context.Context, tx pgx.Tx, user, id, operation string) error {
	_, err := tx.Exec(ctx, `INSERT INTO identidad.cambio_acceso(usuario_id,actor_id,operacion,detalle)
 VALUES($1,$1,$2,jsonb_build_object('origen','http','sesion_id',$3::STRING))`, user, operation, id)
	return err
}

func closeOne(ctx context.Context, tx pgx.Tx, hash []byte, owner string) error {
	var id, user string
	err := tx.QueryRow(ctx, `UPDATE identidad.sesion SET revocado_en=now()
 WHERE token_hash=$1 AND revocado_en IS NULL AND ($2::STRING='' OR usuario_id::STRING=$2)
 RETURNING id::STRING,usuario_id::STRING`, hash, owner).Scan(&id, &user)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	return auditSession(ctx, tx, user, id, "sesion_cerrada")
}

func (s Store) Revoke(ctx context.Context, hash []byte, idle time.Duration, all bool) error {
	return s.authTx(ctx, func(tx pgx.Tx) error {
		if !all {
			return closeOne(ctx, tx, hash, "")
		}
		id, user, _, err := touch(ctx, tx, hash, idle)
		if err != nil {
			return err
		}
		// Version and revocation serialize against logins on the same account.
		if _, err = tx.Exec(ctx, `UPDATE identidad.usuario SET version=version+1,actualizado_en=now() WHERE id=$1`, user); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, `UPDATE identidad.sesion SET revocado_en=now() WHERE usuario_id=$1 AND revocado_en IS NULL`, user); err != nil {
			return err
		}
		return auditSession(ctx, tx, user, id, "sesiones_revocadas")
	})
}
