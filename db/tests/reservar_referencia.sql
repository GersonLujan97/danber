-- Referencia transaccional de RESERVA: no es un procedimiento ni una API.
-- Sustituir solo mediante parámetros de driver en el backend; tests usan datos internos.
-- Reintentar la transacción COMPLETA ante SQLSTATE 40001, con la misma clave/huella/ID.
BEGIN ISOLATION LEVEL SERIALIZABLE;
UPDATE agenda.control_profesional SET revision=revision+1 WHERE profesional_id='$profesional';
INSERT INTO agenda.cita(id,paciente_id,profesional_id,inicio,fin,creado_por,actualizado_por)
SELECT '$cita',p.id,pr.id,'$inicio'::TIMESTAMPTZ,'$fin'::TIMESTAMPTZ,'$actor','$actor'
FROM pacientes.paciente p
JOIN personal.profesional pr ON pr.id='$profesional' AND pr.estado='habilitado_agenda'
JOIN personal.personal pe ON pe.id=pr.personal_id AND pe.estado='activo'
JOIN agenda.control_profesional cp ON cp.profesional_id=pr.id
WHERE p.id='$paciente' AND p.estado='activo'
AND '$fin'::TIMESTAMPTZ >= '$inicio'::TIMESTAMPTZ + INTERVAL '30 minutes'
AND p.fecha_nacimiento <= (now() AT TIME ZONE 'America/Lima')::DATE
AND EXISTS (SELECT 1 FROM pacientes.documento_paciente d JOIN pacientes.tipo_documento td ON td.codigo=d.tipo AND td.activo
 WHERE d.paciente_id=p.id AND d.principal)
AND (
 (p.fecha_nacimiento <= (now() AT TIME ZONE 'America/Lima')::DATE - INTERVAL '18 years'
  AND p.fecha_nacimiento > (now() AT TIME ZONE 'America/Lima')::DATE - INTERVAL '66 years')
 OR EXISTS (SELECT 1 FROM pacientes.paciente_apoderado pa
  JOIN pacientes.apoderado a ON a.id=pa.apoderado_id
  JOIN pacientes.tipo_documento td ON td.codigo=a.tipo_documento AND td.activo
  WHERE pa.paciente_id=p.id AND pa.desde <= (now() AT TIME ZONE 'America/Lima')::DATE
  AND (pa.hasta IS NULL OR pa.hasta > (now() AT TIME ZONE 'America/Lima')::DATE))
)
AND NOT EXISTS (SELECT 1 FROM agenda.solicitud_idempotente s
 WHERE s.actor_id='$actor' AND s.operacion='reservar' AND s.clave='$clave')
AND ('$inicio'::TIMESTAMPTZ AT TIME ZONE 'America/Lima')::DATE =
    ('$fin'::TIMESTAMPTZ AT TIME ZONE 'America/Lima')::DATE
AND EXISTS (
 SELECT 1 FROM (
  SELECT h.inicio_minuto,h.fin_minuto FROM agenda.horario h
  WHERE h.profesional_id=pr.id
   AND h.dia_semana=extract(isodow FROM '$inicio'::TIMESTAMPTZ AT TIME ZONE 'America/Lima')
   AND h.desde <= ('$inicio'::TIMESTAMPTZ AT TIME ZONE 'America/Lima')::DATE
   AND (h.hasta IS NULL OR h.hasta > ('$inicio'::TIMESTAMPTZ AT TIME ZONE 'America/Lima')::DATE)
   AND NOT EXISTS (SELECT 1 FROM agenda.excepcion_horario e WHERE e.profesional_id=pr.id
    AND e.fecha=('$inicio'::TIMESTAMPTZ AT TIME ZONE 'America/Lima')::DATE)
  UNION ALL
  SELECT te.inicio_minuto,te.fin_minuto FROM agenda.excepcion_horario e
  JOIN agenda.tramo_excepcion te ON te.excepcion_id=e.id
  WHERE e.profesional_id=pr.id AND e.fecha=('$inicio'::TIMESTAMPTZ AT TIME ZONE 'America/Lima')::DATE
 ) disponible
 WHERE disponible.inicio_minuto <=
  extract(hour FROM '$inicio'::TIMESTAMPTZ AT TIME ZONE 'America/Lima')*60+
  extract(minute FROM '$inicio'::TIMESTAMPTZ AT TIME ZONE 'America/Lima')+
  extract(second FROM '$inicio'::TIMESTAMPTZ AT TIME ZONE 'America/Lima')/60
 AND disponible.fin_minuto >=
  extract(hour FROM '$fin'::TIMESTAMPTZ AT TIME ZONE 'America/Lima')*60+
  extract(minute FROM '$fin'::TIMESTAMPTZ AT TIME ZONE 'America/Lima')+
  extract(second FROM '$fin'::TIMESTAMPTZ AT TIME ZONE 'America/Lima')/60
)
AND NOT EXISTS (SELECT 1 FROM agenda.bloqueo b WHERE b.profesional_id=pr.id AND b.estado='activo'
 AND b.inicio < '$fin'::TIMESTAMPTZ AND '$inicio'::TIMESTAMPTZ < b.fin)
AND NOT EXISTS (SELECT 1 FROM agenda.cita c WHERE c.profesional_id=pr.id AND c.ocupa_intervalo
 AND c.inicio < '$fin'::TIMESTAMPTZ AND '$inicio'::TIMESTAMPTZ < c.fin);
INSERT INTO agenda.cambio_cita(cita_id,actor_id,operacion,motivo,despues)
SELECT c.id,'$actor','reservar','Reserva de referencia verificada',jsonb_build_object('inicio',c.inicio,'fin',c.fin)
FROM agenda.cita c WHERE c.id='$cita' AND NOT EXISTS (
 SELECT 1 FROM agenda.solicitud_idempotente s WHERE s.actor_id='$actor' AND s.operacion='reservar' AND s.clave='$clave');
INSERT INTO agenda.solicitud_idempotente(actor_id,operacion,clave,huella,cita_id,resultado)
SELECT '$actor','reservar','$clave','$huella',c.id,jsonb_build_object('cita_id',c.id)
FROM agenda.cita c WHERE c.id='$cita' AND NOT EXISTS (
 SELECT 1 FROM agenda.solicitud_idempotente s WHERE s.actor_id='$actor' AND s.operacion='reservar' AND s.clave='$clave');
COMMIT;
-- Sin fila de recibo compatible: reserva rechazada; no es éxito por retornar código de proceso 0.
SELECT cita_id FROM agenda.solicitud_idempotente
WHERE actor_id='$actor' AND operacion='reservar' AND clave='$clave' AND huella='$huella';
