// Package database owns the connection adapter; future business rules do not import pgx.
package database

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool struct{ pool *pgxpool.Pool }

func Open(ctx context.Context, connectionURL string) (*Pool, error) {
	cfg, err := pgxpool.ParseConfig(connectionURL)
	if err != nil {
		return nil, errors.New("configuración SQL inválida")
	}
	cfg.MaxConns = 4
	cfg.MinConns = 0
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second
	cfg.ConnConfig.RuntimeParams["application_name"] = "danber-api"
	p, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, errors.New("no se pudo crear el pool SQL")
	}
	fail := func(message string) (*Pool, error) { p.Close(); return nil, errors.New(message) }
	if err := p.Ping(ctx); err != nil {
		return fail("no se pudo conectar con CockroachDB")
	}
	var version string
	if err := p.QueryRow(ctx, "SELECT version()").Scan(&version); err != nil {
		return fail("no se pudo comprobar la versión del motor")
	}
	if !strings.Contains(version, "CockroachDB") || !strings.Contains(version, " v26.2.7 ") {
		return fail("se requiere CockroachDB 26.2.7, versión verificada de esta entrega")
	}
	// Reject privileged SQL identities even when their name is not root.
	var admin bool
	if err := p.QueryRow(ctx, "SELECT pg_has_role(current_user, 'admin', 'MEMBER')").Scan(&admin); err != nil {
		return fail("no se pudieron comprobar privilegios de la conexión")
	}
	if admin {
		return fail("la conexión de la API no debe tener privilegios de administrador")
	}
	return &Pool{pool: p}, nil
}

func (p *Pool) Check(ctx context.Context) error { return p.pool.Ping(ctx) }
func (p *Pool) Close()                          { p.pool.Close() }
