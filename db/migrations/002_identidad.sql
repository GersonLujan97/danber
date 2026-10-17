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
