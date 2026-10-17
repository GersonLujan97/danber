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
