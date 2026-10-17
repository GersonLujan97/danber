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
