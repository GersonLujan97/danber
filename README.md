# Danber — sistema de gestión odontológica

**CHG-012 · 2026-10-17:** backend y migrador admiten explícitamente CockroachDB 26.2.0 (instalación del usuario) y 26.2.7. Integración real del backend verificada en 26.2.0; pruebas SQL en informe actualizado. No se cambia el servidor ni los datos del usuario y las ocho migraciones conservan su contenido. Revisión de impacto pendiente; reiniciar API para cargar el ajuste.

[Detalle de compatibilidad](docs/decisiones/0012-compatibilidad-cockroach-26-2-0.md).


**Backend en Notion:** [Estructura, ejecución y pruebas](https://app.notion.com/p/3fc2756ebb4481068fb4eba26fb1228b).

**CHG-011 · 2026-10-17:** flujo de ramas adoptado e incremento feat/backend-inicial implementado: servidor Go, configuración, pool pgx, health/live y health/ready, cierre ordenado, contrato HTTP y CI. Integración local con CockroachDB real; nueva migración 008 restringe CREATE en public. Revisión de impacto pendiente. Cuenta Maestro, sesiones y capacidades de negocio aún no implementadas.

[Ejecutar el backend inicial](backend/README.md). Los estados anteriores se conservan como antecedentes.


**CHG-010 · Adoptado en CHG-011:** [Flujo de ramas para trabajar individualmente](docs/decisiones/0010-flujo-de-ramas.md). main estable, ramas cortas por cambio y PR con comprobaciones. Flujo adoptado; protecciones no configuradas. CI del backend incorporada en CHG-011.


**Guía de ejecución en Notion:** [Crear la base local](https://app.notion.com/p/3fc2756ebb44819a8a6df741dc1f3e50).

**CHG-008 · 2026-10-17:** migraciones iniciales de CockroachDB 26.2.7 generadas y verificadas: siete versiones, 28 tablas más control de migración, SQL único y scripts de ejecución local. **37 comprobaciones correctas** en nodo aislado con datos ficticios. Autorización/API aún pendientes; las reglas de negocio no confirmadas conservan ese estado. Revisión de impacto pendiente.

[Guía para crear la base local](db/README.md). Los antecedentes siguientes conservan su estado histórico.


**Entrega en Notion:** [Diseño técnico · Pacientes y Agenda](https://app.notion.com/p/3fc2756ebb4481eb8043c14d6092af51).

**CHG-007 · 2026-10-16:** diseño técnico del primer recorrido elaborado por instrucción del usuario. Diccionario con tipos/obligatoriedad, relaciones y restricciones; flujo/transiciones; matriz de permisos y casos de aceptación. Estado del diseño: **Propuesto; revisión de impacto pendiente**. No equivale a migraciones verificadas ni aplicación implementada. Los antecedentes siguientes conservan su historia; para concretar implementación consultar el diseño vigente.

[Leer el diseño técnico vigente](docs/base-datos/06-diseno-tecnico-pacientes-agenda.md).


**CHG-006 · 2026-10-16:** confirmado: paciente tardío espera un hueco sin desplazar citas o pide reprogramar a otro día; se atiende sin cita cuando hay disponibilidad. Reasignar en este caso significa reprogramar. **Propuesto:** lista de espera sin hora garantizada, liberación explícita y trazable de la reserva no utilizada antes de asignar ese hueco; duración completa mínima de 30 minutos y respeto de la próxima cita. Sin tolerancia ni cancelación automática inventadas. Revisión de impacto pendiente.

Detalle: [ADR 0006](docs/decisiones/0006-tardanzas-y-espera.md).


**Estado vigente · CHG-005 · 2026-10-16:** paciente con DNI u otro documento; campos obligatorios/opcionales según diccionario del usuario. Apoderado para menores y desde los 66 años: nombres/apellidos, parentesco y documento obligatorios. Odontólogo: DNI, nombre, apellidos y especialidad obligatorios. Recepción crea, recibe, cancela, ajusta duración y reasigna por demora (alcance pendiente). Odontólogo gestiona sus horarios y bloqueos; el odontólogo actual es Maestro y también Recepción. **Revisión de impacto pendiente:** estados/responsables de citas, reasignación, conflictos por ausencias y detalles del diccionario. Sustituye parcialmente ADR 0003 en identificación/apoderado y complementa ADR 0004 en permisos; los registros previos se conservan como historia.

Detalle vigente: [ADR 0005](docs/decisiones/0005-identificacion-y-permisos.md).


**CHG-004 · 2026-10-16 — Confirmado:** recepción ajusta duración; una persona puede tener roles de odontólogo y recepción. Citas de 30 minutos por defecto y como mínimo, ampliables según criterio, sin máximo ni paso fijo acordados. Domingo–viernes: 08:00–12:30 y 15:00–19:00; sábado desde 18:00 hasta 20:00 o 21:00 (cierre por concretar). **Propuesto:** bloqueos fechados por odontólogo para ausencias, sin cancelar automáticamente citas existentes; mostrarlas para gestión explícita. **Pendiente de revisión:** cierre sabatino, permisos de bloqueos/horarios y flujo de citas afectadas. Esta actualización resuelve los pendientes anteriores sobre mínimo, recepción y horario habitual; conserva los demás.

Detalle: [ADR 0004](docs/decisiones/0004-horarios-roles-y-bloqueos.md).


**Confirmado · CHG-003 · 2026-10-16:** citas de 30 minutos por defecto con duración ajustable; un odontólogo actual y soporte para varios; sin solapamientos por odontólogo; DNI propio obligatorio para todos los pacientes; nombre del padre o apoderado para menores y pacientes de 66 años o más. A los 65 no aplica ese requisito por edad. **Pendiente de revisión:** límites y permisos del ajuste de duración, horarios, evaluación de edad y detalle del vínculo con el apoderado. El dato del apoderado no concede por sí solo permisos de firma o acceso.

Detalle y consecuencias: [ADR 0003](docs/decisiones/0003-citas-e-identificacion.md). Las propuestas anteriores se leen con esta actualización vigente.


## Documentación principal en Notion

**Empieza por [Danber · Manual del proyecto y decisiones](https://app.notion.com/p/3fb2756ebb4481698e0efb11afbf684d).** Es el punto de lectura del proyecto desde el 16 de octubre de 2026, por instrucción del usuario. Reúne siete capítulos ordenados, explicaciones completas, diagramas y los tres formatos originales.

- [01 · Entender el consultorio](https://app.notion.com/p/3fb2756ebb4481078f27fa7e3f26b172)
- [02 · Comprender la arquitectura](https://app.notion.com/p/3fb2756ebb4481298ed3e48303edcd98)
- [03 · Entender los datos y las reservas](https://app.notion.com/p/3fb2756ebb4481fd9c47c4d8d18b4423)
- [04 · Construir por capacidades](https://app.notion.com/p/3fb2756ebb448191b94df323a221db57)
- [05 · Probar y proteger el sistema](https://app.notion.com/p/3fb2756ebb44812980c5ce05ca042f2f)
- [06 · Aprender los conceptos](https://app.notion.com/p/3fb2756ebb44810ca130cc5179d73e03)
- [07 · Decisiones, cambios y próximos pasos](https://app.notion.com/p/3fb2756ebb44813d8f89fe0371e9edbd)

**Regla obligatoria:** todo cambio de decisiones debe quedar registrado en Notion con fecha, antes/después, motivo, impacto y qué debe revisarse. Conservar el historial y distinguir decisión confirmada de revisión de impacto pendiente. El procedimiento está en el capítulo 07 y en `AGENTS.md`.

Los archivos Markdown se conservan como respaldo técnico versionable. Notion es la documentación principal; no mantener una segunda versión de decisiones sin sincronizar. Si Notion no está accesible, declarar la sincronización pendiente. Esta regla no instala sincronización automática.

## Estado vigente — 16 de octubre de 2026

**CHG-002:** ADR 0001 y ADR 0002 aceptados provisionalmente tras la lectura del usuario. Reglas abiertas y revisión de impacto pendientes. Próximo paso propuesto: fase D del primer recorrido, comenzando por reglas y diccionario de datos. Registro completo en el capítulo 07 de Notion. La entrega inferior conserva su estado histórico del 14 de octubre.

## Referencia de la entrega local original — 14 de octubre de 2026

Entrega inicial de arquitectura y descubrimiento, 14 de octubre de 2026. **Estado: propuesta para revisión; todavía no hay aplicación, base de datos ni migraciones ejecutables.**

El objetivo es acompañar la atención odontológica desde la identificación del paciente y su cita hasta la evolución clínica, los tratamientos y su gestión económica. La primera implementación propuesta es **registrar un paciente → reservar una cita → consultar la agenda**, con autenticación, autorización y prevención de reservas simultáneas.

## Decisiones y evidencia

- **Confirmado:** operación inicial en Perú; tecnologías previstas Go, React con TypeScript y CockroachDB; trabajo por fases y explicación de decisiones antes de generar la aplicación.
- **Confirmado por el usuario:** «Costo» en la hoja de evolución significa precio o importe del procedimiento realizado. No identifica un pago recibido.
- **Confirmado por el usuario:** el recurso para reservar es únicamente el odontólogo. No se incluye reserva de sillones o consultorios.
- **Observado:** existen tres formatos vacíos en [Imagenes](Imagenes): historia clínica, odontograma y evolución de tratamiento. Son evidencia de información y vocabulario; su presencia no demuestra obligatoriedad de todos los campos ni conformidad normativa.
- **Propuesto:** monolito modular, ocho contextos de negocio, API HTTP orientada a recursos y operaciones, esquemas SQL por propietario y desarrollo incremental.
- **Pendiente de validación:** duración de citas, sobreturnos, reglas de cobro, correcciones clínicas, menores, numeración de historias, permisos detallados y condiciones de despliegue.

## Orden de lectura

| Documento | Para qué sirve |
|---|---|
| [A. Descubrimiento y formatos](docs/dominio/01-descubrimiento.md) | Conecta las imágenes y los procesos con las reglas por confirmar |
| [Lenguaje ubicuo](docs/glosario/lenguaje-ubicuo.md) | Evita confundir cita, atención, procedimiento, precio y pago |
| [B. Contextos y mapa de relaciones](docs/dominio/02-contextos.md) | Asigna responsabilidades y propiedad de datos |
| [C. Arquitectura propuesta](docs/arquitectura/03-propuesta.md) | Explica dependencias, comunicación y alternativas |
| [Guía de conceptos](docs/arquitectura/guia-conceptos.md) | Enseña cada concepto con problemas y ejemplos del consultorio |
| [Modelo conceptual y lógico preliminar](docs/base-datos/04-modelo-preliminar.md) | Incluye Mermaid, concurrencia y plan de migraciones |
| [Estructura Go y React](docs/arquitectura/05-estructura.md) | Propone carpetas y bibliotecas sin crearlas anticipadamente |
| [Contratos y documentación de módulos](docs/modulos/06-contratos.md) | Define intercambios, fallos y plantilla reutilizable |
| [Pruebas](docs/pruebas/07-estrategia.md) | Define evidencias de independencia, concurrencia y seguridad |
| [Seguridad y normativa peruana](docs/seguridad/08-seguridad-y-normativa.md) | Registra controles propuestos y fuentes oficiales a validar |
| [Roadmap y decisiones pendientes](docs/09-roadmap-y-validaciones.md) | Delimita el primer recorrido y el siguiente paso |
| [ADR 0001](docs/decisiones/0001-monolito-y-contextos.md) / [ADR 0002](docs/decisiones/0002-propiedad-y-consistencia.md) | Conservan razones, alternativas y consecuencias de decisiones propuestas |

Las fases A, B y C están desarrolladas como propuesta. D, E, F y G se anticipan solo lo necesario para comprobar su coherencia. H, el esqueleto ejecutable, queda para después de revisar las decisiones. No se considera aprobado un ADR por estar escrito.

## Cómo mantener esta documentación

Cada decisión debe marcarse como **Confirmado**, **Propuesto** o **Pendiente de validación**, con origen y fecha cuando cambie. Los términos del glosario deben coincidir en documentación, contratos y futuro código. Las fuentes normativas y técnicas se enlazan junto a las afirmaciones que respaldan; su consulta no constituye certificación legal ni prueba de ejecución.

La carpeta no era un repositorio Git al iniciar esta entrega. Los archivos están preparados para versionarse cuando se inicialice el proyecto; esta entrega no crea commits ni publica información.













