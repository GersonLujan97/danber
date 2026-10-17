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
