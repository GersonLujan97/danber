# Instrucciones de trabajo — Danber

**CHG-015 · 2026-10-18:** registro y búsqueda administrativa de pacientes implementados en `feat/registro-pacientes`; validación, permisos, apoderado, auditoría e idempotencia (011). Verificado localmente en CockroachDB 26.2.0 y 26.2.7; revisión de impacto pendiente. [Guía principal en Notion](https://app.notion.com/p/3fd2756ebb4481c79e62cc2616020ffb) y [respaldo técnico](backend/PACIENTES.md) y [ADR 0015](docs/decisiones/0015-registro-y-busqueda-pacientes.md). PR #3 de autenticación debe integrarse primero.

**CHG-014 · 2026-10-18:** main actualizado a c0fb8ca (PR #2 integrado). El usuario informa creación correcta de su cuenta. Autenticación HTTP y seguridad por incremento implementadas y verificadas localmente en `feat/autenticacion`: sesiones, CSRF/origen, límites de solicitudes/CPU y revocación. Nueva migración 010; no volver a crear la cuenta. Revisión de impacto pendiente.

[Acceso y seguridad en Notion](https://app.notion.com/p/3fd2756ebb4481cfa6ddc7e41e798fe9): ejecución, motivos, ataques de ejemplo, límites y controles previstos. Respaldo: [guía integrada](backend/AUTENTICACION-Y-SEGURIDAD.md) y [ADR 0014](docs/decisiones/0014-autenticacion-y-seguridad-incremental.md).

**CHG-013 · 2026-10-17:** PR #1 integrado y copia local actualizada a main 583343b. Cuenta Maestro inicial por consola en `feat/cuenta-inicial`: contraseña sin eco, Argon2id, roles Maestro + Recepción y auditoría atómicos; migración 009. Implementado y verificado localmente; revisión de impacto pendiente. Tu cuenta real se crea después en tu terminal. Sesiones aún pendientes.

Guía principal: [Cuenta Maestro inicial en Notion](https://app.notion.com/p/3fc2756ebb448101ba65e6e54e1532ca). Respaldo técnico: [crear cuenta](backend/CUENTA-INICIAL.md) y [ADR 0013](docs/decisiones/0013-cuenta-maestro-inicial.md).

**CHG-012 · 2026-10-17:** backend y migrador admiten explícitamente CockroachDB 26.2.0 (instalación del usuario) y 26.2.7. Integración real del backend verificada en 26.2.0; pruebas SQL en informe actualizado. No se cambia el servidor ni los datos del usuario y las ocho migraciones conservan su contenido. Revisión de impacto pendiente; reiniciar API para cargar el ajuste.

[Detalle de compatibilidad](docs/decisiones/0012-compatibilidad-cockroach-26-2-0.md).


**CHG-011 · 2026-10-17:** flujo de ramas adoptado e incremento feat/backend-inicial implementado: servidor Go, configuración, pool pgx, health/live y health/ready, cierre ordenado, contrato HTTP y CI. Integración local con CockroachDB real; nueva migración 008 restringe CREATE en public. Revisión de impacto pendiente. Cuenta Maestro, sesiones y capacidades de negocio aún no implementadas.

[Ejecutar el backend inicial](backend/README.md). Los estados anteriores se conservan como antecedentes.


**CHG-008 · 2026-10-17:** migraciones iniciales de CockroachDB 26.2.7 generadas y verificadas: siete versiones, 28 tablas más control de migración, SQL único y scripts de ejecución local. **37 comprobaciones correctas** en nodo aislado con datos ficticios. Autorización/API aún pendientes; las reglas de negocio no confirmadas conservan ese estado. Revisión de impacto pendiente.

[Guía para crear la base local](db/README.md). Los antecedentes siguientes conservan su estado histórico.


**CHG-007 · 2026-10-16:** diseño técnico del primer recorrido elaborado por instrucción del usuario. Diccionario con tipos/obligatoriedad, relaciones y restricciones; flujo/transiciones; matriz de permisos y casos de aceptación. Estado del diseño: **Propuesto; revisión de impacto pendiente**. No equivale a migraciones verificadas ni aplicación implementada. Los antecedentes siguientes conservan su historia; para concretar implementación consultar el diseño vigente.

[Leer el diseño técnico vigente](docs/base-datos/06-diseno-tecnico-pacientes-agenda.md).


**CHG-006 · 2026-10-16:** confirmado: paciente tardío espera un hueco sin desplazar citas o pide reprogramar a otro día; se atiende sin cita cuando hay disponibilidad. Reasignar en este caso significa reprogramar. **Propuesto:** lista de espera sin hora garantizada, liberación explícita y trazable de la reserva no utilizada antes de asignar ese hueco; duración completa mínima de 30 minutos y respeto de la próxima cita. Sin tolerancia ni cancelación automática inventadas. Revisión de impacto pendiente.

Detalle: [ADR 0006](docs/decisiones/0006-tardanzas-y-espera.md).


**Estado vigente · CHG-005 · 2026-10-16:** paciente con DNI u otro documento; campos obligatorios/opcionales según diccionario del usuario. Apoderado para menores y desde los 66 años: nombres/apellidos, parentesco y documento obligatorios. Odontólogo: DNI, nombre, apellidos y especialidad obligatorios. Recepción crea, recibe, cancela, ajusta duración y reasigna por demora (alcance pendiente). Odontólogo gestiona sus horarios y bloqueos; el odontólogo actual es Maestro y también Recepción. **Revisión de impacto pendiente:** estados/responsables de citas, reasignación, conflictos por ausencias y detalles del diccionario. Sustituye parcialmente ADR 0003 en identificación/apoderado y complementa ADR 0004 en permisos; los registros previos se conservan como historia.

Detalle vigente: [ADR 0005](docs/decisiones/0005-identificacion-y-permisos.md).


**CHG-004 · 2026-10-16 — Confirmado:** recepción ajusta duración; una persona puede tener roles de odontólogo y recepción. Citas de 30 minutos por defecto y como mínimo, ampliables según criterio, sin máximo ni paso fijo acordados. Domingo–viernes: 08:00–12:30 y 15:00–19:00; sábado desde 18:00 hasta 20:00 o 21:00 (cierre por concretar). **Propuesto:** bloqueos fechados por odontólogo para ausencias, sin cancelar automáticamente citas existentes; mostrarlas para gestión explícita. **Pendiente de revisión:** cierre sabatino, permisos de bloqueos/horarios y flujo de citas afectadas. Esta actualización resuelve los pendientes anteriores sobre mínimo, recepción y horario habitual; conserva los demás.

Detalle: [ADR 0004](docs/decisiones/0004-horarios-roles-y-bloqueos.md).


## Alcance y estado

El proyecto avanza por incrementos autorizados; conserva el descubrimiento y las decisiones arquitectónicas como antecedentes. Leer `README.md` y el estado vigente en Notion —glosario, contextos, ADR y bitácora— antes de proponer cambios. Los ADR iniciales fueron aceptados provisionalmente el 2026-10-16 (CHG-002), tras la lectura del usuario. Las reglas abiertas y la revisión de impacto técnico siguen pendientes; no interpretar conformidad como pruebas completadas. Seguir la fase solicitada: backend, cuenta inicial y autenticación tienen sus propios registros y evidencias posteriores; la entrega A–C es histórica. No generar toda la aplicación por adelantado.

Estas instrucciones orientan el trabajo; la autorización y las decisiones explícitas del usuario prevalecen. No convertir pendientes en requisitos confirmados ni solicitar de nuevo información ya confirmada.

## Notion obligatorio — decisión DOC-01, confirmada el 2026-10-16

- Documentación principal: [Danber · Manual del proyecto y decisiones](https://app.notion.com/p/3fb2756ebb4481698e0efb11afbf684d).
- Registro obligatorio: [Decisiones, cambios y próximos pasos](https://app.notion.com/p/3fb2756ebb44813d8f89fe0371e9edbd).
- Notion es el punto principal de lectura y registro de decisiones. Los Markdown se conservan como respaldo técnico versionable; no mantener dos fuentes de decisiones independientes.
- En todo trabajo que proponga, confirme, sustituya o cambie una decisión, consultar Notion y registrar allí una entrada CHG correlativa con fecha, autor/origen conocido, decisión afectada, antes/después, motivo, alternativas, consecuencias y qué debe revisarse.
- Indicar capítulos, reglas, contratos, datos/migraciones, permisos, código y pruebas afectados. Marcar la revisión de impacto como **Pendiente de revisión**, añadir aviso en capítulos afectados y actualizar el resumen de la portada.
- Separar el estado de la decisión del estado de revisión. Una instrucción explícita del usuario puede confirmar la decisión, pero no demuestra que sus impactos ya se revisaron.
- Conservar historia: un ADR reemplazado se marca **Sustituido** y enlaza al sucesor; no borrar el razonamiento anterior. Registrar responsable, fecha y evidencia al cerrar una revisión.
- Actualizar Notion y los documentos técnicos afectados en el mismo trabajo; verificar la escritura leyendo el resultado. No declarar terminada la actualización documental si falta sincronizar Notion.
- Si Notion no está accesible, registrar localmente el cambio pendiente de sincronizar y comunicar la limitación; completar la sincronización al recuperar acceso. No presentar la copia local como si Notion ya estuviera actualizado.
- No aprobar automáticamente ADR propuestos al trasladarlos o editarlos. No publicar ni mover esta documentación a una ubicación compartida sin instrucción del usuario.
- Esta obligación se aplica al trabajo futuro sobre el proyecto; no supone que exista un servicio de sincronización automática ni supervisión de ediciones externas.

## Confirmaciones del consultorio

- El significado de «Costo» en evolución es precio o importe del procedimiento realizado; no pago recibido.
- La reserva depende solo del odontólogo. No añadir sillón o consultorio como recurso sin una necesidad nueva.
- Operación inicial en Perú; Go, React/TypeScript y CockroachDB son tecnologías previstas.
- Las imágenes originales están en `Imagenes/`; conservarlas y no afirmar que cumplen normas clínicas por su apariencia.

## Comunicación y enseñanza

Trabajar en español. Explicar decisiones y alternativas con ejemplos odontológicos. Al introducir un concepto relevante, explicar qué significa, qué problema resuelve, por qué importa, qué alternativa existe, el riesgo de omitirlo y un ejemplo. Se puede enlazar la guía de conceptos ya documentada para evitar repetir definiciones extensas.

Marcar Confirmado, Propuesto y Pendiente de validación. Registrar decisiones importantes en ADR con fecha, estado, alternativas y consecuencias. No afirmar cumplimiento legal, compatibilidad SQL ni pruebas ejecutadas sin evidencia.

## Diseño y código cuando se autorice su fase

- Organizar por capacidades del negocio; dominio independiente de HTTP, drivers SQL y React.
- Cada módulo posee reglas y datos. Usar contratos mínimos para otros módulos; no importar sus implementaciones ni consultar sus tablas indiscriminadamente.
- No usar entidades globales compartidas, repositorios genéricos, microservicios, colas o capas adicionales sin una necesidad explicada.
- Mantener separados cita/atención, plan/tratamiento, importe/cargo/pago/comprobante y usuario/personal/profesional/rol.
- Regla definitiva en backend; frontend ayuda a ingresar datos. Autorización por operación y recurso, con mínimo privilegio.
- Controlar concurrencia e idempotencia. Una consulta previa de disponibilidad o una unicidad por hora inicial no evita todos los solapamientos.
- Verificar sintaxis y comportamiento con la versión fijada de CockroachDB. No asumir compatibilidad completa con SQL Server o PostgreSQL.
- Cambios clínicos finalizados y económicos necesitan trazabilidad; no introducir eliminación en cascada de historia o movimientos.

## Verificación y datos

Aplicar pruebas proporcionadas al cambio. Para persistencia/concurrencia, usar CockroachDB real en entorno aislado; sustitutos sirven para casos de uso, no para demostrar aislamiento SQL. Nunca usar datos de pacientes reales en fixtures ni credenciales de producción en pruebas.

Para cambios documentales, revisar enlaces relativos, coherencia del glosario, diagramas y estados de decisiones. No crear pruebas de aplicación inexistente. Conservar archivos existentes y no publicar ni enviar información a terceros sin instrucción expresa.


## Confirmaciones adicionales — 2026-10-16, CHG-003

**Confirmado · CHG-003 · 2026-10-16:** citas de 30 minutos por defecto con duración ajustable; un odontólogo actual y soporte para varios; sin solapamientos por odontólogo; DNI propio obligatorio para todos los pacientes; nombre del padre o apoderado para menores y pacientes de 66 años o más. A los 65 no aplica ese requisito por edad. **Pendiente de revisión:** límites y permisos del ajuste de duración, horarios, evaluación de edad y detalle del vínculo con el apoderado. El dato del apoderado no concede por sí solo permisos de firma o acceso.

Ver docs/decisiones/0003-citas-e-identificacion.md. No volver a tratar estas reglas como pendientes de confirmación.








