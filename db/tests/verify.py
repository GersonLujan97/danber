"""Integration checks against a real, isolated CockroachDB. Synthetic data only."""
import concurrent.futures
import hashlib
import json
import os
import subprocess
import sys
import tempfile
import threading
import time
import uuid
from pathlib import Path
from string import Template

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "scripts"))
from crdb import ROOT, execute, rows, literal, assert_version, SqlError
import migrate

DB = "danber_test_" + uuid.uuid4().hex[:12]
ACTOR = "00000000-0000-4000-8000-000000000001"
PRO1 = "20000000-0000-4000-8000-000000000001"
PRO2 = "20000000-0000-4000-8000-000000000002"
PAT1 = "40000000-0000-4000-8000-000000000001"
PAT2 = "40000000-0000-4000-8000-000000000002"
TEMPLATE = Template((ROOT / "db/tests/reservar_referencia.sql").read_text(encoding="utf-8"))
checks = []
retry_count = 0
counter_lock = threading.Lock()

def sql(text):
    return execute(text, DB)

def scalar(text):
    result = rows(text, DB)
    return list(result[0].values())[0]

def check(name, condition):
    if not condition:
        raise AssertionError(name)
    checks.append(name)
    print("PASS " + name, flush=True)

def reject(name, text, sqlstate):
    try:
        sql(text)
    except SqlError as error:
        check(name, sqlstate in str(error))
    else:
        raise AssertionError(name + ": la base aceptó SQL que debía rechazar")

def transaction(text):
    global retry_count
    for attempt in range(8):
        try:
            return sql(text)
        except SqlError as error:
            if "40001" not in str(error) and "restart transaction" not in str(error):
                raise
            with counter_lock:
                retry_count += 1
            time.sleep(0.03 * (attempt + 1))
    raise AssertionError("Reintentos agotados")

def reserve(start, end, professional=PRO1, patient=PAT1, key=None, hold=False):
    key = key or str(uuid.uuid4())
    identity = str(uuid.uuid4())
    values = dict(cita=identity, paciente=patient, profesional=professional, actor=ACTOR,
                  inicio=start, fin=end, clave=key)
    fingerprint = hashlib.sha256(json.dumps(
        [patient, professional, start, end], separators=(",", ":")).encode()).hexdigest()
    values["huella"] = fingerprint
    statement = TEMPLATE.substitute(values)
    if hold:
        statement = statement.replace(
            "INSERT INTO agenda.cita", "SELECT pg_sleep(0.5);\nINSERT INTO agenda.cita", 1)
    transaction(statement)
    found = rows("SELECT cita_id,huella FROM agenda.solicitud_idempotente WHERE actor_id=" +
                 literal(ACTOR) + " AND operacion='reservar' AND clave=" + literal(key) + ";", DB)
    return found[0]["cita_id"] if found and found[0]["huella"] == fingerprint else None

def at(day, value):
    return "2030-01-" + str(day).zfill(2) + " " + value + ":00-05:00"

def raw_cita(identifier, patient=PAT1, professional=PRO1, start=None, end=None):
    start = start or at(7,"08:00")
    end = end or at(7,"08:30")
    return ("INSERT INTO agenda.cita(id,paciente_id,profesional_id,inicio,fin) VALUES(" +
            ",".join(map(literal,[identifier,patient,professional,start,end])) + ");")

def run():
    version = assert_version("defaultdb")
    print("Base efímera de pruebas: " + DB, flush=True)
    # Upgrade path: core schemas first, then remaining migrations.
    migrate.init(DB, 3)
    check("Actualización desde versión 3", scalar("SELECT count(*) FROM danber_meta.migracion;") == "3")
    migrate.apply(DB, None)
    migrate.apply(DB, None)
    check("Migraciones y reejecución sin cambios", scalar("SELECT count(*) FROM danber_meta.migracion WHERE estado='aplicada';") == str(len(migrate.entries())))
    sql((ROOT / "db/tests/fixtures.sql").read_text(encoding="utf-8"))
    check("Datos iniciales sin pacientes reales", scalar("SELECT count(*) FROM pacientes.paciente;") == "2")

    # SQL integrity, identity and conditional indexes.
    reject("Paciente sin dirección", "INSERT INTO pacientes.paciente(nombres,apellidos,fecha_nacimiento,direccion,telefono) VALUES('Ficticio','Sin dirección','1990-01-01',' ','000');", "23514")
    reject("DNI duplicado", "INSERT INTO pacientes.documento_paciente(paciente_id,tipo,numero) VALUES(" + literal(PAT2) + ",'DNI','00000101');", "23505")
    reject("Documento principal único", "INSERT INTO pacientes.documento_paciente(paciente_id,tipo,numero) VALUES(" + literal(PAT1) + ",'DNI','00000999');", "23505")
    reject("DNI con formato inválido", "INSERT INTO pacientes.apoderado(nombres,apellidos,tipo_documento,numero) VALUES('A','Ficticio','DNI','123');", "23514")
    sql("INSERT INTO pacientes.tipo_documento(codigo,nombre) VALUES('PASAPORTE_PRUEBA','Solo prueba');")
    sql("INSERT INTO pacientes.documento_paciente(paciente_id,tipo,numero,principal) VALUES(" + literal(PAT2) + ",'PASAPORTE_PRUEBA','00000101',false);")
    check("Igual número con otro tipo no fusiona", scalar("SELECT count(*) FROM pacientes.documento_paciente WHERE numero='00000101';") == "2")
    reject("Documento de tipo no configurado", "INSERT INTO pacientes.apoderado(nombres,apellidos,tipo_documento,numero) VALUES('A','Ficticio','INEXISTENTE','ABC');", "23503")
    reject("Reserva menor a 30 minutos", raw_cita(str(uuid.uuid4()), start=at(7,"08:00"),end=at(7,"08:29")), "23514")
    reject("Paciente inexistente en cita", raw_cita(str(uuid.uuid4()),patient=str(uuid.uuid4())), "23503")
    reject("Sin cascada de historia del paciente", "DELETE FROM pacientes.paciente WHERE id=" + literal(PAT1) + ";", "23503")

    first = reserve(at(7,"09:00"), at(7,"09:30"), key="primera")
    check("Reserva válida con historia y recibo", first is not None and scalar("SELECT count(*) FROM agenda.cambio_cita WHERE cita_id=" + literal(first) + ";") == "1")
    repeat = reserve(at(7,"09:00"),at(7,"09:30"),key="primera")
    check("Idempotencia devuelve misma cita", repeat == first and scalar("SELECT count(*) FROM agenda.cambio_cita WHERE cita_id=" + literal(first) + ";") == "1")
    check("Clave reutilizada con otro contenido no crea cita", reserve(at(7,"11:00"),at(7,"11:30"),key="primera") is None)
    check("Adyacencia permitida", reserve(at(7,"09:30"),at(7,"10:00")) is not None)
    check("Solapamiento parcial rechazado por protocolo", reserve(at(7,"09:15"),at(7,"09:45")) is None)
    check("Almuerzo respetado", reserve(at(7,"12:15"),at(7,"12:45")) is None)
    check("Cierre respetado", reserve(at(7,"18:45"),at(7,"19:15")) is None)
    check("Sábado no se inventa cierre", reserve(at(12,"18:00"),at(12,"18:30")) is None)

    # Race, not just sequential duplicate insertion.
    barrier = threading.Barrier(2)
    def race(patient):
        barrier.wait(timeout=10)
        return reserve(at(8,"10:00"),at(8,"11:00"),patient=patient,hold=True)
    with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
        results = list(pool.map(race,[PAT1,PAT2]))
    check("Dos conexiones competidoras: una reserva", sum(x is not None for x in results) == 1)
    check("Otro odontólogo puede reservar a la misma hora", reserve(at(8,"10:00"),at(8,"11:00"),professional=PRO2) is not None)
    check("Sin solapamientos finales", scalar("""SELECT count(*) FROM agenda.cita a JOIN agenda.cita b
     ON a.id < b.id AND a.profesional_id=b.profesional_id AND a.ocupa_intervalo AND b.ocupa_intervalo
     AND a.inicio < b.fin AND b.inicio < a.fin;""") == "0")

    # Blocks keep existing appointments; exceptions override a weekday.
    sql("BEGIN; UPDATE agenda.control_profesional SET revision=revision+1 WHERE profesional_id=" + literal(PRO1) +
        "; INSERT INTO agenda.bloqueo(profesional_id,inicio,fin,motivo) VALUES(" +
        literal(PRO1) + "," + literal(at(7,"08:00")) + "," + literal(at(10,"00:00")) + ",'Viaje ficticio'); COMMIT;")
    check("Bloqueo conserva cita existente", scalar("SELECT count(*) FROM agenda.cita WHERE id=" + literal(first) + ";") == "1")
    check("Bloqueo impide reserva posterior", reserve(at(9,"15:00"),at(9,"15:30")) is None)
    exception = str(uuid.uuid4())
    sql("INSERT INTO agenda.excepcion_horario(id,profesional_id,fecha,motivo) VALUES(" +
        literal(exception) + "," + literal(PRO1) + ",'2030-01-12','Sábado de prueba');" +
        "INSERT INTO agenda.tramo_excepcion(excepcion_id,inicio_minuto,fin_minuto) VALUES(" + literal(exception) + ",1080,1200);")
    check("Sábado con cierre explícito permite cita", reserve(at(12,"18:00"),at(12,"18:30")) is not None)

    # Atomic rollback of a release before a failed assignment.
    reject("Liberación y asignación fallida revierten juntas",
           "BEGIN; UPDATE agenda.control_profesional SET revision=revision+1 WHERE profesional_id=" + literal(PRO1) +
           "; UPDATE agenda.cita SET ocupa_intervalo=false WHERE id=" + literal(first) +
           ";" + raw_cita(str(uuid.uuid4()),start=at(7,"09:00"),end=at(7,"09:20")) + "COMMIT;", "23514")
    check("Reserva original permanece tras rollback", scalar("SELECT ocupa_intervalo::STRING FROM agenda.cita WHERE id=" + literal(first) + ";") == "true")

    # Reprogram conflict uses the same guard and all-or-nothing transaction.
    before = scalar("SELECT inicio::STRING FROM agenda.cita WHERE id=" + literal(first) + ";")
    sql("BEGIN; UPDATE agenda.control_profesional SET revision=revision+1 WHERE profesional_id=" + literal(PRO1) +
        "; UPDATE agenda.cita c SET inicio=" + literal(at(7,"09:30")) + ",fin=" + literal(at(7,"10:00")) +
        " WHERE c.id=" + literal(first) + " AND NOT EXISTS (SELECT 1 FROM agenda.cita other WHERE other.id<>c.id AND other.profesional_id=c.profesional_id AND other.ocupa_intervalo AND other.inicio<" + literal(at(7,"10:00")) + "::TIMESTAMPTZ AND " + literal(at(7,"09:30")) + "::TIMESTAMPTZ<other.fin); COMMIT;")
    check("Conflicto de reprogramación conserva horario", scalar("SELECT inicio::STRING FROM agenda.cita WHERE id=" + literal(first) + ";") == before)

    # One active attendance per dentist, independent from planned end time.
    sql("UPDATE agenda.cita SET llegada_en='2030-01-07 09:00-05',inicio_real='2030-01-07 09:00-05' WHERE id=" + literal(first) + ";")
    adjacent = scalar("SELECT id::STRING FROM agenda.cita WHERE profesional_id=" + literal(PRO1) + " AND inicio=" + literal(at(7,"09:30")) + ";")
    reject("No iniciar dos atenciones simultáneas",
           "UPDATE agenda.cita SET llegada_en='2030-01-07 09:30-05',inicio_real='2030-01-07 09:30-05' WHERE id=" + literal(adjacent) + ";", "23505")
    reject("Espera no puede enlazar paciente de otra cita",
           "INSERT INTO agenda.espera(paciente_id,profesional_id,cita_origen_id,llegada_en,motivo) VALUES(" +
           literal(PAT2) + "," + literal(PRO1) + "," + literal(first) + ",now(),'tardanza');", "23503")
    reject("Runtime no edita historial",
           "SET ROLE danber_runtime; UPDATE agenda.cambio_cita SET motivo='alterado' WHERE cita_id=" + literal(first) + ";", "42501")
    reject("Runtime no modifica versiones de migración",
           "SET ROLE danber_runtime; DELETE FROM danber_meta.migracion;", "42501")
    reject("Runtime no crea tablas en public",
           "SET ROLE danber_runtime; CREATE TABLE public.forbidden_runtime_sql_test(id INT PRIMARY KEY);", "42501")
    sql("INSERT INTO identidad.usuario_rol(usuario_id,rol_id) SELECT " + literal(ACTOR) + ",id FROM identidad.rol WHERE codigo IN ('maestro','recepcion');")
    check("Una cuenta tiene Maestro y Recepción", scalar("SELECT count(*) FROM identidad.usuario_rol WHERE usuario_id=" + literal(ACTOR) + ";") == "2")

    # Checksum and dirty guards refuse further DDL.
    original_sha = scalar("SELECT sha256 FROM danber_meta.migracion WHERE version=1;")
    sql("UPDATE danber_meta.migracion SET sha256=repeat('0',64) WHERE version=1;")
    try:
        migrate.apply(DB,None)
    except RuntimeError as error:
        check("Checksum divergente detiene ejecutor", "Checksum" in str(error))
    else:
        raise AssertionError("Checksum no detectado")
    sql("UPDATE danber_meta.migracion SET sha256=" + literal(original_sha) + ",estado='aplicando' WHERE version=1;")
    try:
        migrate.apply(DB,None)
    except RuntimeError as error:
        check("Migración incompleta detiene ejecutor", "incompleta" in str(error))
    else:
        raise AssertionError("Estado incompleto no detectado")
    sql("UPDATE danber_meta.migracion SET estado='aplicada' WHERE version=1;")

    # Full standalone script creates an independent database and is adopted without drift.
    exported_db = "danber_test_" + uuid.uuid4().hex[:12]
    target = ROOT / ".local/export-verification.sql"
    migrate.export(target,exported_db)
    execute(target.read_text(encoding="utf-8"),"defaultdb")
    migrate.apply(exported_db,None)
    check("SQL único desde cero coincide con migraciones",
          rows("SELECT count(*) AS total FROM danber_meta.migracion WHERE estado='aplicada';",exported_db)[0]["total"] == str(len(migrate.entries())))

    report = dict(server=version,database=DB,standalone_database=exported_db,
                  checks=checks,total=len(checks),serialization_retries=retry_count,
                  scope="DDL, integridad y protocolo SQL de referencia; no API ni autorización por recurso",
                  completed_utc=time.strftime("%Y-%m-%dT%H:%M:%SZ",time.gmtime()))
    report_path = ROOT / "db/tests/resultado-verificacion.json"
    report_path.write_text(json.dumps(report,ensure_ascii=False,indent=2) + "\n",encoding="utf-8")
    print("RESULTADO: " + str(len(checks)) + " comprobaciones correctas; reintentos " + str(retry_count),flush=True)

if __name__ == "__main__":
    try:
        run()
    except Exception as error:
        print("FAIL " + str(error),file=sys.stderr)
        sys.exit(1)
