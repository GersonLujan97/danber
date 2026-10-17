//go:build integration

package database_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/GersonLujan97/danber/backend/internal/health"
	"github.com/GersonLujan97/danber/backend/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestCockroachConnectivityAndReadiness(t *testing.T) {
	url := os.Getenv("DANBER_TEST_DATABASE_URL")
	if url == "" {
		t.Fatal("DANBER_TEST_DATABASE_URL is required; use an isolated CockroachDB")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := database.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	if err := pool.Check(ctx); err != nil {
		pool.Close()
		t.Fatal("ping failed")
	}
	probe := health.Handler(pool.Check)
	rec := httptest.NewRecorder()
	probe.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if rec.Code != 200 {
		pool.Close()
		t.Fatal("real database not ready")
	}
	pool.Close()
	rec = httptest.NewRecorder()
	probe.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if rec.Code != 503 {
		t.Fatal("closed database was reported ready")
	}
	rec = httptest.NewRecorder()
	probe.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health/live", nil))
	if rec.Code != 200 {
		t.Fatal("liveness incorrectly depends on SQL")
	}
}

func TestAdministratorConnectionRejected(t *testing.T) {
	url := os.Getenv("DANBER_TEST_ADMIN_URL")
	if url == "" {
		t.Fatal("DANBER_TEST_ADMIN_URL is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := database.Open(ctx, url)
	if pool != nil {
		pool.Close()
	}
	if err == nil {
		t.Fatal("admin pool accepted")
	}
}

func TestSQLRuntimeCannotCreateTables(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, os.Getenv("DANBER_TEST_DATABASE_URL"))
	if err != nil {
		t.Fatal("test connection failed")
	}
	defer conn.Close(context.Background())
	// Unique name avoids confusing an existing-table error with denied privileges.
	name := fmt.Sprintf("forbidden_runtime_probe_%d", time.Now().UnixNano())
	_, err = conn.Exec(ctx, "CREATE TABLE public."+name+"(id INT PRIMARY KEY)")
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "42501" {
		t.Fatal("runtime DDL must fail specifically with insufficient_privilege")
	}
}
