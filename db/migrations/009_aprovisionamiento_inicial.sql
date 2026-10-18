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
