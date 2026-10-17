# Roadmap incremental y revisión de decisiones

**CHG-008 · 2026-10-17:** migraciones iniciales de CockroachDB 26.2.7 generadas y verificadas: siete versiones, 28 tablas más control de migración, SQL único y scripts de ejecución local. **37 comprobaciones correctas** en nodo aislado con datos ficticios. Autorización/API aún pendientes; las reglas de negocio no confirmadas conservan ese estado. Revisión de impacto pendiente.

[Guía para crear la base local](../db/README.md). Los antecedentes siguientes conservan su estado histórico.


**CHG-007 · 2026-10-16:** diseño técnico del primer recorrido elaborado por instrucción del usuario. Diccionario con tipos/obligatoriedad, relaciones y restricciones; flujo/transiciones; matriz de permisos y casos de aceptación. Estado del diseño: **Propuesto; revisión de impacto pendiente**. No equivale a migraciones verificadas ni aplicación implementada. Los antecedentes siguientes conservan su historia; para concretar implementación consultar el diseño vigente.

[Leer el diseño técnico vigente](base-datos/06-diseno-tecnico-pacientes-agenda.md).


**CHG-006 · 2026-10-16:** confirmado: paciente tardío espera un hueco sin desplazar citas o pide reprogramar a otro día; se atiende sin cita cuando hay disponibilidad. Reasignar en este caso significa reprogramar. **Propuesto:** lista de espera sin hora garantizada, liberación explícita y trazable de la reserva no utilizada antes de asignar ese hueco; duración completa mínima de 30 minutos y respeto de la próxima cita. Sin tolerancia ni cancelación automática inventadas. Revisión de impacto pendiente.

Detalle: [ADR 0006](decisiones/0006-tardanzas-y-espera.md).


**Estado vigente · CHG-005 · 2026-10-16:** paciente con DNI u otro documento; campos obligatorios/opcionales según diccionario del usuario. Apoderado para menores y desde los 66 años: nombres/apellidos, parentesco y documento obligatorios. Odontólogo: DNI, nombre, apellidos y especialidad obligatorios. Recepción crea, recibe, cancela, ajusta duración y reasigna por demora (alcance pendiente). Odontólogo gestiona sus horarios y bloqueos; el odontólogo actual es Maestro y también Recepción. **Revisión de impacto pendiente:** estados/responsables de citas, reasignación, conflictos por ausencias y detalles del diccionario. Sustituye parcialmente ADR 0003 en identificación/apoderado y complementa ADR 0004 en permisos; los registros previos se conservan como historia.

Detalle vigente: [ADR 0005](decisiones/0005-identificacion-y-permisos.md).


**CHG-004 · 2026-10-16 — Confirmado:** recepción ajusta duración; una persona puede tener roles de odontólogo y recepción. Citas de 30 minutos por defecto y como mínimo, ampliables según criterio, sin máximo ni paso fijo acordados. Domingo–viernes: 08:00–12:30 y 15:00–19:00; sábado desde 18:00 hasta 20:00 o 21:00 (cierre por concretar). **Propuesto:** bloqueos fechados por odontólogo para ausencias, sin cancelar automáticamente citas existentes; mostrarlas para gestión explícita. **Pendiente de revisión:** cierre sabatino, permisos de bloqueos/horarios y flujo de citas afectadas. Esta actualización resuelve los pendientes anteriores sobre mínimo, recepción y horario habitual; conserva los demás.

Detalle: [ADR 0004](decisiones/0004-horarios-roles-y-bloqueos.md).


**Confirmado · CHG-003 · 2026-10-16:** citas de 30 minutos por defecto con duración ajustable; un odontólogo actual y soporte para varios; sin solapamientos por odontólogo; DNI propio obligatorio para todos los pacientes; nombre del padre o apoderado para menores y pacientes de 66 años o más. A los 65 no aplica ese requisito por edad. **Pendiente de revisión:** límites y permisos del ajuste de duración, horarios, evaluación de edad y detalle del vínculo con el apoderado. El dato del apoderado no concede por sí solo permisos de firma o acceso.

Detalle y consecuencias: [ADR 0003](decisiones/0003-citas-e-identificacion.md). Las propuestas anteriores se leen con esta actualización vigente.


Estado: **Propuesto**, 2026-10-14. Esta entrega desarrolla A–C y anticipa decisiones necesarias de D–G. No implica autorización de puesta en producción ni aprobación silenciosa de las propuestas.

## 1. Fases y resultados revisables

| Fase | Resultado | Criterio de salida |
|---|---|---|
| A. Descubrimiento | Procesos, evidencia de formatos y vocabulario | Acordar significado de términos y resolver reglas que afectan el primer recorrido |
| B. Contextos | Propiedad, límites, contratos candidatos y mapa | Revisar que cada dato/regla tenga dueño y las separaciones aporten valor |
| C. Arquitectura | Monolito modular, dependencias, alternativas y ADR | Revisar propuesta y consecuencias; registrar aceptadas o modificadas |
| D. Datos | Diccionario lógico del MVP, restricciones, índices y SQL verificado | Migraciones ejecutadas en CockroachDB fijado, con prueba de integridad y concurrencia |
| E. Repositorio | Estructura mínima Go/React, contratos HTTP | Carpetas ligadas a casos reales; versión y herramientas fijadas |
| F. Comunicación, pruebas y seguridad | Contratos operativos, matriz de acceso y plan de pruebas | Incorporar esas comprobaciones a cada caso; no tratarlas como tarea final |
| G. Roadmap | Recorridos pequeños y dependencias | Prioridad acordada por valor y riesgo, sin calendario inventado |
| H. Esqueleto e implementación controlada | Aplicación mínima que compila, corre y se prueba | Recorrido de extremo a extremo con evidencia de aceptación |

D–G pueden refinarse en conjunto con un alcance pequeño; no hace falta especificar todo el sistema clínico para crear la primera migración de pacientes.

## 2. Primera implementación vertical

**Recorrido propuesto:** iniciar sesión → registrar/buscar paciente → seleccionar odontólogo y horario → reservar → consultar agenda del día.

Infraestructura mínima:

- Go con API HTTP, configuración, logs mínimos sin datos sensibles y conexión a CockroachDB.
- React/TypeScript con login, formulario de paciente, formulario de reserva y agenda diaria.
- Identidad y permisos mínimos; personal/odontólogo y horario inicial configurados mediante un mecanismo controlado. No hace falta una administración completa de personal para el recorrido, pero sí datos reales de disponibilidad en el entorno de prueba.
- Migraciones de Identidad, Personal, Pacientes y Agenda, historial, idempotencia y auditoría pertinente. Cuenta migradora separada.
- Instancia de desarrollo y base efímera para integración con datos sintéticos; versiones fijadas.
- HTTPS en un entorno compartido, secretos seguros y copia/restauración antes de usar datos reales.

No incluye historias completas, odontograma interactivo, tratamientos longitudinales, cobros, recetas ni emisión tributaria. Se conservarán las referencias para incorporarlos después con límites ya definidos.

Criterios de aceptación:

1. Un usuario autorizado registra un paciente con el mínimo de datos acordado y lo encuentra después.
2. Solo se ofrece y confirma disponibilidad conforme al horario y bloqueos del odontólogo.
3. Dos solicitudes competidoras dejan una sola reserva ordinaria; los intervalos adyacentes son posibles.
4. Repetir una solicitud por fallo de red no crea otra cita.
5. La agenda muestra la reserva persistida y no expone información clínica.
6. Una persona sin permisos no ejecuta las operaciones; las acciones quedan atribuidas.
7. Pruebas, versión del motor y migraciones son reproducibles en un ambiente separado.

## 3. Orden posterior por valor y dependencia

| Incremento | Capacidad | Dependencias y evidencia |
|---|---|---|
| 1 | Primer recorrido de paciente y reserva | Contratos de identificación, elegibilidad y reserva; concurrencia probada |
| 2 | Agenda operativa completa | Reprogramación, cancelación, confirmación, llegada, inasistencia y atención sin cita; sin sobreturnos simultáneos para el mismo odontólogo (CHG-003) |
| 3 | Atención clínica mínima | Identificación, autoría, política de acceso, evolución y corrección; validación normativa y profesional |
| 4 | Catálogo de procedimientos y economía básica | Registrar hechos clínicos, definir tarifas/cargos/pagos; fuente única de deuda e idempotencia |
| 5 | Planes, tratamientos y controles | Versiones aceptadas, coordinación con Agenda y resultados de Clínica; tareas recuperables |
| 6 | Odontograma y documentos clínicos ampliados | Convenciones revisadas, dentición y superficies, historial e impresión verificada |
| 7 | Recetas | Reglas profesionales, campos/firma y trazabilidad confirmados |
| 8 | Facturación electrónica y recordatorios | Contratos con proveedores, privacidad, conciliación, reintentos y contingencias |

El odontograma puede adelantarse si es imprescindible para validar la atención mínima; el orden es una propuesta de entrega, no una afirmación clínica sobre qué debe contener una historia real. El MVP clínico se define con el odontólogo antes de sustituir un registro oficial.

Identidad, registro administrativo de pacientes y personal pueden desarrollarse de forma independiente una vez fijados contratos. Economía puede avanzar con escenarios ficticios publicados sin importar el dominio clínico. Facturación necesita una base económica estable; seguimiento necesita conceptos de control, cita y resultado clínico, aunque no exija que todo lo demás esté terminado.

## 4. Decisiones que necesitan revisión

| ID | Tema | Recomendación actual y consecuencia | Cuándo resolver |
|---|---|---|---|
| V01 | Monolito y ocho contextos | Aceptar límites propuestos; una unidad de despliegue y contratos explícitos | Antes de H |
| V02 | Cita: duración y horarios | Confirmado: 30 minutos ajustables y sin solapamientos por odontólogo. Pendientes: límites, permisos de ajuste y horarios | Antes de SQL definitivo de Agenda |
| V03 | DNI y padre/apoderado | Confirmado: DNI para todos; nombre del padre/apoderado para menores y desde los 66 años. Pendiente: momento de evaluación y detalles del vínculo | Antes del alta de pacientes real |
| V04 | Habilitación y baja profesional | Consultar elegibilidad; revisar futuras citas cuando cambia estado | Antes de Agenda operativa; decidir si se exige atomicidad estricta |
| V05 | Moneda y generación de deuda | Separar tarifa, importe de procedimiento, cargo y pago | Antes de Economía |
| V06 | Clínica y firma | Registro finalizado con correcciones trazables; validar forma legal de autoría | Antes de uso clínico real |
| V07 | Odontograma | Validar símbolos, dentición temporal/mixta, superficies y edición histórica | Antes de su diseño detallado |
| V08 | Controles prolongados | Recomendación/ventana independiente de reserva; una reserva activa por control | Antes de Seguimiento |
| V09 | Infraestructura | Un consultorio, en línea; evaluar alojamiento y costo de CockroachDB | Antes de elegir ambiente compartido |
| V10 | Acceso y privacidad | Mínimo privilegio por operación/recurso; obligaciones concretas según establecimiento | Antes de datos reales |
| V11 | Tributación | Emisor, tipo de comprobante, impuestos y momento de emisión validados por contabilidad | Antes de Facturación |

Confirmaciones ya cerradas: **Costo = importe del procedimiento realizado** y **reserva solo por odontólogo**. No volver a tratarlas como dudas ni añadir sillones sin una nueva necesidad.

## 5. Siguiente paso concreto

Con V01 aceptado provisionalmente y las reglas principales de V02–V03 confirmadas (CHG-003), desarrollar el diccionario del primer recorrido. Revisar ajuste de duración con conflicto, DNI obligatorio y registro del padre/apoderado para menores y desde los 66 años. Con esas decisiones se desarrolla el diccionario del primer recorrido y sus migraciones, y después el esqueleto pequeño con pruebas desde el inicio. Las decisiones clínicas y económicas posteriores permanecen abiertas sin impedir trabajo administrativo que no dependa de ellas.

Cada nueva capacidad se incorpora como un recorrido que pueda demostrarse y probarse. No completar todas las tablas futuras como condición para empezar, ni presentar esta documentación como un sistema terminado.










