"""Cockroach CLI adapter; Python standard library only. Never logs connection URLs."""
import csv
import io
import os
import re
import subprocess
from pathlib import Path
from urllib.parse import urlsplit, urlunsplit, quote

ROOT = Path(__file__).resolve().parents[2]
EXPECTED_VERSION = "26.2.7"

class SqlError(RuntimeError):
    pass

def binary():
    value = os.environ.get("COCKROACH_BIN")
    if value:
        return value
    local = ROOT / ".local/cockroach/cockroach-v26.2.7.windows-6.2-amd64/cockroach.exe"
    return str(local) if local.exists() else "cockroach"

def connection_url(database=None):
    value = os.environ.get("DANBER_DATABASE_URL")
    if not value:
        raise RuntimeError("Define DANBER_DATABASE_URL; consulta db/README.md.")
    parts = urlsplit(value)
    if parts.scheme not in ("postgres", "postgresql"):
        raise RuntimeError("La conexión debe usar postgres:// o postgresql://.")
    if database:
        if not re.fullmatch(r"[a-z][a-z0-9_]{0,62}", database):
            raise RuntimeError("Nombre de base inválido.")
        value = urlunsplit((parts.scheme, parts.netloc, "/" + quote(database), parts.query, parts.fragment))
    return value

def execute(sql, database=None, timeout=120):
    url = connection_url(database)
    try:
        result = subprocess.run(
            [binary(), "sql", "--url=" + url, "--format=csv",
             "--set=errexit=true", "--set=show_times=false"],
            input=sql, text=True, encoding="utf-8", capture_output=True, timeout=timeout)
    except subprocess.TimeoutExpired:
        raise SqlError("Tiempo de espera agotado; revisar el estado real de la operación antes de repetirla.") from None
    if result.returncode:
        # No command line or connection string in errors.
        message = result.stderr.replace(url, "<conexion>")
        raise SqlError(message.strip())
    return result.stdout

def rows(sql, database=None):
    output = execute(sql, database)
    return list(csv.DictReader(io.StringIO(output)))

def assert_version(database=None):
    value = rows("SELECT version() AS version;", database)[0]["version"]
    if not re.search(r"\bv" + re.escape(EXPECTED_VERSION) + r"\b", value):
        raise RuntimeError("Se requiere CockroachDB v" + EXPECTED_VERSION +
                           "; no se aplicó SQL de migración. Verifica la versión del servidor.")
    return value

def literal(value):
    return "'" + str(value).replace("'", "''") + "'"
