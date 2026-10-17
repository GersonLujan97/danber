# ADR 0005 — identificación ampliada, apoderado y permisos de Agenda

Fecha: 2026-10-16. Origen: diccionario explícito del usuario y confirmación previa de gestión de horarios. Estado: **Confirmado** para campos y permisos expresos; propuestas técnicas pendientes. Revisión de impacto: **Pendiente de revisión**. Registro CHG-005 en Notion.

## Cambio y alcance

Se aceptan los campos obligatorios/opcionales del [diccionario funcional](../base-datos/05-diccionario-pacientes-agenda.md). DNI u otro documento sustituye DNI exclusivo para pacientes; el apoderado requiere nombres/apellidos, parentesco y DNI u otro documento. Se mantiene el umbral confirmado de 66 años.

Recepción crea citas, registra llegada, cancela, ajusta duración y reasigna por demora. El odontólogo define sus horarios y crea/retira sus bloqueos. El odontólogo actual es usuario maestro, con acceso a todas las funciones y rol de recepción. La interpretación concreta de reasignar sigue pendiente.

## Relación con decisiones anteriores

ADR 0003: **sustituido parcialmente** en DNI exclusivo y datos mínimos del apoderado por este ADR; conserva duración, exclusión por odontólogo y umbral de edad.
ADR 0004: **complementado**; el permiso del odontólogo para gestionar sus horarios/bloqueos queda confirmado. El flujo de bloqueos con citas existentes sigue propuesto.

## Alternativas y consecuencias

Mantener solo DNI contradice la nueva instrucción. Registrar documento del paciente y apoderado como uno solo confunde identidades. Se proponen tipo/número separados, vínculo de parentesco y claves técnicas independientes.

Propuesto: asignar Maestro explícitamente a la cuenta del odontólogo actual, en vez de convertir la profesión en acceso universal para todos los futuros profesionales. Esto permite incorporar personal con permisos distintos sin duplicar cuentas. Acceso total a funciones no permite infringir reglas de reserva ni eliminar trazabilidad.

## Revisión requerida

Capítulos 01, 03 y 07 y portada; diccionario, contratos, permisos, validaciones, documentos históricos y pruebas futuras. Detalle en el diccionario funcional. Faltan estados y responsables de la cita, alcance de reasignación, tipos de documento, contacto de menores, especialidades y cierre sabatino. No hay código ni migraciones ejecutables modificados.

Responsable de validación de negocio: usuario. Validación técnica: por realizar al implementar. No se afirma cumplimiento normativo ni pruebas ejecutadas.

