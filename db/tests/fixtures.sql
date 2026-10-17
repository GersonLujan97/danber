-- Datos completamente ficticios; solo se ejecutan dentro de db/tests/verify.py.
INSERT INTO identidad.usuario(id,nombre_acceso) VALUES ('00000000-0000-4000-8000-000000000001','verificacion_sintetica');
INSERT INTO personal.personal(id,nombres,apellidos) VALUES
('10000000-0000-4000-8000-000000000001','Profesional','Prueba Uno'),
('10000000-0000-4000-8000-000000000002','Profesional','Prueba Dos');
INSERT INTO personal.profesional(id,personal_id,dni,estado) VALUES
('20000000-0000-4000-8000-000000000001','10000000-0000-4000-8000-000000000001','00000001','habilitado_agenda'),
('20000000-0000-4000-8000-000000000002','10000000-0000-4000-8000-000000000002','00000002','habilitado_agenda');
INSERT INTO agenda.control_profesional(profesional_id)
SELECT id FROM personal.profesional;
INSERT INTO personal.especialidad(id,nombre) VALUES ('30000000-0000-4000-8000-000000000001','Especialidad ficticia de prueba');
INSERT INTO personal.profesional_especialidad SELECT id,'30000000-0000-4000-8000-000000000001' FROM personal.profesional;
INSERT INTO pacientes.paciente(id,nombres,apellidos,fecha_nacimiento,direccion,telefono) VALUES
('40000000-0000-4000-8000-000000000001','Paciente','Ficticio Uno','1990-01-01','Direccion ficticia','000000001'),
('40000000-0000-4000-8000-000000000002','Paciente','Ficticio Dos','1991-01-01','Direccion ficticia','000000002');
INSERT INTO pacientes.documento_paciente(paciente_id,tipo,numero) VALUES
('40000000-0000-4000-8000-000000000001','DNI','00000101'),
('40000000-0000-4000-8000-000000000002','DNI','00000102');
INSERT INTO agenda.horario(profesional_id,dia_semana,inicio_minuto,fin_minuto,desde)
SELECT p.id,d.dia,t.inicio,t.fin,'2026-01-01'::DATE FROM personal.profesional p
CROSS JOIN (VALUES(1),(2),(3),(4),(5),(7)) AS d(dia)
CROSS JOIN (VALUES(480,750),(900,1140)) AS t(inicio,fin);
