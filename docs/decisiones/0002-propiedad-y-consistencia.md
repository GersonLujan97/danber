# ADR 0002 — propiedad de datos y consistencia de reservas

Propuesto: 2026-10-14. Estado: **Aceptado provisionalmente el 2026-10-16**. Revisión de impacto: **Pendiente de revisión**.

CHG-002 en Notion: el usuario leyó la documentación y expresó «me parece por el momento correcto». Se registra conformidad con la orientación arquitectónica; las reglas abiertas y la validación técnica siguen pendientes. Se conserva abajo el razonamiento de la propuesta original.

## Contexto

El usuario confirma que se reserva únicamente al odontólogo. Dos recepcionistas pueden intentar ocupar intervalos solapados. CockroachDB está previsto como motor; compatibilidad SQL y tratamiento de reintentos deben verificarse en su versión concreta.

## Decisión propuesta

Una base con esquemas por propietario y migraciones en secuencia global. UUIDv4 como identidad interna; no reutilizar DNI ni número visible de historia como clave técnica. Claves foráneas internas por defecto y referencias cruzadas a identidades estables donde aporten integridad, coordinadas y sin permitir SQL ajeno.

Reservas ordinarias con intervalos semiabiertos y validación/guardado en una transacción serializable. Una fila de control por odontólogo participa mediante actualización en toda escritura de disponibilidad. Reintentar completamente errores serializables con límites; idempotencia para repetición de la misma intención. Reprogramación mantiene la cita anterior si no puede confirmar la nueva.

Base técnica propuesta: CockroachDB v26.2.7; fijar y probar antes de ejecutar. SQL y scripts se generan en D para el alcance del MVP, no en esta entrega de A–C.

## Alternativas

Esquema compartido: simple, propiedad menos visible. Bases separadas: aislamiento mayor con coordinación innecesaria hoy. Franjas discretas únicas: opción válida si se acepta granularidad fija; la duración base confirmada es 30 minutos y debe poder ajustarse (ADR 0003); una granularidad fija no está confirmada. Unicidad por hora inicial: insuficiente para intervalos que empiezan distinto y se solapan.

## Consecuencias

Un punto de conflicto por odontólogo simplifica la corrección, pero serializa cambios de días diferentes; refinar solo si las mediciones lo justifican. Esquemas no aíslan por sí mismos frente a un proceso con acceso amplio. Claves foráneas cruzadas acoplan cambios de esquema y deberán revisarse al separar servicios. Elegibilidad y baja del profesional requieren una política temporal explícita aparte de la exclusión de horarios.

## Evidencia necesaria antes de aceptar implementación

Pruebas con conexiones reales independientes: solapamiento exacto y parcial, adyacencia, bloqueo concurrente, reprogramación, idempotencia y reintentos. Migraciones desde cero y versión anterior en el motor fijado. Detalle y fuentes: [modelo preliminar](../base-datos/04-modelo-preliminar.md), [pruebas](../pruebas/07-estrategia.md).


