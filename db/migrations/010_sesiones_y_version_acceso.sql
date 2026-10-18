-- Existing sessions gain server-side inactivity and account-version checks.
ALTER TABLE identidad.sesion ADD COLUMN ultimo_uso_en TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE identidad.sesion ADD COLUMN usuario_version INT8 NOT NULL DEFAULT 1 CHECK (usuario_version > 0);
CREATE INDEX sesion_usuario_activa ON identidad.sesion (usuario_id) WHERE revocado_en IS NULL;
