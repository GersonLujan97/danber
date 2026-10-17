# ADR 0006 — tardanzas, espera y pacientes sin cita

Fecha: 2026-10-16. Origen: descripción del usuario y solicitud de proponer un proceso ordenado. **Confirmado:** comportamiento del consultorio descrito abajo. **Propuesto:** mecanismo operativo y de datos. Revisión de impacto: **Pendiente de revisión**. CHG-006; complementa ADR 0005 en reasignación.

## Reglas confirmadas

- Si el paciente con cita no llega, puede atenderse a alguien sin cita.
- Quien llega tarde no desplaza a otro paciente con cita; espera un espacio disponible o solicita reprogramar a otro día.
- Se mantiene una sola atención a la vez por odontólogo, duración mínima de reserva de 30 minutos y respeto de horarios/bloqueos.
- En este escenario, reasignar significa reprogramar la cita; no implica cambio automático de odontólogo.

## Proceso recomendado

1. Recepción registra la llegada real de cada paciente y distingue: puntual con cita, tardío con cita y sin cita.
2. La llegada tardía y la llegada sin cita pueden incorporarse a una lista de espera del día. La espera no reserva una hora ni garantiza atención ese día.
3. Las citas vigentes tienen prioridad. Entre personas en espera aptas para un mismo hueco, proponer orden de llegada; no desplazar automáticamente una cita posterior.
4. Antes de asignar el hueco, comprobar que cabe la duración prevista completa (mínimo 30 minutos), que el odontólogo está libre y que no cruza próxima cita, almuerzo, bloqueo ni cierre.
5. Si el hueco procede de una cita cuyo paciente no llegó, recepción debe liberar expresamente su intervalo con motivo, actor y hora. No basta con observar ausencia ni se permite una segunda reserva superpuesta. Conservar la cita y su horario original en el historial.
6. Liberación y ocupación del intervalo se confirman de forma atómica bajo el protocolo de concurrencia del ADR 0002; una reserva competidora no puede ocupar el mismo hueco. Si falla la operación, no dejar una sustitución a medias.
7. Cuando se asigne un espacio al paciente tardío, reprogramar su cita conservando historial y llegada real. Para quien vino sin cita, crear registro administrativo del espacio asignado con origen «sin cita», sin inventar una reserva previa. La atención clínica sigue siendo un registro distinto.
8. Si el paciente prefiere otra fecha, recepción confirma una nueva disponibilidad y reprograma con motivo y consentimiento operativo del paciente. Retirar su entrada de espera activa; no duplicar cita.
9. Al terminar la jornada, resolver las esperas restantes como retirado/no atendido/reprogramado según lo ocurrido; no trasladarlas automáticamente al día siguiente.

## Distinguir hechos y estados

«No llegó a la hora» es una observación temporal, no prueba de inasistencia definitiva. Mostrar «Pendiente de llegada» cuando pasó el inicio y aún no hay llegada; no cancelar automáticamente.

«Llegó tarde» compara llegada real con hora original. «En espera» expresa presencia esperando un hueco. «No asistió» debe reservarse para una cita que terminó sin llegada, según revisión operativa, y puede necesitar rectificación trazable si luego llega. Mantener separados llegada, ocupación del intervalo y resultado de cita.

Propuesta: recepción registra llegada, espera, liberación y reprogramación; odontólogo inicia/finaliza la atención. No liberar una cita ya en atención como si fuera ausencia. El usuario Maestro también respeta estas reglas.

No se fija tolerancia automática de 5, 10 o 15 minutos. La liberación es explícita y visible. Si se adopta una tolerancia después, se registrará como nueva decisión.

## Ejemplos ficticios para validar

- A tiene 09:00–09:30; no llegó. B está sin cita a las 09:00; C tiene 09:30. Si se confirma la liberación de A y B requiere 30 minutos, B puede ocupar 09:00–09:30.
- Si B llega a las 09:10, solo quedan 20 minutos antes de C: no asignar una reserva de 30 ni ocupar 09:10–09:40. Debe esperar otro hueco.
- A llega a las 09:15 mientras B está en atención: se registra su llegada tardía. No interrumpe B ni desplaza C; espera un hueco completo o pide otro día.
- Si una atención se prolonga realmente, registrar la demora y gestionar las citas afectadas explícitamente. La duración real no se falsifica ni provoca desplazamientos automáticos.

## Datos y consecuencias

Proponer entrada de espera con paciente, fecha de llegada, odontólogo solicitado, cita de origen opcional, motivo (tardanza/sin cita), estado y salida. Relacionar entrada con la cita asignada. Evitar dos entradas activas para la misma intención de atención; revalidar al asignar. No convertirla en una cola técnica o un microservicio.

Registrar también liberación de intervalo, motivo, actor y hora; preservar horario reservado original, cambios y llegada. Espera no cuenta como ocupación de agenda. Sin cita no exime de identificación y apoderado conforme al diccionario vigente.

Alternativa de atender en orden puro de llegada: perjudica reservas puntuales. Cancelación automática al minuto de inicio: confunde tardanza con ausencia definitiva. Reservas superpuestas para reemplazos: contradicen exclusión. La propuesta conserva prioridades y trazabilidad a cambio de una gestión explícita de espera/liberación.

## Impacto y revisión pendiente

Actualizar capítulo 03 (diccionario/recorrido), capítulo 07 (ADR/bitácora), portada, README/AGENTS y referencias locales. Revisar contratos de liberar/asignar/reprogramar, permisos, estados, auditoría, esquema de espera y pruebas de concurrencia. No hay código ni migraciones ejecutables modificados.

Confirmado: esperar sin desplazar citas o reprogramar. Propuesto: orden de espera, liberación explícita, responsables y estados detallados. Responsable de revisión funcional: usuario; validación técnica al implementar. Escenarios descritos, no pruebas ejecutadas.

