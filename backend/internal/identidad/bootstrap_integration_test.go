//go:build integration

package identidad_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/GersonLujan97/danber/backend/internal/identidad"
	"github.com/GersonLujan97/danber/backend/internal/platform/database"
	"github.com/jackc/pgx/v5"
)

// Each case creates a NEW test database on the explicitly configured local test node.
// No DROP/TRUNCATE, and no writes to the database named in the supplied URL.
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
	name := fmt.Sprintf("danber_test_cuenta_%d", time.Now().UnixNano())
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

func count(t *testing.T, ctx context.Context, c *pgx.Conn, table string, want int) {
	t.Helper()
	var n int
	if err := c.QueryRow(ctx, "SELECT count(*) FROM "+table).Scan(&n); err != nil || n != want {
		t.Fatalf("%s count=%d want=%d err=%v", table, n, want, err)
	}
}

func TestConcurrentInitialAccountIsAtomicAndRepeatSafe(t *testing.T) {
	ctx, admin, pool := fixture(t, 9)
	start := make(chan struct{})
	results := make(chan error, 4)
	var wg sync.WaitGroup
	for i := range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := identidad.NewInitialAccount(pool).Create(ctx, fmt.Sprintf("Ficticio%d", i), []byte("frase ficticia para probar"), "operador_pruebas")
			results <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	success, rejected := 0, 0
	for err := range results {
		switch {
		case err == nil:
			success++
		case errors.Is(err, identidad.ErrInitialized):
			rejected++
		default:
			t.Fatalf("unexpected concurrent result: %v", err)
		}
	}
	if success != 1 || rejected != 3 {
		t.Fatalf("success=%d rejected=%d", success, rejected)
	}
	for table, want := range map[string]int{"identidad.usuario": 1, "identidad.credencial": 1, "identidad.usuario_rol": 2, "identidad.cambio_acceso": 1, "identidad.sesion": 0, "personal.profesional": 0} {
		count(t, ctx, admin, table, want)
	}
	var valid bool
	err := admin.QueryRow(ctx, `SELECT a.usuario_id=u.id AND a.completado_en IS NOT NULL AND a.revision=1
 AND c.hash_clave LIKE '$argon2id$v=19$m=19456,t=2,p=1$%' AND u.estado='activo'
 FROM identidad.aprovisionamiento_inicial a JOIN identidad.usuario u ON u.id=a.usuario_id JOIN identidad.credencial c ON c.usuario_id=u.id`).Scan(&valid)
	if err != nil || !valid {
		t.Fatal("incomplete provisioning")
	}
	var audit string
	if err = admin.QueryRow(ctx, `SELECT detalle::STRING FROM identidad.cambio_acceso WHERE actor_id=usuario_id AND operacion='cuenta_inicial_creada'`).Scan(&audit); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(audit, "argon2") || strings.Contains(audit, "frase ficticia") || !strings.Contains(audit, "aprovisionamiento_sin_sesion") || !strings.Contains(audit, "operador_pruebas") {
		t.Fatal("invalid or unsafe audit")
	}
	var beforeHash, afterHash string
	admin.QueryRow(ctx, `SELECT hash_clave FROM identidad.credencial`).Scan(&beforeHash)
	_, err = identidad.NewInitialAccount(pool).Create(ctx, "Otro", []byte("otra contraseña ficticia"), "operador_pruebas")
	if !errors.Is(err, identidad.ErrInitialized) {
		t.Fatal("repeat must be refused")
	}
	admin.QueryRow(ctx, `SELECT hash_clave FROM identidad.credencial`).Scan(&afterHash)
	if beforeHash != afterHash {
		t.Fatal("repeat changed existing password")
	}
	count(t, ctx, admin, "identidad.cambio_acceso", 1)
}

func TestExistingUsernameIsNotElevated(t *testing.T) {
	ctx, admin, pool := fixture(t, 9)
	if _, err := admin.Exec(ctx, `INSERT INTO identidad.usuario(nombre_acceso) VALUES('Ficticio')`); err != nil {
		t.Fatal(err)
	}
	_, err := identidad.NewInitialAccount(pool).Create(ctx, " ficticio ", []byte("frase ficticia para probar"), "operador_pruebas")
	if err == nil || !strings.Contains(err.Error(), "ya existe") {
		t.Fatalf("duplicate name not refused: %v", err)
	}
	count(t, ctx, admin, "identidad.usuario", 1)
	count(t, ctx, admin, "identidad.usuario_rol", 0)
	count(t, ctx, admin, "identidad.credencial", 0)
	assertUnused(t, ctx, admin)
}

func TestAuditFailureRollsBackEntireAccount(t *testing.T) {
	ctx, admin, pool := fixture(t, 9)
	if _, err := admin.Exec(ctx, `ALTER TABLE identidad.cambio_acceso ADD CONSTRAINT fallo_solo_prueba CHECK (operacion <> 'cuenta_inicial_creada')`); err != nil {
		t.Fatal(err)
	}
	_, err := identidad.NewInitialAccount(pool).Create(ctx, "Ficticio", []byte("frase ficticia para probar"), "operador_pruebas")
	if err == nil || strings.Contains(err.Error(), "fallo_solo_prueba") {
		t.Fatal("audit failure not sanitized")
	}
	for _, table := range []string{"identidad.usuario", "identidad.credencial", "identidad.usuario_rol", "identidad.cambio_acceso"} {
		count(t, ctx, admin, table, 0)
	}
	assertUnused(t, ctx, admin)
	if _, err = admin.Exec(ctx, `ALTER TABLE identidad.cambio_acceso DROP CONSTRAINT fallo_solo_prueba`); err != nil {
		t.Fatal(err)
	}
	if _, err = identidad.NewInitialAccount(pool).Create(ctx, "Ficticio", []byte("frase ficticia para probar"), "operador_pruebas"); err != nil {
		t.Fatal("clean retry after rollback failed")
	}
}

func TestMissingMigrationAndUpgradeFromEight(t *testing.T) {
	ctx, admin, pool := fixture(t, 8)
	_, err := identidad.NewInitialAccount(pool).Create(ctx, "Ficticio", []byte("frase ficticia para probar"), "operador_pruebas")
	if err == nil || !strings.Contains(err.Error(), "migración 009") {
		t.Fatal("missing migration not detected")
	}
	count(t, ctx, admin, "identidad.usuario", 0)
	_, source, _, _ := runtime.Caller(0)
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(source), "../../../db/migrations/009_aprovisionamiento_inicial.sql"))
	if err != nil {
		t.Fatal(err)
	}
	applyStatements(t, ctx, admin, string(raw))
	if _, err = identidad.NewInitialAccount(pool).Create(ctx, "Ficticio", []byte("frase ficticia para probar"), "operador_pruebas"); err != nil {
		t.Fatal(err)
	}
}

func TestExistingMasterWithoutMarkerIsRefused(t *testing.T) {
	ctx, admin, pool := fixture(t, 9)
	if _, err := admin.Exec(ctx, `INSERT INTO identidad.usuario(id,nombre_acceso,estado) VALUES('00000000-0000-4000-8000-000000000001','Ficticio','suspendido');
 INSERT INTO identidad.usuario_rol(usuario_id,rol_id) SELECT '00000000-0000-4000-8000-000000000001',id FROM identidad.rol WHERE codigo='maestro'`); err != nil {
		t.Fatal(err)
	}
	_, err := identidad.NewInitialAccount(pool).Create(ctx, "Otro", []byte("frase ficticia para probar"), "operador_pruebas")
	if !errors.Is(err, identidad.ErrInitialized) {
		t.Fatal("existing suspended master not detected")
	}
	count(t, ctx, admin, "identidad.usuario", 1)
	assertUnused(t, ctx, admin)
}

func assertUnused(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	var unused bool
	if err := admin.QueryRow(ctx, `SELECT usuario_id IS NULL AND completado_en IS NULL AND revision=0 FROM identidad.aprovisionamiento_inicial`).Scan(&unused); err != nil || !unused {
		t.Fatal("failed transaction consumed bootstrap guard")
	}
}

// Match the production CLI migrator's statement-by-statement commits. Sending
// a whole file in one pgx Exec creates an implicit transaction: CockroachDB
// cannot insert into a newly added FK table until its schema job has committed.
// These fixture files contain simple statements, without semicolons in literals.
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
