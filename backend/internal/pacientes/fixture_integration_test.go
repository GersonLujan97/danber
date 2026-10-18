//go:build integration

package pacientes_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/GersonLujan97/danber/backend/internal/platform/database"
	"github.com/jackc/pgx/v5"
)

// Each fixture creates a new database on an explicitly isolated loopback node.
func fixture(t *testing.T, through int) (context.Context, *pgx.Conn, *database.Pool) {
	t.Helper()
	adminURL, err := url.Parse(os.Getenv("DANBER_TEST_ADMIN_URL"))
	if err != nil {
		t.Fatal("invalid admin test URL")
	}
	serviceURL, err := url.Parse(os.Getenv("DANBER_TEST_DATABASE_URL"))
	if err != nil {
		t.Fatal("invalid runtime test URL")
	}
	if (adminURL.Hostname() != "127.0.0.1" && adminURL.Hostname() != "localhost") || adminURL.Host != serviceURL.Host ||
		!(strings.HasPrefix(adminURL.Path, "/danber_test_") || strings.HasPrefix(adminURL.Path, "/danber_backend_test")) || serviceURL.User == nil {
		t.Fatal("use an explicitly isolated local danber_test_* admin/runtime URL pair")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	t.Cleanup(cancel)
	admin, err := pgx.Connect(ctx, adminURL.String())
	if err != nil {
		t.Fatal("cannot connect test administrator")
	}
	name := fmt.Sprintf("danber_test_pacientes_%d", time.Now().UnixNano())
	_, err = admin.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{name}.Sanitize())
	if err != nil {
		t.Fatal("cannot create isolated test database")
	}
	admin.Close(ctx)
	adminURL.Path = "/" + name
	serviceURL.Path = "/" + name
	admin, err = pgx.Connect(ctx, adminURL.String())
	if err != nil {
		t.Fatal("cannot connect new test database")
	}
	t.Cleanup(func() { admin.Close(context.Background()) })
	_, source, _, _ := runtime.Caller(0)
	files, err := filepath.Glob(filepath.Join(filepath.Dir(source), "../../../db/migrations/[0-9][0-9][0-9]_*.sql"))
	if err != nil || len(files) != 11 {
		t.Fatal("missing migrations")
	}
	for i, path := range files {
		if i >= through {
			break
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		applyStatements(t, ctx, admin, string(raw))
	}
	_, err = admin.Exec(ctx, "GRANT CONNECT ON DATABASE "+pgx.Identifier{name}.Sanitize()+" TO "+pgx.Identifier{serviceURL.User.Username()}.Sanitize())
	if err != nil {
		t.Fatal("test connect grant failed")
	}
	pool, err := database.Open(ctx, serviceURL.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return ctx, admin, pool
}

func applyStatements(t *testing.T, ctx context.Context, admin *pgx.Conn, sql string) {
	t.Helper()
	// All migration comments are full-line comments; some contain semicolons.
	var lines []string
	for _, line := range strings.Split(sql, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "--") {
			lines = append(lines, line)
		}
	}
	sql = strings.Join(lines, "\n")
	for _, statement := range strings.Split(sql, ";") {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		if _, err := admin.Exec(ctx, statement); err != nil {
			t.Fatalf("test migration statement failed: %v", err)
		}
	}
}
