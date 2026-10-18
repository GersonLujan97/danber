# Acceso y seguridad de Danber

CHG-014 · 2026-10-18. Rama `feat/autenticacion`, desde main c0fb8ca con PR #2 integrado. Alcance autorizado por Isai: continuar con inicio de sesión e incorporar seguridad desde cada etapa, explicando motivos y casos prácticos. **Implementado y verificado localmente; revisión de impacto pendiente.** Los parámetros de esta guía son valores técnicos iniciales revisables, no reglas clínicas confirmadas.

El usuario informó que creó correctamente su cuenta con Maestro y Recepción. Ese resultado es evidencia de ejecución comunicada por él; no se inspecciona su contraseña ni se vuelve a crear su cuenta. Las pruebas del incremento usan cuentas y bases ficticias aisladas.

## La seguridad se construye durante el proyecto

No es una carpeta que vuelva segura a toda la aplicación ni una última tarea antes de publicar. Cada capacidad incorpora sus controles y sus pruebas. El despliegue añade otros controles de red y operación. Por ejemplo, limitar login no evita que una futura operación de Pacientes permita leer una ficha ajena: esa operación necesitará comprobar el permiso y el recurso concreto.

**Autenticación:** comprobar quién entra. **Autorización:** decidir qué puede hacer y sobre qué datos. **Sesión:** recordar un acceso autenticado durante un tiempo limitado, sin volver a enviar la contraseña en cada operación. Alternativa de enviar siempre la contraseña: aumenta exposición y dificulta revocar un dispositivo. Ejemplo: recepción inicia sesión una vez y utiliza una cookie para trabajar; al salir se revoca esa sesión en el servidor.

La identidad SQL `danber_api`, el usuario de aplicación, sus roles y su perfil profesional siguen separados. La sesión no crea un odontólogo, no otorga permisos enviados por el cliente y no confirma habilitación profesional. El catálogo actual de permisos no resuelve todavía el acceso a recursos clínicos.

## Qué incorpora esta entrega

| Operación | Uso | Protección |
|---|---|---|
| POST /auth/login | Verificar usuario/contraseña y abrir una sesión nueva | JSON estricto, límites de intentos y CPU, mensaje genérico, cookie nueva |
| GET /auth/session | Consultar identidad, permisos actuales, vencimiento y token CSRF | Cookie válida, cuenta activa, versión, contraseña vigente y ambos vencimientos |
| POST /auth/logout | Cerrar la sesión de ese navegador | Cookie y token CSRF; revocación y auditoría atómicas; repetición sin duplicar historial |
| POST /auth/logout-all | Cerrar todas las sesiones de la misma cuenta | Sesión vigente, CSRF, cambio de versión y revocación en transacción |

La respuesta de login/session contiene `principal`, `expira_en` y `csrf_token`. El token de acceso se entrega solamente en `Set-Cookie`, no en ese JSON. La cuenta inicial no debe crearse otra vez. No hay todavía pantalla React de acceso ni endpoints de gestión de usuarios, cambio de contraseña, recuperación, Pacientes o Agenda.

## Controles, ejemplos y límites

| Control y estado en CHG-014 | Por qué y ejemplo práctico | Alcance y siguiente revisión |
|---|---|---|
| Argon2id y verificación acotada; implementado | Si alguien obtiene una copia de credenciales, no encuentra la contraseña legible. Un hash lento encarece probar un diccionario. Solo se aceptan parámetros PHC conocidos para evitar una asignación arbitraria de memoria | Mantiene política CHG-013; medir costo en despliegue. MFA y comprobación de contraseñas comprometidas siguen pendientes |
| Respuesta genérica y hash ficticio para cuenta ausente; implementado | Probar `Isai` o un usuario inexistente devuelve el mismo 401. No se informa «ese usuario sí existe»; ambos casos hacen trabajo de verificación | Reduce enumeración; no se afirma igualdad perfecta de tiempos de red/base/CPU |
| Rate limit por IP real y nombre normalizado; implementado | Un programa que prueba miles de contraseñas recibe 429 y Retry-After. Cambiar X-Forwarded-For no cambia la identidad de origen usada por el backend | Memoria por proceso y ventana fija; reiniciar la API reinicia contadores. Varias instancias requerirán coordinación compartida antes de desplegarlas. Un tercero también puede agotar un presupuesto y causar esperas a usuarios legítimos; revisar ajustes y MFA según exposición |
| Dos verificaciones costosas simultáneas; implementado | Solicitudes paralelas no disparan decenas de derivaciones Argon2id a la vez, limitando consumo de memoria/CPU | Protege este proceso. Una defensa de volumen/DDoS exige también límites de red/proxy en despliegue |
| Sesión aleatoria de 256 bits, hash en SQL y rotación; implementado | Adivinar números consecutivos no sirve para obtener una sesión. Una copia de la tabla no contiene el token reutilizable. Reautenticar la misma cuenta cierra la cookie anterior presentada por ese navegador | Una cookie válida robada sigue siendo un secreto de acceso; revocación, expiración, HTTPS y futuras protecciones de navegador complementan el control |
| Cookie HttpOnly, SameSite=Strict, sin Domain y Secure con TLS; implementado | Un script del frontend no puede leer directamente la cookie HttpOnly. SameSite y el origen estricto reducen envíos desde otra página | XSS puede operar dentro del origen aunque no lea la cookie. El frontend deberá prevenir XSS y ajustar su CSP; no basta HttpOnly |
| Origen y Host exactos, Fetch Metadata, JSON y cabecera de cliente; implementado | Una página maliciosa no puede enviar un formulario al login y cambiar silenciosamente la cuenta usada por recepción. También se rechaza Host de otro sitio conectado a loopback | No se habilita CORS general ni se confían proxies. La futura interfaz deberá compartir origen, mediante una configuración explícita |
| CSRF ligado a la sesión para cierres; implementado | Una página ajena no puede cerrar todas las sesiones mediante un formulario: le falta X-Danber-CSRF correcto y no supera el origen permitido | El token CSRF no sustituye la cookie ni autoriza una cuenta diferente. Los futuros endpoints de escritura deberán aplicar la misma protección |
| Expiración e inactividad comprobadas en SQL; implementado | Si recepción deja una sesión sin actividad, deja de autorizar después del límite de inactividad. Cambiar el reloj del navegador no amplía el acceso | El backend mide actividad por solicitudes autenticadas. El futuro frontend deberá evitar que sondeos automáticos simulen actividad humana |
| Revocación, versión de cuenta y permisos actuales; implementado | Cerrar todas las sesiones invalida las cookies anteriores. Retirar un rol elimina sus permisos en la siguiente consulta autenticada | Una solicitud en curso observa su transacción. Los futuros cambios de estado/roles/contraseña deben seguir el protocolo y cada operación de negocio comprobar recurso y permiso |
| Cuerpo de credenciales limitado, JSON estricto, SQL parametrizado y errores genéricos; implementado | Se rechazan cuerpos enormes, campos repetidos o un campo `rol: maestro`. Un nombre con sintaxis SQL se trata como dato | Los módulos futuros necesitan sus propias validaciones y pruebas. No se afirma inmunidad general a inyección o abuso |
| Auditoría atómica y respuestas sin caché; implementado | Si falla el registro de apertura/cierre, también se revierte la sesión. Cache-Control no-store evita indicar que estas respuestas pueden almacenarse en cachés compartidas | No se registran contraseñas, cookies, CSRF ni hashes en auditoría. Retención, monitoreo de incidentes y alertas de producción siguen por definir |

## Parámetros iniciales y decisiones revisables

- Login: hasta 10 solicitudes por IP/minuto y 60 por proceso/minuto; además, hasta 5 intentos por nombre de acceso normalizado/minuto. Se cuentan intentos correctos e incorrectos. Un rechazo 429 incluye cuánto esperar; no suspende permanentemente la cuenta.
- Consultar/cerrar sesiones: 120 solicitudes por IP/minuto, 600 por proceso/minuto. Health no consume estos presupuestos de autenticación.
- Cada limitador tiene como máximo 4096 claves. Al llenarse, elimina entradas vencidas y rechaza claves nuevas si aún no hay lugar; no expulsa claves activas para permitir eludir límites. Hay presupuesto global y un máximo de dos verificaciones de contraseña simultáneas. Las ventanas fijas permiten ráfagas cerca de un límite de minuto; revisar si la exposición futura necesita otra política.
- Los límites de intentos están versionados en el código y se ajustan mediante cambio revisado. No se añaden Redis ni infraestructura distribuida para el proceso local actual. La IP viene de RemoteAddr; las cabeceras Forwarded/X-Forwarded-For no son confiables en esta entrega.
- Sesión: **15 minutos de inactividad** y **8 horas de duración absoluta**. La actividad no prolonga la duración absoluta. Configuración: DANBER_SESSION_IDLE y DANBER_SESSION_TTL. Límites admitidos: inactividad de 1m a 1h, duración de 30m a 24h, y la primera no supera la segunda. Son valores iniciales sujetos a revisión con la operación del consultorio.
- Tiempo máximo de cada operación HTTP de autenticación: 5 segundos. Cuerpo de credenciales: 4096 bytes; no campos desconocidos, repetidos ni JSON adicional. Los límites HTTP de cabeceras/lectura/escritura del backend siguen vigentes.
- Los fallos de usuario inexistente, clave equivocada o cuenta suspendida devuelven el mismo 401. Los eventos anónimos fallidos se registran de forma genérica en el log, sin inventar un usuario/actor de auditoría. Las aperturas y revocaciones autenticadas se registran en cambio_acceso con usuario y sesión, sin secretos.
- El hash SHA-256 del token de sesión es adecuado porque el token tiene 256 bits aleatorios; no se aplica ese razonamiento a contraseñas humanas. Para estas se conserva Argon2id. El token CSRF usa HMAC-SHA-256 ligado al secreto de sesión y comparación segura; no otorga acceso por sí solo.

## Ejecutar en tu equipo, sin volver a crear la cuenta

Después de integrar el PR y actualizar main, mantén tu CockroachDB local iniciado. Desde la raíz, con conexión migradora:

```powershell
cd C:\GERSON\ISAI\DANBER\CODE
# Si necesitas indicar tu ejecutable: $env:COCKROACH_BIN = 'C:\ruta\cockroach.exe'
$env:DANBER_DATABASE_URL = 'postgresql://root@localhost:26257/defaultdb?sslmode=disable'
python db/scripts/migrate.py apply --database danber
```

La última versión será **010_sesiones_y_version_acceso.sql**. Conserva las nueve anteriores. Agrega último uso y versión de cuenta a las sesiones y un índice de revocación; no borra usuarios, contraseñas ni datos existentes. El SQL inicial exportado es solo para bases nuevas.

En otra terminal:

```powershell
cd C:\GERSON\ISAI\DANBER\CODE\backend
$env:DANBER_DATABASE_URL = 'postgresql://danber_api@localhost:26257/danber?sslmode=disable'
$env:DANBER_HTTP_ADDR = '127.0.0.1:8090'
$env:DANBER_ALLOW_INSECURE_AUTH = 'true'
# Opcionales: $env:DANBER_SESSION_IDLE='15m'; $env:DANBER_SESSION_TTL='8h'
go run ./cmd/api
```

**La excepción HTTP es explícita y exclusiva de desarrollo en loopback.** Sin ella, las rutas de autenticación por HTTP devuelven 426; health continúa respondiendo. La API comprueba las columnas de 010 al arrancar, sin ejecutar migraciones automáticamente. Usa exactamente `http://127.0.0.1:8090` para estas pruebas: Host y origen deben coincidir con DANBER_HTTP_ADDR. `localhost` y `127.0.0.1` son orígenes diferentes.

Este backend todavía no incluye despliegue público ni terminación TLS. El adaptador exige HTTPS por defecto y marca la cookie Secure cuando recibe TLS; la excepción local usa cookie sin Secure. No habilitar esta excepción en un despliegue remoto ni confundir `--insecure` de CockroachDB con una instalación segura. Antes de uso real se prepararán TLS, certificados y las demás condiciones pendientes.

## Ejemplo de acceso desde PowerShell

En otra terminal, utiliza tu nombre de acceso real; el ejemplo usa Isai. La contraseña se pide sin mostrarla. Se convierte temporalmente a texto en memoria para enviarla; no se incluye en el comando ni se guarda en un archivo. No imprimas `$cuerpo`, la cookie ni el token CSRF.

```powershell
$clave = Read-Host 'Contraseña de Danber' -AsSecureString
$credencial = [System.Net.NetworkCredential]::new('', $clave)
$cuerpo = @{ usuario = 'Isai'; contrasena = $credencial.Password } | ConvertTo-Json
try {
    $acceso = Invoke-RestMethod -Method Post `
        -Uri 'http://127.0.0.1:8090/auth/login' `
        -ContentType 'application/json' -Body $cuerpo `
        -Headers @{ 'X-Danber-Client' = '1' } -SessionVariable sesionDanber
} finally {
    Remove-Variable cuerpo, credencial, clave -ErrorAction SilentlyContinue
}
$acceso.principal
```

El resultado muestra tu usuario, roles y permisos. La cookie queda en `sesionDanber`, en memoria. Eliminar variables evita conservar referencias deliberadamente; no promete borrado completo de memoria administrada. No vuelvas a ejecutar crear-maestro para resolver un fallo de login.

```powershell
# Consultar la sesión y sus permisos actuales:
$actual = Invoke-RestMethod -Uri 'http://127.0.0.1:8090/auth/session' -WebSession $sesionDanber
$actual.principal

# Cerrar esta sesión:
Invoke-RestMethod -Method Post -Uri 'http://127.0.0.1:8090/auth/logout' `
    -WebSession $sesionDanber `
    -Headers @{ 'X-Danber-Client' = '1'; 'X-Danber-CSRF' = $actual.csrf_token }
```

Para cerrar **todas** las sesiones de la misma cuenta, usa `/auth/logout-all` con una sesión aún vigente y su CSRF. Eso también cierra la sesión que solicita la operación. Necesitarás iniciar sesión nuevamente. Cerrar una cookie en el cliente sin revocarla en el servidor no basta para invalidar una copia robada.

| Respuesta | Significado y acción |
|---|---|
| 200 | Login o consulta correctos |
| 204 | Cierre confirmado; sin cuerpo |
| 400 / 413 / 415 | Datos, tamaño o Content-Type no admitidos |
| 401 | Credenciales o sesión no válidas; no distingue existencia/estado de cuenta |
| 403 | Origen, Host, cabecera de cliente o CSRF no admitido |
| 426 | Se requiere HTTPS o la excepción local explícita |
| 429 | Se excedió un presupuesto; respetar Retry-After |
| 503 | No se pudo confirmar la operación; no se expone el error SQL ni se declara éxito |

## Protocolo que deben respetar los siguientes módulos

Cada operación de negocio deberá autenticar y autorizar permiso + recurso. Principal.Allows solo comprueba el catálogo de capacidades; no decide relación asistencial ni propiedad del horario. Ejemplo: poder gestionar bloqueos no permite bloquear el horario de otro odontólogo sin una regla expresamente autorizada.

Suspender/reactivar cuentas debe incrementar usuario.version y revocar sesiones en la misma transacción; cambiar una contraseña debe actualizar credencial.cambiado_en y revocar las sesiones pertinentes. La lectura de sesión comprueba ambos mecanismos. Esas pantallas/endpoints de administración no se implementan todavía: no usar SQL manual como interfaz normal del consultorio.

El repositorio conserva SQL parametrizado, separación de dominio/adaptadores, usuario SQL sin administración, migraciones inmutables y comprobaciones de versión de CockroachDB. Las sesiones y su auditoría se guardan/revocan en transacciones serializables; reintentos acotados solo ante 40001. Un resultado incierto de COMMIT no se reproduce automáticamente.

## Controles previstos para las siguientes etapas

| Momento | Trabajo pendiente y caso práctico |
|---|---|
| Cada capacidad de negocio | Autorización por operación y recurso, validaciones propias, conflictos e idempotencia. Impedir acceder a la ficha de otro paciente cambiando un ID |
| Interfaz React | Formularios y salida segura contra XSS, política CSP adecuada al frontend, flujo de CSRF y expiración, manejo de caché/datos clínicos. Un nombre de paciente no debe ejecutarse como HTML |
| Administración de acceso | Cambios de rol/estado y contraseña con revocación, recuperación por canal verificado, revisión de contraseñas comprometidas y MFA para acceso privilegiado/remoto. Mitigar el uso de una contraseña robada, que un rate limit no detecta por sí solo |
| Antes de publicar o usar datos reales | HTTPS y TLS SQL, secretos y usuarios de despliegue, proxy confiable y sus límites, plataforma, actualizaciones y análisis de dependencias, copias cifradas y restauración probada, retención, monitoreo e incidentes. Poder recuperar el consultorio ante pérdida de la máquina |
| Antes de varias instancias | Rate limit coordinado y pruebas de concurrencia entre procesos; los límites locales no se suman como si fueran uno |
| Según exposición y evidencia | Protección de borde/WAF, detección de abuso y controles adicionales medidos; no se añaden servicios por nombre sin una necesidad |

Estas etapas requieren revisión de impacto y validación operativa/normativa aplicable. Ningún conjunto de controles garantiza ausencia de vulnerabilidades ni acredita cumplimiento legal. Se conserva la revisión normativa inicial, con sus fuentes y fechas, sin presentarla como revalidada por este cambio técnico.

## Evidencia y revisión

Estado de ejecución: **verificado localmente** con Go 1.26.1 en Windows. Pruebas Go, vet (incluidas fuentes de integración), compilación y módulos correctos. Integración completa del backend en CockroachDB 26.2.0; los diez escenarios HTTP/SQL de autenticación también pasan en 26.2.7. Actualización 009→010 comprobada con cuenta/credencial/sesión existentes; conserva datos y hash. Suite SQL en 26.2.0: 38 comprobaciones correctas, incluida instalación nueva y SQL único. Ejemplo PowerShell probado contra cmd/api con cuenta ficticia: login, consulta y logout sin imprimir secretos. OpenAPI YAML/referencias internas y enlaces locales revisados. Informe: [resultado-autenticacion.json](tests/resultado-autenticacion.json). CI remota se registra en la entrega del PR. Revisión de impacto: **Pendiente de revisión**; estas pruebas no equivalen a una prueba de penetración externa ni a una aprobación operativa o legal.

Pruebas verificadas: mensajes genéricos, verificación ficticia, parámetros PHC acotados, cookies y ausencia del bearer en JSON/SQL legible, CSRF/origen/Host, cuerpo/JSON, límites y memoria concurrentes, expiración absoluta/inactividad, rotación, cierre idempotente/concurrente, revocación total frente a login, suspensión/versiones/cambio de clave, permisos actuales y rollback si falla auditoría.

Fuentes técnicas consultadas: [OWASP Authentication](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html), [OWASP Session Management](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html), [OWASP CSRF](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html), [Go CrossOriginProtection](https://pkg.go.dev/net/http#CrossOriginProtection), [Go crypto/rand](https://pkg.go.dev/crypto/rand), [Go HMAC](https://pkg.go.dev/crypto/hmac). Los controles se verifican en este incremento; las referencias no implican conformidad integral con sus guías.
