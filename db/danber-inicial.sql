-- Danber: instalacion inicial en base NUEVA, CockroachDB v26.2.0 o v26.2.7.

-- Ejecutar con errexit; DDL no es una transaccion global.

-- No borrar una base existente para volver a ejecutar.

CREATE DATABASE "danber";

SET DATABASE = "danber";


CREATE SCHEMA IF NOT EXISTS danber_meta;
CREATE TABLE IF NOT EXISTS danber_meta.migracion (
 version INT8 PRIMARY KEY,
 nombre STRING NOT NULL,
 sha256 STRING(64) NOT NULL,
 estado STRING NOT NULL CHECK (estado IN ('aplicando','aplicada')),
 iniciada_en TIMESTAMPTZ NOT NULL DEFAULT now(),
 aplicada_en TIMESTAMPTZ NULL
);


-- 001_esquemas.sql

INSERT INTO danber_meta.migracion(version,nombre,sha256,estado) VALUES (1,'001_esquemas.sql','a4b5f94f77cea8ca8546159b9952b66a3818c98d06f97f7578230ca597afc090','aplicando');

-- Solo los cuatro propietarios del primer recorrido.
CREATE SCHEMA identidad;
CREATE SCHEMA personal;
CREATE SCHEMA pacientes;
CREATE SCHEMA agenda;


UPDATE danber_meta.migracion SET estado='aplicada', aplicada_en=now() WHERE version=1 AND estado='aplicando';

-- 002_identidad.sql

INSERT INTO danber_meta.migracion(version,nombre,sha256,estado) VALUES (2,'002_identidad.sql','00cb63c33fc74db01f1ebd84d481348e724209a2644284427c19151dbcd45425','aplicando');

CREATE TABLE identidad.usuario (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  nombre_acceso STRING(150) NOT NULL CHECK (length(btrim(nombre_acceso)) > 0),
  nombre_acceso_normalizado STRING(150) AS (lower(btrim(nombre_acceso))) STORED,
  estado STRING(20) NOT NULL DEFAULT 'activo' CHECK (estado IN ('activo','suspendido')),
  version INT8 NOT NULL DEFAULT 1 CHECK (version > 0),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (nombre_acceso_normalizado)
);
CREATE TABLE identidad.credencial (
  usuario_id UUID PRIMARY KEY REFERENCES identidad.usuario(id),
  hash_clave STRING(500) NOT NULL CHECK (length(btrim(hash_clave)) > 0),
  cambiado_en TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE identidad.rol (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  codigo STRING(50) NOT NULL UNIQUE CHECK (codigo IN ('maestro','recepcion','odontologo')),
  nombre STRING(100) NOT NULL CHECK (length(btrim(nombre)) > 0)
);
CREATE TABLE identidad.permiso (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  codigo STRING(100) NOT NULL UNIQUE,
  descripcion STRING(200) NOT NULL CHECK (length(btrim(descripcion)) > 0)
);
CREATE TABLE identidad.usuario_rol (
  usuario_id UUID NOT NULL REFERENCES identidad.usuario(id),
  rol_id UUID NOT NULL REFERENCES identidad.rol(id),
  asignado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  asignado_por UUID NULL REFERENCES identidad.usuario(id),
  PRIMARY KEY (usuario_id, rol_id)
);
CREATE TABLE identidad.rol_permiso (
  rol_id UUID NOT NULL REFERENCES identidad.rol(id),
  permiso_id UUID NOT NULL REFERENCES identidad.permiso(id),
  PRIMARY KEY (rol_id, permiso_id)
);
CREATE TABLE identidad.sesion (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  usuario_id UUID NOT NULL REFERENCES identidad.usuario(id),
  token_hash BYTES NOT NULL UNIQUE CHECK (length(token_hash) >= 32),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  expira_en TIMESTAMPTZ NOT NULL,
  revocado_en TIMESTAMPTZ NULL,
  CHECK (expira_en > creado_en)
);
CREATE TABLE identidad.cambio_acceso (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  usuario_id UUID NOT NULL REFERENCES identidad.usuario(id),
  actor_id UUID NOT NULL REFERENCES identidad.usuario(id),
  operacion STRING(100) NOT NULL CHECK (length(btrim(operacion)) > 0),
  fecha TIMESTAMPTZ NOT NULL DEFAULT now(),
  detalle JSONB NOT NULL
);


UPDATE danber_meta.migracion SET estado='aplicada', aplicada_en=now() WHERE version=2 AND estado='aplicando';

-- 003_personal.sql

INSERT INTO danber_meta.migracion(version,nombre,sha256,estado) VALUES (3,'003_personal.sql','314e36d9ea3d2f1f2f06ed935c3911d16d29900619176d3879b87c60988f12bd','aplicando');

CREATE TABLE personal.personal (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  nombres STRING(150) NOT NULL CHECK (length(btrim(nombres)) > 0),
  apellidos STRING(150) NOT NULL CHECK (length(btrim(apellidos)) > 0),
  usuario_id UUID NULL UNIQUE REFERENCES identidad.usuario(id),
  estado STRING(20) NOT NULL DEFAULT 'activo' CHECK (estado IN ('activo','inactivo')),
  version INT8 NOT NULL DEFAULT 1 CHECK (version > 0),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  creado_por UUID NULL REFERENCES identidad.usuario(id),
  actualizado_por UUID NULL REFERENCES identidad.usuario(id)
);
CREATE TABLE personal.profesional (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  personal_id UUID NOT NULL UNIQUE REFERENCES personal.personal(id),
  dni STRING(8) NOT NULL UNIQUE CHECK (dni ~ '^[0-9]{8}$'),
  estado STRING(20) NOT NULL DEFAULT 'pendiente' CHECK (estado IN ('pendiente','habilitado_agenda','inactivo')),
  version INT8 NOT NULL DEFAULT 1 CHECK (version > 0),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  creado_por UUID NULL REFERENCES identidad.usuario(id),
  actualizado_por UUID NULL REFERENCES identidad.usuario(id)
);
CREATE TABLE personal.especialidad (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  nombre STRING(120) NOT NULL CHECK (length(btrim(nombre)) > 0),
  nombre_normalizado STRING(120) AS (lower(btrim(nombre))) STORED,
  UNIQUE(nombre_normalizado),
  version INT8 NOT NULL DEFAULT 1 CHECK (version > 0),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  creado_por UUID NULL REFERENCES identidad.usuario(id),
  actualizado_por UUID NULL REFERENCES identidad.usuario(id)
);
CREATE TABLE personal.profesional_especialidad (
  profesional_id UUID NOT NULL REFERENCES personal.profesional(id),
  especialidad_id UUID NOT NULL REFERENCES personal.especialidad(id),
  PRIMARY KEY (profesional_id, especialidad_id)
);
-- habilitado_agenda es elegibilidad operativa, no acreditación clínica/legal.


UPDATE danber_meta.migracion SET estado='aplicada', aplicada_en=now() WHERE version=3 AND estado='aplicando';

-- 004_pacientes.sql

INSERT INTO danber_meta.migracion(version,nombre,sha256,estado) VALUES (4,'004_pacientes.sql','84b29f55142375d39ce7f5fc8fb45746d1c034f9742614f30aa917d09e9eb7ca','aplicando');

CREATE TABLE pacientes.tipo_documento (
  codigo STRING(30) PRIMARY KEY CHECK (codigo ~ '^[A-Z][A-Z0-9_]{0,29}$'),
  nombre STRING(120) NOT NULL CHECK (length(btrim(nombre)) > 0),
  activo BOOL NOT NULL DEFAULT true
);
CREATE TABLE pacientes.paciente (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  nombres STRING(150) NOT NULL CHECK (length(btrim(nombres)) > 0),
  apellidos STRING(150) NOT NULL CHECK (length(btrim(apellidos)) > 0),
  fecha_nacimiento DATE NOT NULL,
  estado_civil STRING(30) NULL CHECK (estado_civil IS NULL OR length(btrim(estado_civil)) > 0),
  ocupacion STRING(150) NULL CHECK (ocupacion IS NULL OR length(btrim(ocupacion)) > 0),
  direccion STRING(500) NOT NULL CHECK (length(btrim(direccion)) > 0),
  telefono STRING(30) NOT NULL CHECK (length(btrim(telefono)) > 0),
  correo STRING(254) NULL CHECK (correo IS NULL OR length(btrim(correo)) > 0),
  estado STRING(20) NOT NULL DEFAULT 'activo' CHECK (estado IN ('activo','inactivo')),
  version INT8 NOT NULL DEFAULT 1 CHECK (version > 0),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  creado_por UUID NULL REFERENCES identidad.usuario(id),
  actualizado_por UUID NULL REFERENCES identidad.usuario(id)
);
CREATE TABLE pacientes.documento_paciente (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  paciente_id UUID NOT NULL REFERENCES pacientes.paciente(id),
  tipo STRING(30) NOT NULL REFERENCES pacientes.tipo_documento(codigo),
  numero STRING(40) NOT NULL CHECK (length(btrim(numero)) > 0),
  numero_normalizado STRING(40) AS (upper(btrim(numero))) STORED,
  principal BOOL NOT NULL DEFAULT true,
  CHECK (tipo <> 'DNI' OR numero ~ '^[0-9]{8}$'),
  UNIQUE (tipo, numero_normalizado),
  version INT8 NOT NULL DEFAULT 1 CHECK (version > 0),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  creado_por UUID NULL REFERENCES identidad.usuario(id),
  actualizado_por UUID NULL REFERENCES identidad.usuario(id)
);
CREATE UNIQUE INDEX documento_principal_unico ON pacientes.documento_paciente (paciente_id) WHERE principal;
CREATE INDEX documento_por_paciente ON pacientes.documento_paciente(paciente_id);
CREATE TABLE pacientes.apoderado (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  nombres STRING(150) NOT NULL CHECK (length(btrim(nombres)) > 0),
  apellidos STRING(150) NOT NULL CHECK (length(btrim(apellidos)) > 0),
  tipo_documento STRING(30) NOT NULL REFERENCES pacientes.tipo_documento(codigo),
  numero STRING(40) NOT NULL CHECK (length(btrim(numero)) > 0),
  numero_normalizado STRING(40) AS (upper(btrim(numero))) STORED,
  CHECK (tipo_documento <> 'DNI' OR numero ~ '^[0-9]{8}$'),
  UNIQUE (tipo_documento, numero_normalizado),
  version INT8 NOT NULL DEFAULT 1 CHECK (version > 0),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  creado_por UUID NULL REFERENCES identidad.usuario(id),
  actualizado_por UUID NULL REFERENCES identidad.usuario(id)
);
CREATE TABLE pacientes.paciente_apoderado (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  paciente_id UUID NOT NULL REFERENCES pacientes.paciente(id),
  apoderado_id UUID NOT NULL REFERENCES pacientes.apoderado(id),
  parentesco STRING(80) NOT NULL CHECK (length(btrim(parentesco)) > 0),
  desde DATE NOT NULL,
  hasta DATE NULL,
  principal BOOL NOT NULL DEFAULT true,
  CHECK (hasta IS NULL OR hasta > desde),
  UNIQUE(paciente_id, apoderado_id, desde),
  version INT8 NOT NULL DEFAULT 1 CHECK (version > 0),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  creado_por UUID NULL REFERENCES identidad.usuario(id),
  actualizado_por UUID NULL REFERENCES identidad.usuario(id)
);
CREATE UNIQUE INDEX apoderado_principal_abierto ON pacientes.paciente_apoderado(paciente_id) WHERE principal AND hasta IS NULL;
CREATE TABLE pacientes.cambio_paciente (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  paciente_id UUID NOT NULL REFERENCES pacientes.paciente(id),
  actor_id UUID NOT NULL REFERENCES identidad.usuario(id),
  fecha TIMESTAMPTZ NOT NULL DEFAULT now(),
  operacion STRING(100) NOT NULL CHECK (length(btrim(operacion)) > 0),
  motivo STRING(300) NOT NULL CHECK (length(btrim(motivo)) > 0),
  antes JSONB NULL,
  despues JSONB NOT NULL
);
CREATE INDEX cambio_por_paciente ON pacientes.cambio_paciente(paciente_id,fecha);
CREATE INDEX paciente_por_nombre ON pacientes.paciente(apellidos,nombres);


UPDATE danber_meta.migracion SET estado='aplicada', aplicada_en=now() WHERE version=4 AND estado='aplicando';

-- 005_agenda.sql

INSERT INTO danber_meta.migracion(version,nombre,sha256,estado) VALUES (5,'005_agenda.sql','ef82c8b65ef88fb52388e6f9ee0b6cfe7f61a0c38b43584cab9fffca19e46840','aplicando');

CREATE TABLE agenda.control_profesional (
  profesional_id UUID PRIMARY KEY REFERENCES personal.profesional(id),
  revision INT8 NOT NULL DEFAULT 0 CHECK (revision >= 0)
);
CREATE TABLE agenda.horario (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  profesional_id UUID NOT NULL REFERENCES personal.profesional(id),
  dia_semana INT8 NOT NULL CHECK (dia_semana BETWEEN 1 AND 7),
  inicio_minuto INT8 NOT NULL CHECK (inicio_minuto >= 0),
  fin_minuto INT8 NOT NULL CHECK (fin_minuto <= 1440 AND fin_minuto > inicio_minuto),
  desde DATE NOT NULL,
  hasta DATE NULL CHECK (hasta IS NULL OR hasta > desde),
  zona STRING(64) NOT NULL DEFAULT 'America/Lima' CHECK (zona = 'America/Lima'),
  version INT8 NOT NULL DEFAULT 1 CHECK (version > 0),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  creado_por UUID NULL REFERENCES identidad.usuario(id),
  actualizado_por UUID NULL REFERENCES identidad.usuario(id)
);
CREATE TABLE agenda.excepcion_horario (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  profesional_id UUID NOT NULL REFERENCES personal.profesional(id),
  fecha DATE NOT NULL,
  motivo STRING(300) NOT NULL CHECK (length(btrim(motivo)) > 0),
  zona STRING(64) NOT NULL DEFAULT 'America/Lima' CHECK (zona = 'America/Lima'),
  UNIQUE(profesional_id,fecha),
  version INT8 NOT NULL DEFAULT 1 CHECK (version > 0),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  creado_por UUID NULL REFERENCES identidad.usuario(id),
  actualizado_por UUID NULL REFERENCES identidad.usuario(id)
);
CREATE TABLE agenda.tramo_excepcion (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  excepcion_id UUID NOT NULL REFERENCES agenda.excepcion_horario(id),
  inicio_minuto INT8 NOT NULL CHECK (inicio_minuto >= 0),
  fin_minuto INT8 NOT NULL CHECK (fin_minuto <= 1440 AND fin_minuto > inicio_minuto)
);
CREATE TABLE agenda.bloqueo (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  profesional_id UUID NOT NULL REFERENCES personal.profesional(id),
  inicio TIMESTAMPTZ NOT NULL,
  fin TIMESTAMPTZ NOT NULL,
  motivo STRING(300) NOT NULL CHECK (length(btrim(motivo)) > 0),
  estado STRING(20) NOT NULL DEFAULT 'activo' CHECK (estado IN ('activo','retirado')),
  CHECK (fin > inicio),
  version INT8 NOT NULL DEFAULT 1 CHECK (version > 0),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  creado_por UUID NULL REFERENCES identidad.usuario(id),
  actualizado_por UUID NULL REFERENCES identidad.usuario(id)
);
CREATE TABLE agenda.cita (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  paciente_id UUID NOT NULL REFERENCES pacientes.paciente(id),
  profesional_id UUID NOT NULL REFERENCES personal.profesional(id),
  inicio TIMESTAMPTZ NOT NULL,
  fin TIMESTAMPTZ NOT NULL,
  zona STRING(64) NOT NULL DEFAULT 'America/Lima' CHECK (zona = 'America/Lima'),
  origen STRING(20) NOT NULL DEFAULT 'programada' CHECK (origen IN ('programada','sin_cita')),
  resultado STRING(20) NOT NULL DEFAULT 'pendiente' CHECK (resultado IN ('pendiente','atendida','cancelada','no_asistio')),
  ocupa_intervalo BOOL NOT NULL DEFAULT true,
  llegada_en TIMESTAMPTZ NULL,
  inicio_real TIMESTAMPTZ NULL,
  fin_real TIMESTAMPTZ NULL,
  CHECK (fin >= inicio + INTERVAL '30 minutes'),
  CHECK (fin_real IS NULL OR (inicio_real IS NOT NULL AND fin_real >= inicio_real)),
  CHECK (inicio_real IS NULL OR (llegada_en IS NOT NULL AND inicio_real >= llegada_en)),
  CHECK (resultado <> 'atendida' OR fin_real IS NOT NULL),
  CHECK (resultado NOT IN ('cancelada','no_asistio') OR (NOT ocupa_intervalo AND inicio_real IS NULL)),
  CHECK (resultado <> 'no_asistio' OR llegada_en IS NULL),
  UNIQUE(id,paciente_id,profesional_id),
  version INT8 NOT NULL DEFAULT 1 CHECK (version > 0),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  creado_por UUID NULL REFERENCES identidad.usuario(id),
  actualizado_por UUID NULL REFERENCES identidad.usuario(id)
);
CREATE UNIQUE INDEX una_atencion_abierta ON agenda.cita(profesional_id) WHERE inicio_real IS NOT NULL AND fin_real IS NULL;
CREATE INDEX citas_por_profesional ON agenda.cita(profesional_id,inicio) STORING(fin,ocupa_intervalo,resultado);
CREATE INDEX citas_por_paciente ON agenda.cita(paciente_id,inicio);
CREATE INDEX bloqueos_por_profesional ON agenda.bloqueo(profesional_id,inicio) STORING(fin,estado);
CREATE INDEX horario_por_profesional ON agenda.horario(profesional_id,dia_semana,desde);
CREATE TABLE agenda.espera (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  paciente_id UUID NOT NULL REFERENCES pacientes.paciente(id),
  profesional_id UUID NOT NULL REFERENCES personal.profesional(id),
  cita_origen_id UUID NULL,
  cita_asignada_id UUID NULL,
  llegada_en TIMESTAMPTZ NOT NULL,
  motivo STRING(20) NOT NULL CHECK (motivo IN ('tardanza','sin_cita')),
  estado STRING(20) NOT NULL DEFAULT 'esperando' CHECK (estado IN ('esperando','asignada','retirada','cerrada')),
  salida_en TIMESTAMPTZ NULL,
  FOREIGN KEY (cita_origen_id,paciente_id,profesional_id) REFERENCES agenda.cita(id,paciente_id,profesional_id),
  FOREIGN KEY (cita_asignada_id,paciente_id,profesional_id) REFERENCES agenda.cita(id,paciente_id,profesional_id),
  CHECK (motivo <> 'tardanza' OR cita_origen_id IS NOT NULL),
  CHECK (estado <> 'asignada' OR cita_asignada_id IS NOT NULL),
  CHECK ((estado = 'esperando' AND salida_en IS NULL AND cita_asignada_id IS NULL) OR (estado <> 'esperando' AND salida_en IS NOT NULL)),
  CHECK (salida_en IS NULL OR salida_en >= llegada_en),
  version INT8 NOT NULL DEFAULT 1 CHECK (version > 0),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  creado_por UUID NULL REFERENCES identidad.usuario(id),
  actualizado_por UUID NULL REFERENCES identidad.usuario(id)
);
CREATE UNIQUE INDEX espera_activa_unica ON agenda.espera(paciente_id,profesional_id) WHERE estado = 'esperando';
CREATE INDEX espera_por_llegada ON agenda.espera(profesional_id,estado,llegada_en);
CREATE TABLE agenda.cambio_cita (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cita_id UUID NOT NULL REFERENCES agenda.cita(id),
  actor_id UUID NOT NULL REFERENCES identidad.usuario(id),
  fecha TIMESTAMPTZ NOT NULL DEFAULT now(),
  operacion STRING(100) NOT NULL CHECK (length(btrim(operacion)) > 0),
  motivo STRING(300) NOT NULL CHECK (length(btrim(motivo)) > 0),
  antes JSONB NULL,
  despues JSONB NOT NULL
);
CREATE INDEX cambio_por_cita ON agenda.cambio_cita(cita_id,fecha);
CREATE TABLE agenda.cambio_disponibilidad (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  profesional_id UUID NOT NULL REFERENCES personal.profesional(id),
  actor_id UUID NOT NULL REFERENCES identidad.usuario(id),
  fecha TIMESTAMPTZ NOT NULL DEFAULT now(),
  operacion STRING(100) NOT NULL CHECK (length(btrim(operacion)) > 0),
  motivo STRING(300) NOT NULL CHECK (length(btrim(motivo)) > 0),
  antes JSONB NULL,
  despues JSONB NOT NULL
);
CREATE TABLE agenda.solicitud_idempotente (
  actor_id UUID NOT NULL REFERENCES identidad.usuario(id),
  operacion STRING(100) NOT NULL,
  clave STRING(100) NOT NULL CHECK (length(btrim(clave)) > 0),
  huella STRING(64) NOT NULL CHECK (huella ~ '^[0-9a-f]{64}$'),
  cita_id UUID NOT NULL REFERENCES agenda.cita(id),
  resultado JSONB NOT NULL,
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (actor_id,operacion,clave)
);


UPDATE danber_meta.migracion SET estado='aplicada', aplicada_en=now() WHERE version=5 AND estado='aplicando';

-- 006_catalogos_y_permisos.sql

INSERT INTO danber_meta.migracion(version,nombre,sha256,estado) VALUES (6,'006_catalogos_y_permisos.sql','f17fc2dd1df545f396a8aa304a4ce7d8b7d8aed5079799cbe9cb50fecc19c612','aplicando');

-- Solo catálogos y permisos; no usuarios, contraseñas, pacientes ni profesionales.
BEGIN;
INSERT INTO pacientes.tipo_documento(codigo,nombre) VALUES ('DNI','Documento Nacional de Identidad');
INSERT INTO identidad.rol(codigo,nombre) VALUES ('maestro','Maestro'),('recepcion','Recepción'),('odontologo','Odontólogo');
INSERT INTO identidad.permiso(codigo,descripcion) VALUES
('pacientes.leer','Consultar identificación administrativa'),
('pacientes.crear','Registrar paciente y apoderado'),
('pacientes.editar','Modificar identificación con historial'),
('agenda.leer','Consultar agenda autorizada'),
('agenda.reservar','Crear cita'),
('agenda.duracion','Ajustar duración'),
('agenda.llegada','Registrar llegada'),
('agenda.cancelar','Cancelar cita'),
('agenda.reprogramar','Reprogramar cita'),
('agenda.espera','Gestionar espera y liberar hueco'),
('agenda.inasistencia','Registrar o rectificar inasistencia'),
('agenda.horario','Definir horario propio'),
('agenda.bloqueo','Crear o retirar bloqueo propio'),
('agenda.atencion','Iniciar/finalizar atención administrativa propia'),
('identidad.administrar','Administrar cuentas y roles');
INSERT INTO identidad.rol_permiso(rol_id,permiso_id) SELECT r.id,p.id FROM identidad.rol r CROSS JOIN identidad.permiso p WHERE r.codigo='maestro';
INSERT INTO identidad.rol_permiso(rol_id,permiso_id) SELECT r.id,p.id FROM identidad.rol r CROSS JOIN identidad.permiso p WHERE r.codigo='recepcion' AND p.codigo IN ('pacientes.leer','pacientes.crear','pacientes.editar','agenda.leer','agenda.reservar','agenda.duracion','agenda.llegada','agenda.cancelar','agenda.reprogramar','agenda.espera','agenda.inasistencia');
INSERT INTO identidad.rol_permiso(rol_id,permiso_id) SELECT r.id,p.id FROM identidad.rol r CROSS JOIN identidad.permiso p WHERE r.codigo='odontologo' AND p.codigo IN ('pacientes.leer','pacientes.crear','pacientes.editar','agenda.leer','agenda.horario','agenda.bloqueo','agenda.atencion');
COMMIT;
-- Permisos adicionales a los confirmados: propuesta ADR 0007, revisión pendiente.


UPDATE danber_meta.migracion SET estado='aplicada', aplicada_en=now() WHERE version=6 AND estado='aplicando';

-- 007_privilegios_runtime.sql

INSERT INTO danber_meta.migracion(version,nombre,sha256,estado) VALUES (7,'007_privilegios_runtime.sql','3d98602dd59ff15cc2527517cd4d02e2fc5dc9f00381a4e59f92477067be70a1','aplicando');

-- Rol SQL del backend. No es el rol Maestro de la aplicación.
-- Ejecutar como propietario/migrador; no concede CREATE ni acceso a danber_meta.
CREATE ROLE IF NOT EXISTS danber_runtime;
GRANT USAGE ON SCHEMA identidad, personal, pacientes, agenda TO danber_runtime;
GRANT SELECT, INSERT, UPDATE ON TABLE
 identidad.usuario, identidad.credencial, identidad.sesion,
 personal.personal, personal.profesional, personal.especialidad,
 pacientes.paciente, pacientes.documento_paciente, pacientes.apoderado, pacientes.paciente_apoderado,
 agenda.control_profesional, agenda.horario, agenda.excepcion_horario, agenda.tramo_excepcion,
 agenda.bloqueo, agenda.cita, agenda.espera
TO danber_runtime;
GRANT SELECT ON TABLE identidad.rol, identidad.permiso, identidad.rol_permiso TO danber_runtime;
GRANT SELECT, INSERT, DELETE ON TABLE identidad.usuario_rol, personal.profesional_especialidad TO danber_runtime;
GRANT SELECT, INSERT ON TABLE identidad.cambio_acceso, pacientes.cambio_paciente,
 agenda.cambio_cita, agenda.cambio_disponibilidad, agenda.solicitud_idempotente
TO danber_runtime;
GRANT SELECT ON TABLE pacientes.tipo_documento TO danber_runtime;
-- Los catálogos y permisos se amplían mediante migrador/control administrativo.
-- Usuario LOGIN y contraseña/certificados se provisionan fuera del repositorio.


UPDATE danber_meta.migracion SET estado='aplicada', aplicada_en=now() WHERE version=7 AND estado='aplicando';

-- 008_restringir_ddl_public.sql

INSERT INTO danber_meta.migracion(version,nombre,sha256,estado) VALUES (8,'008_restringir_ddl_public.sql','674461041983d7b9ed71895b24b8eb41af4dc7a74264a2ea7897d7587f4144e3','aplicando');

-- Backend inicial: el rol public no debe permitir DDL al usuario de ejecución.
-- No modifica tablas ni datos; propietario y administrador conservan sus permisos.
REVOKE CREATE ON SCHEMA public FROM public;


UPDATE danber_meta.migracion SET estado='aplicada', aplicada_en=now() WHERE version=8 AND estado='aplicando';

-- 009_aprovisionamiento_inicial.sql

INSERT INTO danber_meta.migracion(version,nombre,sha256,estado) VALUES (9,'009_aprovisionamiento_inicial.sql','46b9dbeac986b9acd83188940fabd4ede72beefb6ec34034fd54f4ce48dc6a61','aplicando');

-- One existing row serializes the local, one-time bootstrap command.
CREATE TABLE identidad.aprovisionamiento_inicial (
  clave STRING(50) PRIMARY KEY CHECK (clave = 'maestro_inicial'),
  usuario_id UUID NULL UNIQUE REFERENCES identidad.usuario(id),
  completado_en TIMESTAMPTZ NULL,
  revision INT8 NOT NULL DEFAULT 0 CHECK (revision >= 0),
  CHECK ((usuario_id IS NULL AND completado_en IS NULL)
      OR (usuario_id IS NOT NULL AND completado_en IS NOT NULL))
);
INSERT INTO identidad.aprovisionamiento_inicial (clave) VALUES ('maestro_inicial');
GRANT SELECT, UPDATE ON TABLE identidad.aprovisionamiento_inicial TO danber_runtime;


UPDATE danber_meta.migracion SET estado='aplicada', aplicada_en=now() WHERE version=9 AND estado='aplicando';

-- 010_sesiones_y_version_acceso.sql

INSERT INTO danber_meta.migracion(version,nombre,sha256,estado) VALUES (10,'010_sesiones_y_version_acceso.sql','21499c1e86a55f8f56cc0191506f645fbee6291576b5a198dd074cc8c4720abd','aplicando');

-- Existing sessions gain server-side inactivity and account-version checks.
ALTER TABLE identidad.sesion ADD COLUMN ultimo_uso_en TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE identidad.sesion ADD COLUMN usuario_version INT8 NOT NULL DEFAULT 1 CHECK (usuario_version > 0);
CREATE INDEX sesion_usuario_activa ON identidad.sesion (usuario_id) WHERE revocado_en IS NULL;


UPDATE danber_meta.migracion SET estado='aplicada', aplicada_en=now() WHERE version=10 AND estado='aplicando';

-- 011_registro_paciente_idempotente.sql

INSERT INTO danber_meta.migracion(version,nombre,sha256,estado) VALUES (11,'011_registro_paciente_idempotente.sql','9a6d567e027f996d06d8f64b989fa94c13b1676ff9e69ba3c0cd215362946629','aplicando');

-- Reintentos del alta: misma cuenta + clave conserva la misma respuesta.
-- La huella es SHA-256 del comando normalizado; no almacena el cuerpo HTTP.
CREATE TABLE pacientes.registro_idempotente (
  actor_id UUID NOT NULL REFERENCES identidad.usuario(id),
  clave UUID NOT NULL,
  huella BYTES NOT NULL CHECK (length(huella) = 32),
  paciente_id UUID NULL REFERENCES pacientes.paciente(id),
  creado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (actor_id, clave)
);
-- El caso de uso completa paciente_id en la misma transacción que reclama la clave.
GRANT SELECT, INSERT, UPDATE ON TABLE pacientes.registro_idempotente TO danber_runtime;


UPDATE danber_meta.migracion SET estado='aplicada', aplicada_en=now() WHERE version=11 AND estado='aplicando';

SELECT version,nombre,sha256,estado FROM danber_meta.migracion ORDER BY version;
