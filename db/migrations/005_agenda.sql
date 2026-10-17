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
