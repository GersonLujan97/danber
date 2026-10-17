# ADR 0004 — duración, roles y disponibilidad del odontólogo

**Complementado el 2026-10-16 por [ADR 0005](0005-identificacion-y-permisos.md):** odontólogo autorizado a gestionar sus horarios y crear/retirar sus bloqueos; odontólogo actual Maestro y Recepción. El flujo de citas afectadas sigue pendiente.


Fecha: 2026-10-16. Origen: respuesta explícita del usuario sobre operación del consultorio. Estado: **Confirmado** para reglas operativas; **Propuesto** para el manejo detallado de bloqueos. Revisión de impacto: **Pendiente de revisión**. Bitácora: CHG-004 en Notion. Complementa ADR 0003; conserva sus reglas de identificación.

## Confirmado

- Recepción puede modificar la duración. Una misma persona puede tener roles de odontólogo y recepción.
- Duración predeterminada y mínima: **30 minutos**. Puede ampliarse según criterio; no se ha acordado un máximo ni incrementos fijos. La propuesta anterior de 15 a 120 minutos en pasos de 15 no fue aceptada.
- Domingo a viernes: **08:00–12:30 y 15:00–19:00**, hora de Lima. El almuerzo queda fuera del horario reservable.
- Sábado: inicio **18:00**, cierre a **20:00 o 21:00**. Debe concretarse el cierre aplicable a cada fecha; no asumir 21:00 para todos los sábados.
- El odontólogo puede comunicar ausencias planificadas por parte de un día o por varios días. Se mantiene la prohibición de atender dos pacientes simultáneamente.

## Propuesta: horario habitual y bloqueos por fecha

Un horario habitual describe cuándo suele atender el profesional. Un bloqueo es una excepción con inicio y fin que impide nuevas reservas para ese profesional. La disponibilidad se obtiene restando al horario los bloqueos y las citas que ocupan tiempo.

Ejemplos: ausencia del lunes por la mañana → bloqueo del lunes 08:00–12:30; viaje por tres días → bloqueo desde el inicio del primer día hasta el inicio del cuarto día, usando fechas concretas y zona America/Lima. El horario semanal permanece intacto.

Para un almuerzo recurrente, usar dos tramos de horario; no crear manualmente un bloqueo cada día. Si un sábado se extiende hasta las 21:00, registrar el horario específico de esa fecha o una ampliación explícita. Un bloqueo solo reduce disponibilidad, no amplía atención.

Toda cita debe caber completa en un tramo disponible. Ejemplo: una cita de 12:15 a 12:45 no cabe antes del almuerzo aunque empiece dentro del horario.

## Propuesta: citas existentes dentro de una ausencia

1. Mostrar al crear el bloqueo la lista de citas afectadas y el intervalo.
2. Tras confirmar el bloqueo, impedir nuevas reservas en ese intervalo, conservar las citas existentes y señalarlas como conflictos pendientes de gestión. Esta marca no implica cancelación ni cambio de estado clínico.
3. Recepción gestiona cada cita: contactar al paciente y reprogramar o cancelar de forma explícita. Con varios profesionales, una reasignación exige disponibilidad y elegibilidad del nuevo odontólogo.
4. Mantener visibles los conflictos hasta resolverlos, con responsable y trazabilidad. No borrar ni mover citas automáticamente.
5. Al retirar el bloqueo, recalcular conflictos y disponibilidad; no restaurar automáticamente citas canceladas.

Confirmación del bloqueo y reservas competidoras deben usar el mismo protocolo de concurrencia del ADR 0002. Una reserva que haya entrado antes del bloqueo debe aparecer entre las afectadas; una posterior no puede ocuparlo. SQL aún no implementado ni probado.

## Roles y propiedad de datos

Propuesto: una cuenta personal con varias asignaciones de rol; no duplicar cuentas ni profesionales por desempeñar recepción. El permiso de ajustar duración procede de recepción, aunque esa persona también sea odontólogo. No conceder automáticamente ese permiso a todo odontólogo por su profesión.

Identidad y Acceso administra roles/permisos; Personal administra el perfil profesional; Agenda administra horario, bloqueos, citas y cambios. Toda modificación conserva actor, fecha, valores y motivo cuando corresponda.

## Alternativas y consecuencias

Cambiar la semana habitual por una ausencia puntual puede cerrar futuros lunes por error y perder la explicación del viaje. Los bloqueos fechados conservan el patrón y la excepción.

Cancelar automáticamente todas las citas elimina la gestión individual y puede dejar pacientes sin aviso. Rechazar cualquier bloqueo que tenga citas permitiría seguir reservando durante una ausencia real. Se propone bloquear nuevas reservas y gestionar explícitamente las citas afectadas.

El costo es incorporar una lista visible de conflictos, historial y pruebas de concurrencia; el beneficio es no prometer horarios en los que el odontólogo estará ausente.

## Impacto y pendientes

Revisar portada y capítulos 01, 03 y 07; diccionario de horario/bloqueo/cita, contratos para ajustar duración y registrar ausencias, permisos, auditoría y futuros escenarios de prueba. No hay código ni migraciones ejecutables modificados.

Pendiente de validación: cierre de cada sábado, quién puede crear/retirar bloqueos y cambiar horarios, y aceptación del flujo propuesto para citas existentes. No inventar un máximo de duración; toda ampliación sigue limitada por horario, bloqueos y otras citas.

Criterios para futuras pruebas: mínimo de 30; no cruzar almuerzo/cierre; bloqueo parcial y de varios días; reserva concurrente con bloqueo; conservar citas afectadas; usuario con ambos roles; aplicar exclusión por profesional, no global. Son escenarios documentados, no pruebas ejecutadas.


