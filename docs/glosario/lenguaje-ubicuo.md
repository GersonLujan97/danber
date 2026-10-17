# Lenguaje ubicuo de Danber

**CHG-007 · 2026-10-16:** diseño técnico del primer recorrido elaborado por instrucción del usuario. Diccionario con tipos/obligatoriedad, relaciones y restricciones; flujo/transiciones; matriz de permisos y casos de aceptación. Estado del diseño: **Propuesto; revisión de impacto pendiente**. No equivale a migraciones verificadas ni aplicación implementada. Los antecedentes siguientes conservan su historia; para concretar implementación consultar el diseño vigente.

[Leer el diseño técnico vigente](../base-datos/06-diseno-tecnico-pacientes-agenda.md).


**Estado vigente · CHG-005 · 2026-10-16:** paciente con DNI u otro documento; campos obligatorios/opcionales según diccionario del usuario. Apoderado para menores y desde los 66 años: nombres/apellidos, parentesco y documento obligatorios. Odontólogo: DNI, nombre, apellidos y especialidad obligatorios. Recepción crea, recibe, cancela, ajusta duración y reasigna por demora (alcance pendiente). Odontólogo gestiona sus horarios y bloqueos; el odontólogo actual es Maestro y también Recepción. **Revisión de impacto pendiente:** estados/responsables de citas, reasignación, conflictos por ausencias y detalles del diccionario. Sustituye parcialmente ADR 0003 en identificación/apoderado y complementa ADR 0004 en permisos; los registros previos se conservan como historia.

Detalle vigente: [ADR 0005](../decisiones/0005-identificacion-y-permisos.md).


**CHG-004 · 2026-10-16 — Confirmado:** recepción ajusta duración; una persona puede tener roles de odontólogo y recepción. Citas de 30 minutos por defecto y como mínimo, ampliables según criterio, sin máximo ni paso fijo acordados. Domingo–viernes: 08:00–12:30 y 15:00–19:00; sábado desde 18:00 hasta 20:00 o 21:00 (cierre por concretar). **Propuesto:** bloqueos fechados por odontólogo para ausencias, sin cancelar automáticamente citas existentes; mostrarlas para gestión explícita. **Pendiente de revisión:** cierre sabatino, permisos de bloqueos/horarios y flujo de citas afectadas. Esta actualización resuelve los pendientes anteriores sobre mínimo, recepción y horario habitual; conserva los demás.

Detalle: [ADR 0004](../decisiones/0004-horarios-roles-y-bloqueos.md).


**Confirmado · CHG-003 · 2026-10-16:** citas de 30 minutos por defecto con duración ajustable; un odontólogo actual y soporte para varios; sin solapamientos por odontólogo; DNI propio obligatorio para todos los pacientes; nombre del padre o apoderado para menores y pacientes de 66 años o más. A los 65 no aplica ese requisito por edad. **Pendiente de revisión:** límites y permisos del ajuste de duración, horarios, evaluación de edad y detalle del vínculo con el apoderado. El dato del apoderado no concede por sí solo permisos de firma o acceso.

Detalle y consecuencias: [ADR 0003](../decisiones/0003-citas-e-identificacion.md). Las propuestas anteriores se leen con esta actualización vigente.


Estado: propuesta inicial; las dos confirmaciones del usuario están indicadas. Lenguaje ubicuo significa usar las mismas palabras con el mismo significado entre consultorio, documentación, pruebas y código. Evita que «tratamiento» signifique una deuda en una pantalla y una intervención en otra. La alternativa es mantener sinónimos locales y traducirlos constantemente, con riesgo de registrar o cobrar el hecho equivocado.

## Identidad y organización

| Término | Significado propuesto | Ejemplo o distinción |
|---|---|---|
| Persona | Ser humano del mundo real; no implica una tabla universal | Rosa puede ser recepcionista y paciente |
| Usuario | Cuenta que puede autenticarse; tiene estado, credenciales y sesiones | `usuario_id` no es `personal_id` ni `paciente_id` |
| Personal | Vínculo administrativo de alguien con el consultorio | Un técnico puede estar registrado sin cuenta |
| Profesional | Perfil con datos y capacidades profesionales verificables | No se habilita mediante un rol de software solamente |
| Odontólogo | Profesional de odontología elegible para las actividades que correspondan | Un rol llamado «odontólogo» no acredita título, colegiatura o habilitación |
| Rol | Agrupación de permisos de aplicación | Recepción puede incluir reservar y reprogramar |
| Permiso | Operación autorizable, junto con restricciones de recurso | `clinica.historia.leer` requiere evaluar también a qué paciente se accede |
| Profesional tratante | Quien asume o realiza una atención o intervención específica | Se conserva el autor real de cada atención |
| Profesional remitente | Quien deriva al paciente o recomienda la atención | Puede ser externo; no se reemplaza por el tratante |

## Paciente, agenda y clínica

| Término | Significado propuesto | Ejemplo o distinción |
|---|---|---|
| Paciente | Identidad asistencial y datos administrativos de quien recibe atención | Cambiar su teléfono no modifica evoluciones firmadas |
| Historia clínica | Conjunto organizado y trazable de información de salud de ese paciente | No es su ficha de contacto ni un único formulario mutable |
| Cita | Compromiso de tiempo con un odontólogo para un paciente | Existe aunque finalmente no se realice atención |
| Agenda | Vista de citas, horarios y bloqueos de un período | No requiere una entidad que contenga toda la vida laboral del profesional |
| Horario | Regla de trabajo del odontólogo, con vigencia | Lunes de 09:00 a 13:00 desde una fecha |
| Disponibilidad | Resultado calculado a partir de horario, excepciones, duración y reservas vigentes | Un hueco mostrado puede ocuparse antes de confirmar; el servidor revalida |
| Bloqueo | Excepción que impide reservar un intervalo | Ausencia o reunión; su motivo clínico/personal no tiene por qué ser público |
| Atención clínica | Encuentro asistencial efectivamente registrado | Puede existir sin cita; una llegada no prueba que la atención se completó |
| Procedimiento | Tipo de intervención clínica definida en un catálogo | Restauración es un tipo; la realizada hoy sobre una pieza es una ocurrencia |
| Procedimiento realizado | Registro de una ejecución concreta, autor, fecha y localización cuando corresponda | No es automáticamente un pago ni un cargo |
| Evolución | Relato y hallazgos de cómo progresa la situación clínica | «Evolución de TTO» del papel es una vista que puede incorporar varias entidades |
| Odontograma | Representación estructurada y fechada del estado y registros dentarios | Debe distinguir hallazgos, localización y convenciones aprobadas |
| Adenda | Corrección o ampliación vinculada a un registro finalizado, con autor, fecha y motivo | Conserva el registro previo para reconstruir qué se conocía en cada momento |
| Receta | Documento de una prescripción emitida por profesional autorizado | Puede relacionarse con una atención sin representar todo un tratamiento |

## Tratamiento, seguimiento y economía

| Término | Significado propuesto | Ejemplo o distinción |
|---|---|---|
| Plan de tratamiento | Propuesta clínica versionada de objetivos e intervenciones | Incluye alternativas y decisiones de aceptación; el importe se acuerda aparte |
| Tratamiento | Proceso clínico de ejecución y seguimiento de una intervención o conjunto coherente | Un tratamiento de ortodoncia dura meses y tiene múltiples atenciones |
| Control | Necesidad o encuentro de seguimiento, según se especifique | Usar `control recomendado` para la intención y `atención de control` para el encuentro |
| Control recomendado | Recomendación con fecha objetivo o ventana, propósito y estado | Se mantiene pendiente si se cancela la cita sin una nueva reserva |
| Precio | Valor monetario ofrecido o acordado por un concepto | La tarifa vigente es una referencia; el precio aceptado conserva su valor histórico |
| Importe del procedimiento | Valor económico correspondiente al procedimiento realizado | **Confirmado:** este es el sentido de «Costo» en el formato aportado |
| Presupuesto | Oferta económica versionada con conceptos, cantidades, moneda y condiciones | Puede corresponder a un plan; no sustituye su contenido clínico |
| Cargo u obligación de pago | Importe cuya exigibilidad se origina por una regla o acuerdo definido | La aceptación, ejecución o cuota puede ser el hecho generador; falta decidirlo |
| Cuenta por cobrar | Conjunto de obligaciones pendientes del deudor | El deudor puede ser un responsable distinto del paciente; validarlo |
| Pago | Dinero recibido, con importe, moneda, medio, fecha y responsable del registro | Puede ser parcial, anticipo o aplicarse a varios cargos |
| Abono | Pago parcial en el vocabulario del consultorio | No necesita ser una entidad adicional si cumple las mismas reglas de Pago |
| Aplicación de pago | Asignación de una parte de un pago a un cargo | Explica qué deuda disminuyó y admite un pago repartido |
| Saldo | Resultado de obligaciones, ajustes y aplicaciones válidas | No es un número que recepción pueda editar libremente |
| Comprobante tributario | Documento emitido según las reglas tributarias aplicables | Su ciclo de emisión y aceptación no es el ciclo del ingreso de dinero |
| Constancia interna | Evidencia interna de una operación | No presentarla como boleta, factura o recibo por honorarios válido |

## Comparaciones que gobiernan el modelo

| Diferencia solicitada | Consecuencia concreta |
|---|---|
| Usuario / personal / odontólogo / rol | Identidades y ciclos distintos; vínculo opcional y permisos explícitos |
| Paciente / historia clínica | Pacientes administra identificación; Clínica controla información de salud |
| Cita / agenda | Cita es un compromiso; agenda es una consulta organizada |
| Cita / atención | Reserva no prueba atención; atención sin reserva es posible |
| Control / procedimiento | Seguimiento puede no realizar intervención; un procedimiento puede ocurrir en una atención de control |
| Tratamiento / plan | Plan describe intención; tratamiento registra ejecución y continuidad |
| Tratamiento / presupuesto | Proceso clínico y acuerdo de importes cambian por razones distintas |
| Presupuesto / cuenta por cobrar | Ofertar o aceptar no define por sí solo cuándo nace una obligación |
| Pago / comprobante | Entrada de dinero y documento tributario tienen estados y fallos separados |
| Receta / tratamiento | Indicación farmacológica no equivale al proceso odontológico completo |
| Tratante / remitente | Se conservan ambos cuando existan; no se sobrescribe la derivación al cambiar de tratante |

## Convención de nombres

Propuesto: español sin tildes en identificadores de negocio (`Paciente`, `ReservarCita`, `paciente_id`, `procedimiento_realizado`), español natural en interfaz y documentación. Términos técnicos convencionales de Go/React permanecen en inglés (`context`, `http`, `main`, `components`). No alternar `patient` y `paciente` para el mismo contrato.

«Costo» no se usará como nombre técnico del dinero recibido; `Pago.importe` será diferente de `DetallePresupuesto.precioUnitario`. Evitar `TratamientoDTO` universal: cada operación publica solo los campos que su consumidor necesita.




