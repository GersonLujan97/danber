"""Versioned migrations, standalone SQL export, and checksums. No third-party dependencies."""
import argparse
import hashlib
import sys
from pathlib import Path

from crdb import ROOT, execute, rows, assert_version, literal, SqlError

MIGRATIONS = ROOT / "db/migrations"
META = """
CREATE SCHEMA IF NOT EXISTS danber_meta;
CREATE TABLE IF NOT EXISTS danber_meta.migracion (
 version INT8 PRIMARY KEY,
 nombre STRING NOT NULL,
 sha256 STRING(64) NOT NULL,
 estado STRING NOT NULL CHECK (estado IN ('aplicando','aplicada')),
 iniciada_en TIMESTAMPTZ NOT NULL DEFAULT now(),
 aplicada_en TIMESTAMPTZ NULL
);
"""
def entries():
    result = []
    for path in sorted(MIGRATIONS.glob("[0-9][0-9][0-9]_*.sql")):
        raw = path.read_bytes()
        result.append((int(path.name[:3]), path.name, hashlib.sha256(raw).hexdigest(), raw.decode("utf-8")))
    if not result or [v[0] for v in result] != list(range(1, len(result) + 1)):
        raise RuntimeError("Secuencia de migraciones inválida.")
    return result

def start_sql(v, name, checksum):
    return ("INSERT INTO danber_meta.migracion(version,nombre,sha256,estado) VALUES (" +
            str(v) + "," + literal(name) + "," + literal(checksum) + ",'aplicando');")

def finish_sql(v):
    return ("UPDATE danber_meta.migracion SET estado='aplicada', aplicada_en=now() " +
            "WHERE version=" + str(v) + " AND estado='aplicando';")

def apply(database, through):
    version = assert_version(database)
    execute(META, database)
    recorded = rows("SELECT version,nombre,sha256,estado FROM danber_meta.migracion ORDER BY version;", database)
    expected = {str(v): (name, sha) for v, name, sha, _ in entries()}
    for item in recorded:
        if item["version"] not in expected:
            raise RuntimeError("La base contiene una versión desconocida por estos archivos.")
        if (item["nombre"], item["sha256"]) != expected[item["version"]]:
            raise RuntimeError("Checksum/nombre modificado en migración " + item["version"] + ". No se continúa.")
        if item["estado"] != "aplicada":
            raise RuntimeError("Migración incompleta " + item["version"] +
                               ". Revisar estado real; no se reejecuta ni se marca aplicada automáticamente.")
    done = {int(item["version"]) for item in recorded}
    if done and done != set(range(1, max(done) + 1)):
        raise RuntimeError("Historial no contiguo.")
    for v, name, sha, sql in entries():
        if through is not None and v > through:
            break
        if v in done:
            print("Sin cambios: " + name)
            continue
        # Claim is committed before DDL. A competing migrator cannot claim the same version.
        execute(start_sql(v, name, sha), database)
        try:
            execute(sql, database)
            execute(finish_sql(v), database)
        except Exception:
            print("Detenido en " + name + "; queda aplicando para revisión manual.", file=sys.stderr)
            raise
        print("Aplicada: " + name)
    print("Servidor comprobado: " + version)

def init(database, through):
    assert_version("defaultdb")
    exists = rows("SELECT database_name FROM [SHOW DATABASES] WHERE database_name=" + literal(database) + ";", "defaultdb")
    if not exists:
        execute('CREATE DATABASE "' + database + '";', "defaultdb")
        print("Base creada: " + database)
    # Existing schema is never erased; apply checks history.
    apply(database, through)

def export(path, database):
    parts = [
        "-- Danber: instalacion inicial en base NUEVA, CockroachDB v26.2.0 o v26.2.7.",
        "-- Ejecutar con errexit; DDL no es una transaccion global.",
        "-- No borrar una base existente para volver a ejecutar.",
        'CREATE DATABASE "' + database + '";',
        'SET DATABASE = "' + database + '";',
        META]
    for v, name, sha, sql in entries():
        parts += ["-- " + name, start_sql(v, name, sha), sql, finish_sql(v)]
    parts += ["SELECT version,nombre,sha256,estado FROM danber_meta.migracion ORDER BY version;"]
    with path.open("w", encoding="utf-8", newline="\n") as output:
        output.write("\n\n".join(parts) + "\n")
    print("Script generado: " + str(path))

def main():
    parser = argparse.ArgumentParser(description="Migraciones Danber / CockroachDB v26.2.0 o v26.2.7")
    parser.add_argument("command", choices=["init", "apply", "status", "export"])
    parser.add_argument("--database", default="danber")
    parser.add_argument("--through", type=int)
    parser.add_argument("--output", type=Path, default=ROOT / "db/danber-inicial.sql")
    args = parser.parse_args()
    # Validate database identifier even in export mode.
    import re
    if not re.fullmatch(r"[a-z][a-z0-9_]{0,62}", args.database):
        parser.error("Nombre de base inválido.")
    if args.command == "export":
        export(args.output, args.database)
    elif args.command == "init":
        init(args.database, args.through)
    elif args.command == "apply":
        apply(args.database, args.through)
    else:
        assert_version(args.database)
        for item in rows("SELECT version,nombre,sha256,estado FROM danber_meta.migracion ORDER BY version;", args.database):
            print(item["version"], item["estado"], item["nombre"], item["sha256"])

if __name__ == "__main__":
    try:
        main()
    except Exception as error:
        # Errors are sanitized by the adapter; never echo arguments or environment.
        print(str(error), file=sys.stderr)
        sys.exit(1)
