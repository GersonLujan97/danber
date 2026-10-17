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
