# Diccionario funcional inicial — Pacientes y Agenda

**CHG-007 · 2026-10-16:** diseño técnico del primer recorrido elaborado por instrucción del usuario. Diccionario con tipos/obligatoriedad, relaciones y restricciones; flujo/transiciones; matriz de permisos y casos de aceptación. Estado del diseño: **Propuesto; revisión de impacto pendiente**. No equivale a migraciones verificadas ni aplicación implementada. Los antecedentes siguientes conservan su historia; para concretar implementación consultar el diseño vigente.

[Leer el diseño técnico vigente](06-diseno-tecnico-pacientes-agenda.md).


**CHG-006 · 2026-10-16:** confirmado: paciente tardío espera un hueco sin desplazar citas o pide reprogramar a otro día; se atiende sin cita cuando hay disponibilidad. Reasignar en este caso significa reprogramar. **Propuesto:** lista de espera sin hora garantizada, liberación explícita y trazable de la reserva no utilizada antes de asignar ese hueco; duración completa mínima de 30 minutos y respeto de la próxima cita. Sin tolerancia ni cancelación automática inventadas. Revisión de impacto pendiente.

Detalle: [ADR 0006](../decisiones/0006-tardanzas-y-espera.md).


Fecha: 2026-10-16. Fuente: diccionario y permisos comunicados por el usuario. CHG-005 / ADR 0005. **Confirmado** indica campos/reglas expresos; **Propuesto** indica diseño aún revisable. No es SQL ejecutable.

## Paciente — confirmado

| Campo | Obligatorio | Observación |
|---|---|---|
| DNI u otro documento | Sí | Tipo y número se proponen como campos separados; tipos admitidos por concretar |
| Nombre | Sí | Texto |
| Apellidos | Sí | Texto |
| Fecha de nacimiento | Sí | Fecha; permite calcular la edad |
| Estado civil | No | Catálogo por definir |
| Ocupación | No | Texto |
| Dirección | Sí | Estructura detallada por definir |
| Teléfono/celular | Sí | Al menos un contacto; se propone permitir teléfono del apoderado para menores, pendiente de validación |
| Correo electrónico | No | Validar formato cuando se informa |

**Cambio vigente:** DNI u otro documento sustituye la obligación exclusiva de DNI del ADR 0003. La identidad técnica del paciente sigue siendo independiente del documento.

## Apoderado — confirmado cuando corresponde

Para menor de edad o paciente de **66 años o más**, de acuerdo con la aclaración previa del usuario, registrar:

| Campo | Obligatorio |
|---|---|
| Nombre y apellidos | Sí |
| Parentesco | Sí |
| DNI u otro documento del apoderado | Sí |

El documento del apoderado no reemplaza al del paciente. Parentesco pertenece al vínculo entre ambos. No deducir permisos de firma ni acceso clínico de ese registro. Momento de evaluación de edad, varios apoderados y criterio de menor de edad (propuesta: menos de 18) siguen por concretar.

## Odontólogo — confirmado

| Campo | Obligatorio |
|---|---|
| DNI | Sí |
| Nombre | Sí |
| Apellidos | Sí |
| Especialidad | Sí |

Pendiente: una o varias especialidades y cómo registrar odontología general. Requisitos de habilitación profesional no se consideran resueltos por esta lista.

## Permisos confirmados

- Recepción: crear cita, registrar llegada/recepción del paciente, cancelar, ajustar duración y reasignar por demora.
- Odontólogo: consultar citas, definir sus horarios y crear/retirar sus bloqueos. Esto incorpora la confirmación del turno previo que había quedado pendiente de sincronizar.
- El odontólogo actual será usuario maestro, con acceso a todas las funciones y también rol de recepción.
- **Propuesto:** representar Maestro como asignación explícita de permisos, separada del perfil profesional; no otorgarla automáticamente a futuros odontólogos. Una cuenta personal puede reunir varios roles.
- Acceso maestro no significa saltarse duración mínima, no solapamiento, integridad ni trazabilidad.

**Confirmado (CHG-006):** ante tardanza, esperar un hueco sin desplazar citas o reprogramar a otro día. No se confirma cambio de odontólogo. Ver ADR 0006 para propuesta de espera y liberación trazable de intervalos.

## Agenda — reglas confirmadas y datos propuestos

Confirmado: duración predeterminada y mínima 30 minutos, ampliable según criterio; sin solapamiento por odontólogo. Horario habitual domingo–viernes 08:00–12:30 y 15:00–19:00; sábado 18:00 hasta 20:00 o 21:00, cierre por fecha pendiente. Zona de operación: Lima.

| Registro propuesto | Datos mínimos propuestos | Regla |
|---|---|---|
| Cita | Identidad interna, paciente, odontólogo, inicio, fin, estado, versión | Duración mínima 30; intervalo completo disponible |
| Horario | Odontólogo, día, inicio/fin, vigencia y zona | Permitir varios tramos diarios |
| Excepción de horario | Odontólogo, fecha, tramos y motivo | Concretar cierre de sábado o ampliación puntual |
| Bloqueo | Odontólogo, inicio, fin, motivo, actor, fecha de registro, estado | Impide nuevas reservas; no cancela citas automáticamente |
| Cambio de cita | Cita, operación, antes/después, actor, fecha, motivo | Conservar reprogramaciones y cancelaciones |
| Llegada | Cita, fecha/hora, actor | Llegada no equivale a atención realizada |
| Control de concurrencia/idempotencia | Según ADR 0002 | Evitar reservas dobles y duplicación por reintentos |

## Propuesta de recorrido y estados

Buscar por documento/nombre → crear o actualizar paciente si corresponde → reservar → registrar llegada → iniciar atención → finalizar atención. Cancelación e inasistencia son salidas alternativas.

Estados propuestos de cita: Programada, Llegó, En atención, Atendida, Cancelada, No asistió. Confirmación telefónica solo se añade si se necesita. Reprogramación es una operación con historial sobre la cita, no un estado terminal. Atención clínica permanece en su contexto; registrar llegada o cerrar agenda no crea contenido clínico automáticamente.

Pendiente: quién inicia/finaliza, quién marca inasistencia y si existe tolerancia; no inventar minutos de espera ni mover citas automáticamente ante retrasos.

## Decisiones necesarias para cerrar el primer recorrido

1. Validar el mecanismo propuesto de espera y liberación/asignación de intervalos (ADR 0006); esperar o reprogramar por tardanza ya está confirmado.
2. Validar estados, responsables y manejo de no asistencia.
3. Validar bloqueo con citas existentes: impedir nuevas reservas y dejar afectadas visibles para gestión explícita, sin cancelación automática.
4. Precisar tipos de documento, teléfono de contacto en menores y cardinalidad de especialidad/apoderado.
5. Fijar cierre por sábado y política al cambiar horarios con citas ya reservadas.

## Diseño técnico propuesto y siguiente entrega

Diccionario físico: tipos, longitudes, nulabilidad, relaciones, índices y restricciones; IDs internos independientes del documento. Proponer unicidad de tipo+número de documento normalizado para evitar fichas duplicadas y una búsqueda que permita recuperar al paciente existente; definir correcciones/duplicados sin borrar historia. Esas reglas técnicas se revisarán antes del SQL.

Probar con casos ficticios: alta con otro documento; menor con documento propio y apoderado completo; adulto de 66; mismo apoderado de dos pacientes; reserva con almuerzo; ampliación que choca con otra cita; dos reservas simultáneas; bloqueo con citas; usuario maestro sujeto a no solapamiento.

Tras validar el recorrido, diseñar migraciones y contratos mínimos, verificar en CockroachDB fijado y construir el recorrido de extremo a extremo. Revisión técnica **pendiente**; no se ejecutaron pruebas de aplicación.



