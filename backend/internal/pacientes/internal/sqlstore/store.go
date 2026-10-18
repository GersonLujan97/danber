package sqlstore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/GersonLujan97/danber/backend/internal/pacientes/internal/app"
	"github.com/GersonLujan97/danber/backend/internal/pacientes/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Database interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}
type Store struct{ DB Database }

func (s Store) transaction(ctx context.Context, work func(pgx.Tx) error) error {
	for attempt := 0; attempt < 5; attempt++ {
		err := s.once(ctx, work)
		if err == nil {
			return nil
		}
		var p *pgconn.PgError
		if errors.As(err, &p) && p.Code == "40001" {
			timer := time.NewTimer(time.Duration(attempt+1) * 20 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return app.ErrUnavailable
			case <-timer.C:
			}
			continue
		}
		if errors.As(err, &p) && p.Code == "23505" && p.TableName == "documento_paciente" {
			return app.ErrDuplicate
		}
		for _, expected := range []error{app.ErrDuplicate, app.ErrGuardianConflict, app.ErrKeyConflict, app.ErrNotFound} {
			if errors.Is(err, expected) {
				return expected
			}
		}
		var invalid *domain.Validation
		if errors.As(err, &invalid) {
			return invalid
		}
		return app.ErrUnavailable
	}
	return app.ErrUnavailable
}
func (s Store) once(ctx context.Context, work func(pgx.Tx) error) error {
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
	// An unknown COMMIT outcome is not retried internally. Caller repeats its same key.
	return tx.Commit(ctx)
}
func (s Store) CheckSchema(ctx context.Context) error {
	return s.transaction(ctx, func(tx pgx.Tx) error {
		rows, e := tx.Query(ctx, `SELECT actor_id,clave,huella,paciente_id FROM pacientes.registro_idempotente LIMIT 0`)
		if e != nil {
			return e
		}
		rows.Close()
		return rows.Err()
	})
}
func activeType(ctx context.Context, tx pgx.Tx, code, field string) error {
	var ok bool
	if e := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pacientes.tipo_documento WHERE codigo=$1 AND activo)`, code).Scan(&ok); e != nil {
		return e
	}
	if !ok {
		return domain.Invalid(field, "tipo_no_habilitado")
	}
	return nil
}
func (s Store) Register(ctx context.Context, actor, key string, hash []byte, r domain.Registration, today string) (app.Result, error) {
	var result app.Result
	err := s.transaction(ctx, func(tx pgx.Tx) error {
		result = app.Result{}
		// Claim and complete in ONE transaction. The primary key serializes equal retries.
		tag, e := tx.Exec(ctx, `INSERT INTO pacientes.registro_idempotente(actor_id,clave,huella) VALUES($1,$2,$3) ON CONFLICT(actor_id,clave) DO NOTHING`, actor, key, hash)
		if e != nil {
			return e
		}
		var priorHash []byte
		var priorID *string
		if e = tx.QueryRow(ctx, `SELECT huella,paciente_id::STRING FROM pacientes.registro_idempotente WHERE actor_id=$1 AND clave=$2 FOR UPDATE`, actor, key).Scan(&priorHash, &priorID); e != nil {
			return e
		}
		if !bytes.Equal(priorHash, hash) {
			return app.ErrKeyConflict
		}
		if tag.RowsAffected() == 0 {
			if priorID == nil {
				return app.ErrUnavailable
			}
			result = app.Result{ID: *priorID, Version: 1}
			return nil
		}
		day, e := time.Parse("2006-01-02", today)
		if e != nil {
			return e
		}
		if e = r.ValidateDate(day); e != nil {
			return e
		}
		if e = activeType(ctx, tx, r.Document.Type, "documento.tipo"); e != nil {
			return e
		}
		if r.Guardian != nil {
			if e = activeType(ctx, tx, r.Guardian.Document.Type, "apoderado.documento.tipo"); e != nil {
				return e
			}
		}
		var exists bool
		if e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pacientes.documento_paciente WHERE tipo=$1 AND numero_normalizado=$2)`, r.Document.Type, r.Document.Number).Scan(&exists); e != nil {
			return e
		}
		if exists {
			return app.ErrDuplicate
		}
		if e = tx.QueryRow(ctx, `INSERT INTO pacientes.paciente(nombres,apellidos,fecha_nacimiento,estado_civil,ocupacion,direccion,telefono,correo,creado_por,actualizado_por)
 VALUES($1,$2,$3::DATE,NULLIF($4,''),NULLIF($5,''),$6,$7,NULLIF($8,''),$9,$9) RETURNING id::STRING,version`, r.Names, r.Surnames, r.BirthDate, r.CivilStatus, r.Occupation, r.Address, r.Phone, r.Email, actor).Scan(&result.ID, &result.Version); e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, `INSERT INTO pacientes.documento_paciente(paciente_id,tipo,numero,creado_por,actualizado_por) VALUES($1,$2,$3,$4,$4)`, result.ID, r.Document.Type, r.Document.Number, actor); e != nil {
			return e
		}
		if g := r.Guardian; g != nil {
			if _, e = tx.Exec(ctx, `INSERT INTO pacientes.apoderado(nombres,apellidos,tipo_documento,numero,creado_por,actualizado_por) VALUES($1,$2,$3,$4,$5,$5) ON CONFLICT(tipo_documento,numero_normalizado) DO NOTHING`, g.Names, g.Surnames, g.Document.Type, g.Document.Number, actor); e != nil {
				return e
			}
			var guardianID, names, surnames string
			if e = tx.QueryRow(ctx, `SELECT id::STRING,nombres,apellidos FROM pacientes.apoderado WHERE tipo_documento=$1 AND numero_normalizado=$2 FOR UPDATE`, g.Document.Type, g.Document.Number).Scan(&guardianID, &names, &surnames); e != nil {
				return e
			}
			if names != g.Names || surnames != g.Surnames {
				return app.ErrGuardianConflict
			}
			if _, e = tx.Exec(ctx, `INSERT INTO pacientes.paciente_apoderado(paciente_id,apoderado_id,parentesco,desde,creado_por,actualizado_por) VALUES($1,$2,$3,$4::DATE,$5,$5)`, result.ID, guardianID, g.Relationship, today, actor); e != nil {
				return e
			}
		}
		snapshot, e := json.Marshal(r)
		if e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, `INSERT INTO pacientes.cambio_paciente(paciente_id,actor_id,operacion,motivo,despues) VALUES($1,$2,'paciente_registrado','Alta administrativa',$3::JSONB)`, result.ID, actor, string(snapshot)); e != nil {
			return e
		}
		_, e = tx.Exec(ctx, `UPDATE pacientes.registro_idempotente SET paciente_id=$3 WHERE actor_id=$1 AND clave=$2`, actor, key, result.ID)
		return e
	})
	if err != nil {
		return app.Result{}, err
	}
	return result, nil
}
func (s Store) Find(ctx context.Context, d domain.Document) (domain.Summary, error) {
	var result domain.Summary
	err := s.transaction(ctx, func(tx pgx.Tx) error {
		result = domain.Summary{}
		e := tx.QueryRow(ctx, `SELECT p.id::STRING,p.nombres,p.apellidos,p.estado,d.tipo,d.numero FROM pacientes.paciente p JOIN pacientes.documento_paciente d ON d.paciente_id=p.id WHERE d.tipo=$1 AND d.numero_normalizado=$2`, d.Type, d.Number).Scan(&result.ID, &result.Names, &result.Surnames, &result.State, &result.Document.Type, &result.Document.Number)
		if errors.Is(e, pgx.ErrNoRows) {
			return app.ErrNotFound
		}
		return e
	})
	if err != nil {
		return domain.Summary{}, err
	}
	return result, nil
}
func (s Store) Get(ctx context.Context, id string) (domain.Patient, error) {
	var p domain.Patient
	err := s.transaction(ctx, func(tx pgx.Tx) error {
		p = domain.Patient{}
		e := tx.QueryRow(ctx, `SELECT p.id::STRING,p.version,p.estado,p.nombres,p.apellidos,p.fecha_nacimiento::STRING,coalesce(p.estado_civil,''),coalesce(p.ocupacion,''),p.direccion,p.telefono,coalesce(p.correo,''),d.tipo,d.numero
 FROM pacientes.paciente p JOIN pacientes.documento_paciente d ON d.paciente_id=p.id AND d.principal WHERE p.id=$1`, id).Scan(&p.ID, &p.Version, &p.State, &p.Names, &p.Surnames, &p.BirthDate, &p.CivilStatus, &p.Occupation, &p.Address, &p.Phone, &p.Email, &p.Document.Type, &p.Document.Number)
		if errors.Is(e, pgx.ErrNoRows) {
			return app.ErrNotFound
		}
		if e != nil {
			return e
		}
		g := &domain.Guardian{}
		e = tx.QueryRow(ctx, `SELECT a.nombres,a.apellidos,a.tipo_documento,a.numero,pa.parentesco FROM pacientes.apoderado a JOIN pacientes.paciente_apoderado pa ON pa.apoderado_id=a.id
 WHERE pa.paciente_id=$1 AND pa.principal AND pa.desde<=(now() AT TIME ZONE 'America/Lima')::DATE AND (pa.hasta IS NULL OR pa.hasta>(now() AT TIME ZONE 'America/Lima')::DATE)`, id).Scan(&g.Names, &g.Surnames, &g.Document.Type, &g.Document.Number, &g.Relationship)
		if errors.Is(e, pgx.ErrNoRows) {
			return nil
		}
		if e != nil {
			return e
		}
		p.Guardian = g
		return nil
	})
	if err != nil {
		return domain.Patient{}, err
	}
	return p, nil
}
func (s Store) Types(ctx context.Context) ([]domain.DocumentType, error) {
	var types []domain.DocumentType
	err := s.transaction(ctx, func(tx pgx.Tx) error {
		types = []domain.DocumentType{}
		rows, e := tx.Query(ctx, `SELECT codigo,nombre FROM pacientes.tipo_documento WHERE activo ORDER BY codigo LIMIT 101`)
		if e != nil {
			return e
		}
		defer rows.Close()
		for rows.Next() {
			var d domain.DocumentType
			if e = rows.Scan(&d.Code, &d.Name); e != nil {
				return e
			}
			types = append(types, d)
		}
		if len(types) > 100 {
			return app.ErrUnavailable
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return types, nil
}
