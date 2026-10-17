# ADR 0003 — reglas iniciales de citas e identificación

**Sustituido parcialmente el 2026-10-16:** DNI exclusivo y datos mínimos del apoderado reemplazados por [ADR 0005](0005-identificacion-y-permisos.md). El texto inferior conserva la decisión histórica; se mantienen las demás reglas.


Fecha: 2026-10-16. Estado: **Confirmado por el usuario** para las reglas de negocio siguientes. Revisión de impacto técnico: **Pendiente de revisión**. Registro: CHG-003 en [Notion](https://app.notion.com/p/3fb2756ebb44813d8f89fe0371e9edbd).

## Contexto y origen

El usuario especifica citas de 30 minutos con posibilidad de variar el tiempo, un odontólogo actual y preparación para incorporar más, sin atender dos pacientes a la vez. Confirma DNI para todos los pacientes y nombre del padre o apoderado para menores y adultos mayores. En una aclaración posterior precisa el umbral: **a partir de los 66 años**.

## Decisiones confirmadas

- Duración predeterminada: **30 minutos**, con un control para ajustar la duración de la cita.
- Hay **un odontólogo actualmente**; el modelo debe admitir varios en el futuro. Cada cita identifica a su odontólogo.
- Un odontólogo no atiende a dos pacientes simultáneamente: **no permitir reservas que se solapen para el mismo profesional**. No habilitar sobreturnos simultáneos como excepción pendiente.
- Todo paciente se registra con **su propio DNI**, incluidos menores y adultos mayores. No se habilita alta sin DNI dentro del alcance confirmado.
- Para menores de edad y pacientes de **66 años o más**, registrar también el **nombre del padre o apoderado**, separado del nombre y DNI del paciente.
- A los **65 años** no se activa esa condición por edad. No deducir de este dato permisos de acceso, firma o representación legal.
- Se mantiene la reserva únicamente por odontólogo, sin sillón ni consultorio como recurso adicional.

## Consecuencias de diseño propuestas para la fase D

- Pacientes administra DNI, nombres, fecha de nacimiento y vínculo con padre/apoderado. Conservar identidad técnica independiente del DNI y de la numeración clínica.
- Proponer fecha de nacimiento para calcular edad en la fecha de evaluación, sin guardar una edad que envejece. El momento de reevaluar el requisito y el tratamiento de cumpleaños durante la operación deben concretarse.
- Agenda administra odontólogo, paciente, inicio y fin. Los 30 minutos son un valor inicial; el intervalo real determina la disponibilidad.
- Una modificación de duración o reprogramación revalida el intervalo completo en backend. Si falla, conserva la cita anterior.
- La exclusión se aplica por odontólogo. En el futuro, dos profesionales diferentes podrán tener citas a la misma hora si cada uno está disponible; no usar un bloqueo global del establecimiento.
- Interpretar el nombre del padre/apoderado como dato administrativo de una persona relacionada. No exigir su DNI, teléfono ni otros datos no confirmados.
- La mayoría de edad se propone en el modelo como 18 años; validar el criterio operativo antes de cerrar la regla, sin presentar esta propuesta como revisión jurídica.

## Alternativas y razones

Duración fija inmodificable: simple, pero no satisface el ajuste solicitado. Elegir duración con inicio/fin permite variaciones sin convertir la agenda en bloques indivisibles de 30 minutos.

Un único odontólogo implícito: reduce campos hoy, pero obliga a reinterpretar datos históricos al incorporar otro. Guardar la referencia al profesional desde el inicio evita ese problema.

Sobreturnos simultáneos: descartados por la regla expresa del usuario. Registrar sin DNI: fuera del alcance confirmado. Reemplazar el DNI del menor por el del padre: contradice la identificación individual solicitada.

## Ejemplos de aceptación para revisar

- Nueva cita a las 09:00: sugerir fin a las 09:30.
- Ajustar hasta las 09:45: rechazar otra cita a las 09:30 para ese odontólogo.
- Propuesta técnica de intervalos semiabiertos: permitir la siguiente cita exactamente a las 09:45, si el horario lo permite.
- Con dos odontólogos disponibles, permitir una cita por profesional a las 09:00.
- Alta sin DNI del paciente: rechazar.
- Paciente de 12 años o de 66 años: exigir su DNI y nombre del padre/apoderado.
- Paciente de 65 años: exigir DNI; no exigir nombre del apoderado por este umbral de edad.

## Revisión pendiente e impacto

Capítulos 01, 03 y 07, portada, glosario, descubrimiento, modelo preliminar, roadmap y ADR 0002 deben reflejar estas confirmaciones. Contratos de alta/reserva, permisos para ajustar duración, validaciones y futuras migraciones deben incorporar las reglas.

Pendiente: límites e incrementos de duración, quién puede modificarlos, horarios/bloqueos del profesional, momento de evaluar la edad y detalles del vínculo con el apoderado. No inventar esos requisitos ni volver a preguntar duración base, solapamiento, DNI o umbral ya confirmados.

Los escenarios anteriores son criterios documentales; **no son pruebas ejecutadas**. La concurrencia debe probarse con CockroachDB real cuando se implemente. No existen código ni migraciones que modificar en este cambio.


