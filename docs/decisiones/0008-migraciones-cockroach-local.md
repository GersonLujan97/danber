# ADR 0008 — migraciones iniciales y ejecución local

Fecha: 2026-10-17. Estado: **Implementado para desarrollo y verificado en el alcance indicado**; las reglas nuevas de negocio mantienen estado Propuesto/Pendiente de validación. Revisión de impacto: **Pendiente de revisión**. CHG-008. Origen: el usuario solicita generar scripts de base y elige CockroachDB local.

## Antes y después

Antes: diseño técnico documental (ADR 0007), sin migraciones ejecutables. Ahora: siete migraciones, 28 tablas de negocio/técnicas más tabla de control de migración, esquemas Identidad/Personal/Pacientes/Agenda, SQL inicial exportado y scripts PowerShell/Python estándar para instalación/inicio/aplicación.

Version fijada: CockroachDB 26.2.7. Binario Windows oficial verificado con SHA256 b6a0b253b1f7853d6786b1423b507c56878a11411afeebb0c5c5ed2a36098aef. Windows está publicado como experimental; el alcance probado es desarrollo en un nodo. No se eligió alojamiento ni plataforma de producción.

## Decisiones técnicas

- DDL por archivo fuera de una transacción global; estado aplicando antes de DDL, aplicada al finalizar. SHA256 y nombre inmutables. Un error deja marcador y requiere revisar estado real; no se simula rollback global de DDL.
- Índices únicos parciales para documento principal, apoderado principal abierto, espera activa y una atención abierta por profesional. Requerir al menos un documento/apoderado/especialidad se completa en casos de uso; no con CHECK entre tablas.
- Control de profesional y protocolo SERIALIZABLE para reservas, historial y recibo atómicos. SQL de referencia verificado; INSERT directo no implementa ese protocolo.
- Campos creados_por/actualizado_por admiten NULL para aprovisionamiento técnico inicial; el backend debe exigir actor de sesión en toda operación de usuario. Historial de cambios de aplicación sí exige actor. No insertar actor ficticio en el arranque.
- Tipo documental: catálogo con DNI inicial; admitir otro tipo requiere configurarlo, no identificar todos como un mismo «OTRO». Normalización v1 upper/btrim; reglas adicionales por tipo se revisan al configurar.
- No imponer unicidad de paciente entre profesionales por decisión no confirmada. No sembrar personas, especialidades no validadas, horarios sabatinos ni contraseñas.
- Rol SQL danber_runtime separado de Maestro: permisos operativos sin DDL ni escritura del historial de migración; tablas de cambios/recibos sin UPDATE/DELETE. La autorización por operación/recurso corresponde al backend aún no implementado.
- Elegibilidad en la referencia se lee dentro del orden serializable; una baja debe gestionar futuras citas. No se declara probada la operación concurrente de baja ni otros casos de uso inexistentes.

## Alternativas

Un SQL monolítico sin historial facilita el primer pegado pero complica futuras actualizaciones. Se entrega como exportación reproducible con el mismo historial, junto al ejecutor recomendado. Borrar y recrear no es estrategia de migración.

Usar PostgreSQL/SQLite para validar sería más fácil hoy, pero no prueba sintaxis ni aislamiento de CockroachDB. Se usó nodo real aislado, solo localhost y memoria. No se desplegó la base persistente del usuario.

## Evidencia

37 comprobaciones correctas en CockroachDB CCL v26.2.7, Windows amd64, build 8652ac22d9f90501e48ae4f24efd7faab33cf7c5. Informe reproducible en [resultado de verificación](../../db/tests/resultado-verificacion.json); suite [verify.py](../../db/tests/verify.py). Incluye actualización desde versión 3, reejecución, checksums/estado incompleto, exportación SQL desde cero, restricciones, dos conexiones competidoras, idempotencia, almuerzo/cierre, bloqueos, rollback y privilegios SQL.

En la carrera observada, una reserva confirmó y la otra no obtuvo espacio; hubo **0 errores 40001 visibles al cliente**. La lógica de reintento completo está incluida, pero su rama no quedó demostrada por este resultado. No se afirma prueba de todo el catálogo de aceptación de ADR 0007 ni de API, RBAC por recurso, baja concurrente o carrera reserva/bloqueo.

## Impacto, pendientes y siguiente paso

Afecta portada y capítulos 03, 04, 05, 07 y diseño técnico; README/AGENTS; db completo. Migraciones y pruebas reales entregadas, revisión de impacto pendiente. No aprueba silenciosamente catálogos, contacto de menores, representantes, roles propuestos ni cierre sabatino.

Guía: [db/README](../../db/README.md). Próximo trabajo: revisar decisiones aplicables pendientes y construir casos de uso Go con autorización y protocolos transaccionales, después interfaz React. La base por sí sola no es la aplicación final.

Fuentes: [releases](https://docs.cockroachlabs.com/docs/releases/v26.2), [DDL](https://docs.cockroachlabs.com/docs/v26.2/online-schema-changes), [índices parciales](https://docs.cockroachlabs.com/docs/v26.2/partial-indexes), [reintentos](https://docs.cockroachlabs.com/docs/v26.2/transaction-retry-error-reference).


## Verificación editorial final

Guía y seis páginas relacionadas de Notion leídas de vuelta; copia SQL idéntica al archivo exportado. Nueve documentos locales: cero enlaces rotos. Python compila y los tres scripts PowerShell pasan análisis sintáctico. Revisión de impacto de negocio pendiente.
