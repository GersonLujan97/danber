# Contratos entre capacidades y ficha de módulo

**CHG-011 · 2026-10-17:** flujo de ramas adoptado e incremento feat/backend-inicial implementado: servidor Go, configuración, pool pgx, health/live y health/ready, cierre ordenado, contrato HTTP y CI. Integración local con CockroachDB real; nueva migración 008 restringe CREATE en public. Revisión de impacto pendiente. Cuenta Maestro, sesiones y capacidades de negocio aún no implementadas.

[Ejecutar el backend inicial](../../backend/README.md). Los estados anteriores se conservan como antecedentes.


**CHG-008 · 2026-10-17:** migraciones iniciales de CockroachDB 26.2.7 generadas y verificadas: siete versiones, 28 tablas más control de migración, SQL único y scripts de ejecución local. **37 comprobaciones correctas** en nodo aislado con datos ficticios. Autorización/API aún pendientes; las reglas de negocio no confirmadas conservan ese estado. Revisión de impacto pendiente.

[Guía para crear la base local](../../db/README.md). Los antecedentes siguientes conservan su estado histórico.


**CHG-007 · 2026-10-16:** diseño técnico del primer recorrido elaborado por instrucción del usuario. Diccionario con tipos/obligatoriedad, relaciones y restricciones; flujo/transiciones; matriz de permisos y casos de aceptación. Estado del diseño: **Propuesto; revisión de impacto pendiente**. No equivale a migraciones verificadas ni aplicación implementada. Los antecedentes siguientes conservan su historia; para concretar implementación consultar el diseño vigente.

[Leer el diseño técnico vigente](../base-datos/06-diseno-tecnico-pacientes-agenda.md).


Estado: **Propuesto**. Un contrato describe datos, significado, errores, autorización y consistencia. Que dos funciones tengan los mismos campos no demuestra que prometan lo mismo. Las entidades internas y los objetos SQL nunca son el contrato.

## 1. Contratos mínimos por relación

| Consumidor → proveedor | Operación propuesta | Datos mínimos y garantía | Errores relevantes |
|---|---|---|---|
| Agenda → Pacientes | `ObtenerPacienteParaCita(pacienteID)` | ID, nombre visible, estado administrativo y versión; lectura autoritativa al consultar | No encontrado, acceso denegado, no elegible según política, fallo temporal |
| Agenda → Personal | `ObtenerProfesionalParaAgenda(profesionalID)` | ID, nombre visible, perfil de odontólogo, estado y versión | No encontrado, no elegible, fallo temporal; la habilitación no se presume del rol |
| Agenda → Agenda | `ConsultarDisponibilidad(profesionalID, periodo, duracion)` | Intervalos disponibles calculados con horarios y bloqueos; informativos hasta confirmar reserva | Período inválido, duración no permitida, profesional no disponible |
| Clínica → Pacientes | `ObtenerIdentificacionClinica(pacienteID)` | Identificación y datos de filiación necesarios para el registro; procedencia y fecha | No encontrado, identificación incompleta según operación |
| Clínica → Personal | `ObtenerIdentidadProfesionalParaRegistro(profesionalID)` | Identidad, datos profesionales pertinentes y evidencia/fecha de validación | No autorizado para operación, no elegible, información pendiente |
| Coordinador de seguimiento → Tratamientos | `ObtenerControlPendiente(controlID)` | Paciente, objetivo, ventana/fecha, tratamiento opcional y versión | Ya cumplido/cancelado, versión conflictiva |
| Coordinador de seguimiento → Agenda | `ReservarControl(controlID, cita, clave)` | Reserva con correlación estable; resultado recuperable por clave | Conflicto horario, reserva existente, comando diferente con misma clave |
| Economía → Tratamientos | `ObtenerItemsPlan(versionPlanID)` | Ítems y referencias clínicas mínimas para cotizar; versión inmutable | Versión inexistente o sustituida según operación |
| Economía → Clínica | `ObtenerTiposProcedimiento(ids)` | Códigos, denominaciones y vigencia del catálogo publicado, sin información de pacientes | Código inexistente o retirado; no modificar ofertas históricas |
| Coordinador de atención → Tratamientos | `ValidarReferenciaTratamiento(tratamientoID, pacienteID)` | Confirma pertenencia y estado permitido antes de vincular una ejecución clínica | Paciente distinto, referencia inexistente o estado incompatible |
| Economía ← Clínica | `ProcedimientoRealizadoRegistrado.v1` cuando exista consumidor | ID de evento, procedimiento, paciente, tipo, fecha, referencia de tratamiento/plan si existe; sin nota clínica completa | Duplicado se reconoce sin repetir cargo; sin regla económica se deja pendiente de revisión |
| Facturación → Economía | `ObtenerBaseEmision(baseID, version)` | Receptor/deudor pertinente, conceptos, moneda, importes y referencias de origen aprobadas | Base no apta, versión obsoleta, emisión previa |
| Recetas → Clínica y Personal | Caso de uso clínico `EmitirReceta` | Atención/paciente, prescriptor elegible y contenido validado | Falta de autorización, atención incompatible, validación clínica/profesional pendiente |

Agenda es propietaria de disponibilidad. Consultar Personal no pregunta por un «horario libre»; pregunta quién es el profesional y si es elegible. Esta diferencia evita que dos módulos calculen huecos con reglas distintas.

La última fila no representa un contexto Recetas independiente: es una capacidad de Clínica que usa su propia atención y el contrato de Personal.

Toda asociación debe verificar pertenencia: una cita enlazada a una atención corresponde al mismo paciente; un procedimiento vinculado a tratamiento corresponde a su paciente; un plan y su presupuesto mantienen identidad consistente. La existencia de dos UUID no prueba esa concordancia. Los contratos y coordinadores validan estos cruces; dentro de un contexto se añaden restricciones compuestas cuando proceda. El autor y el prescriptor se vinculan a la sesión y a una delegación explícita, no se aceptan como una identidad arbitraria enviada por la interfaz.

## 2. Contrato preliminar de reserva

```text
ReservarCita
  Principal confiable: obtenido de sesión; nunca del cuerpo enviado por navegador
  pacienteID
  profesionalID
  inicio: instante con desplazamiento explícito
  fin: instante con desplazamiento explícito
  zona: America/Lima para el consultorio inicial
  controlRecomendadoID: opcional
  claveIdempotencia

Resultado
  citaID
  pacienteID
  profesionalID
  inicio, fin, zona
  estado: reservada
  version

Garantías
  Confirmación solo después del commit
  Mismo actor + operación + clave + contenido -> mismo efecto
  Misma clave con contenido distinto -> conflicto
  Reserva ordinaria sin solapamiento bajo el protocolo transaccional
  Resultado sin antecedentes, diagnósticos ni datos económicos del paciente
```

El contrato no garantiza que el paciente llegue o que el odontólogo siga activo eternamente. Tampoco garantiza éxito por haber consultado disponibilidad antes. Actualización CHG-007: duración predeterminada y mínima confirmada de 30 minutos; sin solapamientos por odontólogo ni sobreturnos simultáneos. Anticipación y reservas del paciente en paralelo siguen pendientes. Resultado propuesto vigente: pendiente más ocupa_intervalo=true, mostrado como Programada; ver diseño técnico.

Para reprogramar se añade `versionEsperada`, motivo y nuevo intervalo; actualizar una versión distinta devuelve conflicto de edición. La verificación de versiones protege cambios sobre una misma cita; el protocolo de intervalos protege conflictos entre citas diferentes.

## 3. HTTP preliminar

| Operación | Ruta propuesta | Resultado |
|---|---|---|
| Iniciar/cerrar sesión | `POST /api/v1/sesiones`, `DELETE /api/v1/sesion-actual` | Cookie segura o revocación; no exponer token de sesión en una respuesta destinada a almacenamiento local |
| Registrar paciente | `POST /api/v1/pacientes` | Identidad administrativa y versión |
| Buscar pacientes | `GET /api/v1/pacientes` | Página de resultados con campos mínimos y límite |
| Listar profesionales reservables | `GET /api/v1/profesionales-reservables` | Datos publicados para agenda |
| Consultar disponibilidad | `GET /api/v1/disponibilidad` | Huecos orientativos para período/duración |
| Reservar | `POST /api/v1/citas` | Cita confirmada en persistencia y clave de idempotencia gestionada |
| Consultar agenda | `GET /api/v1/agenda` | Citas y bloqueos autorizados por profesional/período |
| Reprogramar, posteriormente | `POST /api/v1/citas/{id}/reprogramaciones` | Nueva versión e historial |

Respuesta de error propuesta: `codigo`, mensaje comprensible, errores de campos cuando corresponda y `correlationId`. No retornar sentencias SQL, trazas, notas clínicas o secretos.

Mapeo inicial: `400` formato inválido; `401` sesión ausente; `403` operación no autorizada; `404` recurso inexistente o no visible según política; `409` horario ocupado, edición obsoleta o clave incompatible; `422` regla de negocio incumplida cuando se diferencie de formato; `503` servicio temporalmente incapaz de completar. Documentar una sola convención y probarla. Un error técnico no se convierte en «paciente inexistente».

## 4. Comunicación sin acoplar implementaciones

Propuesto: Agenda define un puerto con `ObtenerPacienteParaCita`; un adaptador consume la fachada publicada de Pacientes y traduce su resultado. Agenda puede probarse con un sustituto en memoria que cumple el puerto. No necesita iniciar Facturación ni consultar directamente `pacientes.paciente`.

Un coordinador de lectura compone agenda y nombres mediante consultas por lotes para evitar una llamada por fila. La autorización se aplica antes de componer. Si el nombre no puede recuperarse, no se sustituye una referencia inválida silenciosamente por un paciente vacío; se diferencia fallo técnico de registro no disponible. Para grandes consultas se puede añadir una proyección con retraso explícito.

Las identidades estables pueden cruzar límites; los datos mutables se leen por contrato. Los documentos finalizados conservan copias históricas identificadas, porque consultar el nombre o precio actual cambiaría retrospectivamente su representación. Una copia no se convierte en fuente maestra ni autoriza sincronización bidireccional improvisada.

## 5. Fallos y recuperación entre módulos

### Control recomendado y cita

Propuesta: Tratamientos guarda el control; un coordinador lee su versión y pide reserva a Agenda con `controlID` y clave estable. Agenda impide dos reservas activas para el mismo control dentro de su transacción. Si se guarda la cita pero falla la confirmación del vínculo en Tratamientos, el coordinador registra/retoma el paso pendiente, consulta por correlación y completa el vínculo. No repite una reserva con una clave nueva.

Al cancelar la cita se comunica el hecho de forma durable cuando esta integración exista. El control vuelve a pendiente salvo decisión clínica de cancelarlo o cumplirlo por otra atención. Se necesita detección de tareas pendientes y reconciliación; un callback en memoria no cubre una caída del proceso.

### Procedimiento y cargo

La realización clínica se confirma aunque falle el procesamiento económico. Economía recibe referencia del hecho y busca un acuerdo aplicable. La generación del cargo requiere una regla validada; no convierte todo evento clínico en deuda. Un origen económico único evita crear dos cargos por recepción repetida del mismo evento.

Si la obligación surge por cuota o aceptación, registrar ese origen, y no crear después otro cargo por la ejecución ya cubierta. El modelo conserva importe del procedimiento, cargo y pago como conceptos diferentes.

### Pago y comprobante

Registrar pago y emisión son acciones con estados independientes. Facturación conserva intención y correlación antes de llamar al proveedor; si hay timeout consulta el resultado antes de volver a emitir. Rechazo o desconocimiento queda visible para conciliación. Revertir un pago no implica borrar un comprobante ni emitir automáticamente una nota sin la regla tributaria adecuada.

### Auditoría y cambios críticos

El historial de una reserva se guarda en su misma transacción. Las lecturas clínicas necesitan una política definida de auditoría durable: si no es posible registrar el acceso, la propuesta es fallar de forma controlada en operaciones sensibles, con cualquier excepción asistencial explícita y autorizada. No se pretende que un log de consola sea evidencia suficiente.

## 6. Versionado y pruebas de contrato

Los contratos locales se comprueban por compilación y pruebas semánticas. Cambiar un campo opcional a obligatorio, un estado o el significado de «activo» puede romper consumidores aunque el código compile. Agregar campos no es automáticamente compatible si existen validadores estrictos o consumidores que esperan conjuntos cerrados.

Para HTTP se mantiene OpenAPI con ejemplos de éxito, conflictos y permisos; un chequeo de diferencias detecta cambios incompatibles. Los eventos durables llevan nombre/versión, `eventID`, origen, fecha y correlación. El consumidor registra eventos procesados y debe tolerar repetición; si el orden importa, usar versión del agregado y una política explícita de orden/reconciliación.

## 7. Plantilla para documentar cada módulo

Copiar esta estructura cuando se implemente una capacidad; los candidatos actuales están resumidos en [contextos](../dominio/02-contextos.md).

```markdown
# Nombre del módulo

Estado y responsables de validación:

## Propósito y límites
Problema, capacidades propias y capacidades externas.

## Lenguaje ubicuo
Términos y enlaces al glosario común.

## Reglas
ID, certeza, origen, ejemplo válido y ejemplo rechazado.

## Modelo
Entidades, objetos de valor, agregados y alcance transaccional.

## Casos de uso
Actor, entrada, permisos, precondiciones, resultado y errores.

## Contratos y dependencias
Proveedores, consumidores, campos mínimos, versiones y consistencia.

## Persistencia
Tablas propias, claves, índices, historiales y migraciones.

## Seguridad
Datos sensibles, acceso por recurso, auditoría y retención aplicable.

## Pruebas
Reglas, casos, contratos, integración y fallos esperados.

## Decisiones y pendientes
ADR, supuestos, consecuencias y criterios para revisarlos.
```




