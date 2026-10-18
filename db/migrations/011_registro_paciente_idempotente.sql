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
