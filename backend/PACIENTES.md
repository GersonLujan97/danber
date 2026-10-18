# Pacientes · Registro, búsqueda y pruebas locales

CHG-015 · 2026-10-18. Implementación autorizada por Isai, verificada localmente en CockroachDB 26.2.0 y 26.2.7. Documentación principal: [Pacientes en Notion](https://app.notion.com/p/3fd2756ebb4481c79e62cc2616020ffb). Revisión de impacto: **Pendiente de revisión**. Decisiones y alternativas: [ADR 0015](../docs/decisiones/0015-registro-y-busqueda-pacientes.md). Este incremento es una API; todavía no hay formulario React.

## Qué permite

Registrar una ficha administrativa con documento principal, contacto, apoderado cuando corresponde y auditoría. Buscar por tipo y documento exactos y consultar ficha por UUID. El usuario autenticado necesita `pacientes.crear` para el alta y `pacientes.leer` para consulta. Los roles actuales reciben estos permisos según el catálogo; el nombre del rol no concede un bypass.

En el único establecimiento, la lectura administrativa abarca todos los pacientes; no da acceso a historia clínica, pagos ni funciones de Agenda. El servidor comprueba permiso antes de consultar existencia de la ficha. Nuevos establecimientos o restricciones por profesional requieren revisar esta política.

| Operación | Ruta | Resultado |
|---|---|---|
| Alta | POST /pacientes | 201 con id y version inicial 1; Location de ficha |
| Buscar documento | POST /pacientes/buscar | 200 con id, nombres, apellidos, estado y documento; 404 sin coincidencia |
| Ficha administrativa | GET /pacientes/{id} | 200 con campos administrativos y apoderado principal vigente |
| Catálogo | GET /pacientes/tipos-documento | Tipos activos; requiere leer o crear |

No existe listado ni búsqueda por nombres todavía. Editar/baja/corregir documentos, fusionar y administrar apoderados quedan para otro incremento; nunca se sobrescriben nombres de apoderados durante un alta nueva.

## Validaciones

- Obligatorios: documento, nombres, apellidos, fecha de nacimiento, dirección y teléfono. Opcionales: estado civil, ocupación y correo. Recortar exteriores, rechazar vacío/control y respetar límites de las migraciones. Nombres con tildes o apóstrofos son válidos. No forzar un teléfono móvil peruano; puede ser el contacto del apoderado.
- DNI: texto de ocho dígitos ASCII, conserva ceros. Catálogo inicial solo DNI. Otros tipos deben estar habilitados; número 1–40 caracteres, sin controles y upper/trim. No se implementa verificación oficial ni formatos específicos aún: **Pendientes de validación**. No crear un código genérico OTRO sin describir el tipo.
- Fecha civil YYYY-MM-DD, válida y no futura al alta. Edad evaluada en America/Lima. Menor se implementa como <18 (criterio operativo propuesto); desde 66 requiere apoderado (confirmado). A los 65 no se exige por edad. Para 29/02 se cumple año el 01/03 si no hay 29/02: **propuesta operativa pendiente de validar**, sin interpretación legal.
- Apoderado: nombres, apellidos, parentesco y documento completos; no puede usar el mismo tipo/número del paciente. Se puede informar en otros adultos. Un documento de apoderado ya existente reutiliza su registro solo con nombres/apellidos idénticos tras recortar. Discrepancia: 409 `apoderado_datos_difieren`; revisar ficha con el responsable, sin alterar datos silenciosamente. No concede representación legal, firma ni acceso por sí solo.
- El cumpleaños no cancela citas. Volver a evaluar identificación al reservar/llegar corresponde a la futura Agenda. No almacenar una edad que envejezca en la tabla.

## Antes de probar

Integrar primero [PR #3 de autenticación](https://github.com/GersonLujan97/danber/pull/3), pendiente al inicio de este trabajo, y después el PR de pacientes. La rama de pacientes nace de feat/autenticacion; se explica la dependencia en el PR. No hay merge automático.

Con tu nodo local ya iniciado y la CLI en la ruta correcta, desde la raíz del repositorio aplica pendientes en una terminal para migraciones:

```powershell
$env:COCKROACH_BIN = 'C:\ruta\a\cockroach.exe'
$env:DANBER_DATABASE_URL = 'postgresql://root@localhost:26257/defaultdb?sslmode=disable'
python db/scripts/migrate.py apply --database danber
python db/scripts/migrate.py status --database danber
```

La nueva 011 solo agrega la tabla técnica de idempotencia y sus permisos; no recrea la cuenta ni altera migraciones previas. Si tu base ya existe, no repetir danber-inicial.sql: ese archivo se genera para instalaciones nuevas. No borrar la BD para actualizar. El backend detecta ausencia de 011 y no abre el servidor.

En otra terminal, el proceso de API usa la identidad SQL restringida:

```powershell
cd C:\GERSON\ISAI\DANBER\CODE\backend
$env:DANBER_DATABASE_URL = 'postgresql://danber_api@localhost:26257/danber?sslmode=disable'
$env:DANBER_HTTP_ADDR = '127.0.0.1:8090'
$env:DANBER_ALLOW_INSECURE_AUTH = 'true'
go run ./cmd/api
```

Esta excepción permite probar HTTP exclusivamente en loopback. Mantener la misma dirección exacta en todas las solicitudes. Antes de exposición real se necesitan HTTPS y los controles de despliegue pendientes de la [guía de seguridad](AUTENTICACION-Y-SEGURIDAD.md).

## Ejemplo PowerShell con datos ficticios

Primero inicia sesión con tu cuenta existente, sin ejecutar todavía el ejemplo de cierre, siguiendo [Acceso y seguridad](AUTENTICACION-Y-SEGURIDAD.md). Conserva las variables `$sesionDanber` (WebRequestSession con cookie) y `$acceso` (respuesta login con csrf_token). No volver a ejecutar crear-maestro. El secreto, cookie y token CSRF no deben imprimirse ni copiarse a Notion/Git.

```powershell
$cabeceras = @{ 'X-Danber-Client' = '1'; 'X-Danber-CSRF' = $acceso.csrf_token }
Invoke-RestMethod http://127.0.0.1:8090/pacientes/tipos-documento -WebSession $sesionDanber

$datosPaciente = @{
  nombres = 'Lucía'
  apellidos = 'Paciente Ficticia'
  fecha_nacimiento = '1990-01-01'
  direccion = 'Dirección ficticia de prueba'
  telefono = '111111111'
  documento = @{ tipo = 'DNI'; numero = '00123456' }
} | ConvertTo-Json -Depth 4 -Compress

# Generar UNA clave para esta intención de alta. Conservar cuerpo y clave al reintentar.
$claveAlta = [guid]::NewGuid().ToString()
$cabecerasAlta = @{
  'X-Danber-Client' = '1'
  'X-Danber-CSRF' = $acceso.csrf_token
  'Idempotency-Key' = $claveAlta
}
$alta = Invoke-RestMethod http://127.0.0.1:8090/pacientes -Method Post `
  -ContentType 'application/json; charset=utf-8' -Headers $cabecerasAlta `
  -WebSession $sesionDanber -Body ([Text.Encoding]::UTF8.GetBytes($datosPaciente))
$alta  # Solo UUID y versión, no secretos de sesión.

$documentoBuscar = @{ tipo = 'DNI'; numero = '00123456' } | ConvertTo-Json -Compress
$encontrado = Invoke-RestMethod http://127.0.0.1:8090/pacientes/buscar -Method Post `
  -ContentType 'application/json' -Headers $cabeceras -WebSession $sesionDanber -Body $documentoBuscar
Invoke-RestMethod "http://127.0.0.1:8090/pacientes/$($encontrado.id)" -WebSession $sesionDanber
```

Para un menor o paciente desde 66 agrega antes de serializar:

```json
"apoderado": {
  "nombres": "María",
  "apellidos": "Apoderada Ficticia",
  "parentesco": "Madre",
  "documento": { "tipo": "DNI", "numero": "00234567" }
}
```

Los ejemplos son ficticios. No usar documentos de pacientes reales en pruebas ni publicar capturas con sus datos.

## Reintentos, concurrencia y auditoría

**Idempotencia** permite repetir una solicitud sin repetir su efecto. Si la red falla después de registrar a Lucía, conserva el mismo Idempotency-Key y cuerpo: recibirás el mismo ID y versión inicial con 201, sin otra ficha ni auditoría. Cambiar cuerpo manteniendo clave devuelve 409 `clave_incompatible`; para corregir una ficha ya creada hará falta edición, no un alta nueva. Cada cuenta tiene su propio espacio de claves.

La huella SHA-256 se calcula sobre datos normalizados, sin incluir fecha actual; un cumpleaños posterior no altera un resultado ya confirmado. Cada reintento sigue exigiendo sesión y permiso actuales. Las claves no se borran ni caducan automáticamente: retención pendiente. Su huella no es anonimización; protegerla como dato técnico sensible. El resultado inicial no reemplaza la consulta de ficha actual.

**Transacción serializable** confirma ficha/documento/apoderado/vínculo/auditoría/clave juntos. La tabla de claves permite un resultado temporal NULL solo dentro de la transacción en curso; un NULL previamente confirmado se rechaza para revisión. Ante conflicto serializable se repite todo hasta cinco intentos; un commit de resultado desconocido no se repite internamente. Reintenta con la misma clave para recuperar el resultado. Si dos personas registran simultáneamente el mismo documento con claves distintas, el índice único y el caso de uso permiten una ficha y un 409. Consultar antes, por sí solo, no garantiza esto.

La auditoría registra actor, operación, momento y snapshot del alta en `pacientes.cambio_paciente`, dentro de SQL; no en logs HTTP. Si falla la auditoría, se deshace todo. La cuenta SQL de API no puede borrar historial ni claves, pero no es un almacén inmutable frente al administrador; copias y controles de despliegue siguen pendientes.

## Errores y seguridad práctica

| Estado | Significado | Qué hacer |
|---|---|---|
| 400 | JSON inválido, claves duplicadas/desconocidas, tipos incorrectos | Corregir estructura; no enviar roles, estado ni actor desde cliente |
| 401 | Falta sesión, expiró o fue revocada | Iniciar sesión; conservar intención/clave si hubo respuesta incierta |
| 403 | Falta permiso, CSRF o falla origen | Usar cuenta autorizada, CSRF actual y mismo host/origen |
| 404 | Sin coincidencia/ficha | Ofrecer alta después de verificar documento |
| 409 | Documento duplicado, apoderado distinto o clave incompatible | Buscar con permiso de lectura; resolver discrepancia sin duplicar |
| 413 / 415 | Más de 16 KiB / no application/json | Ajustar solicitud |
| 422 | Error de negocio/campo, incluye campo y código sin su valor | Completar obligatorios, fecha/documento/apoderado |
| 426 | HTTP sin excepción local explícita | Configurar entorno local o HTTPS al desplegar |
| 429 | Límite del módulo excedido | Respetar Retry-After; no reintentar en bucle |
| 503 | No se pudo confirmar operación | Conservar clave y cuerpo; revisar servicio y reintentar sin crear otra intención |

Ejemplos: cambiar el UUID de la URL no evita la comprobación del permiso; sin permiso se recibe 403 antes de consultar existencia. Una web ajena no puede ordenar un alta usando solo tu cookie: se comprueban origen/CSRF. Un texto con apóstrofo es un dato: se envía como parámetro SQL, sin concatenación. Una automatización que recorre documentos comparte un presupuesto de solicitudes del módulo, aunque alterne rutas. Estos controles reducen riesgos, no eliminan el abuso de una cuenta legítima ni acreditan una auditoría completa.

Límite técnico inicial: 120 solicitudes por IP y 600 globales por minuto para TODO Pacientes, contadores acotados a 4096 claves; memoria por proceso y reinicio reinicia contadores. Sin confiar cabeceras de proxy. Leer/alta requieren sesión activa y permiso recargado; no hay CORS general. JSON estricto, límite 16 KiB y contexto de cinco segundos; respuestas no-store. No hay logs de cuerpo/documento/dirección ni telemetría clínica. Frontend futuro deberá escapar datos para evitar XSS; validación de entrada no reemplaza ese control.

## Pruebas y próximos pasos

Evidencia final de esta entrega se registra en `tests/resultado-pacientes.json` y ADR 0015. Integración usa CockroachDB real, nodos en memoria y bases nuevas con datos ficticios; nunca la base del usuario. CI verifica Go con detector de carreras en Linux; integración SQL local por separado.

Verificado localmente: once subcasos de pacientes y actualización desde 010 en cada motor; regresión completa Go en 26.2.0, 38 comprobaciones SQL con once migraciones/SQL único, vet/build/mod verify y recorrido PowerShell correctos. Contrato revisado como YAML con referencias internas, sin validador semántico integral. El detector de carreras local requiere CGO, no disponible en este entorno Windows; su resultado en Linux se registra en la entrega remota. Estas pruebas no cierran las revisiones de impacto ni autorizan uso real.

Después: perfil profesional del odontólogo y contratos operativos antes de reserva/agenda. Edición administrativa tendrá versión esperada y auditoría propia; no inventar cambios con un alta. Pendientes: validación del consultorio indicada en ADR, interfaz y controles antes de uso real de ADR 0014.

Fuentes consultadas 2026-10-18: [OWASP autorización](https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html), [validación de entrada](https://cheatsheetseries.owasp.org/cheatsheets/Input_Validation_Cheat_Sheet.html).
