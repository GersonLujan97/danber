# Crear la cuenta Maestro inicial

CHG-013 · 2026-10-17. Incremento `feat/cuenta-inicial`, posterior al PR #1 integrado en `main` (583343b). Alcance autorizado por Isai: continuar con la cuenta inicial. Revisión de impacto: **Pendiente de revisión**. La creación real de tu cuenta se hace en tu terminal, con una contraseña que eliges tú.

## Qué se crea

Una cuenta de la aplicación activa con roles Maestro y Recepción, su hash de contraseña y un registro de auditoría. No crea un usuario SQL, perfil profesional, DNI, especialidad ni sesión. `danber_api` sigue siendo la conexión técnica a CockroachDB; `Isai`, por ejemplo, será el nombre de acceso a Danber. El inicio de sesión corresponde al siguiente incremento.

El comando funciona una sola vez por base y rechaza nombres existentes sin elevarlos ni reemplazar contraseñas. Si ya existe cualquier Maestro, incluso suspendido, no crea otro. La cuenta inicial no limita el número futuro de odontólogos: estos tendrán perfiles y permisos propios.

## Pasos en tu equipo

Mantén encendido tu nodo CockroachDB 26.2.0 o 26.2.7. Desde la raíz del proyecto, aplica las migraciones con el usuario migrador:

```powershell
cd C:\GERSON\ISAI\DANBER\CODE
# Si cockroach no está en PATH, indica dónde está tu ejecutable:
# $env:COCKROACH_BIN = 'C:\ruta\cockroach.exe'
$env:DANBER_DATABASE_URL = 'postgresql://root@localhost:26257/defaultdb?sslmode=disable'
python db/scripts/migrate.py apply --database danber
```

Debe mostrar las versiones anteriores sin cambios y la nueva `009_aprovisionamiento_inicial.sql` aplicada. No repitas `danber-inicial.sql` en una base existente ni borres datos. Si aparece una migración incompleta o checksum distinto, detente y revisa el estado; no lo marques aplicado manualmente.

La identidad SQL `danber_api` debe existir y pertenecer a `danber_runtime`, como en la guía del backend. Si aún no la preparaste, ejecuta esto como migrador en la consola SQL de `danber`:

```sql
CREATE USER IF NOT EXISTS danber_api;
GRANT CONNECT ON DATABASE danber TO danber_api;
GRANT danber_runtime TO danber_api;
```

Después, en una terminal interactiva PowerShell:

```powershell
cd C:\GERSON\ISAI\DANBER\CODE\backend
$env:DANBER_DATABASE_URL = 'postgresql://danber_api@localhost:26257/danber?sslmode=disable'
go run ./cmd/crear-maestro --usuario Isai
```

Escribe tu contraseña y repítela. No verás caracteres ni asteriscos al escribir: es normal. Pulsa Enter después de cada entrada. Usa entre **15 y 128 caracteres**; puedes usar una frase larga con espacios. No hay contraseña por defecto. No la escribas como argumento, variable de entorno, en Notion, Git ni en este chat. El comando rechaza redirecciones de entrada para evitar dejarla en archivos o scripts.

Éxito: “Cuenta inicial creada con roles Maestro y Recepción”. Una segunda ejecución informa que ya está creada y conserva su contraseña y roles. Si las entradas no coinciden, no hay escrituras. Si falta 009, aplica migraciones y repite. Si se pierde la conexión al confirmar, el resultado puede ser desconocido: vuelve a ejecutar el comando; si la primera operación se confirmó, se rechazará la segunda sin duplicarla. No existe todavía un mecanismo de recuperación o cambio de contraseña; diseñarlo por separado.

El nodo `--insecure` continúa siendo solo desarrollo local con datos ficticios. El hash protege el secreto almacenado, pero no convierte ese nodo en una instalación segura ni impide que un administrador SQL modifique datos.

## Por qué se hace así

**Aprovisionamiento** significa crear el primer acceso cuando aún no existe una sesión autorizada. Se usa un comando local con acceso SQL de ejecución en lugar de una ruta pública que cualquiera podría llamar. Ejemplo: crear el acceso del dueño del consultorio antes de permitir que la recepción ingrese. La operación local se basa en el control de tu equipo; no autentica al operador en un nodo inseguro.

**Hash con Argon2id:** guarda una derivación que permite comprobar la contraseña después, sin guardar la contraseña legible. Un salt aleatorio de 16 bytes hace diferentes los hashes aunque dos personas escojan la misma frase. Parámetros iniciales: 19 MiB, dos iteraciones, paralelismo uno, resultado 32 bytes; formato PHC con versión y parámetros. Un hash rápido como SHA-256 facilitaría probar muchas contraseñas robadas. Los parámetros deberán medirse en el entorno de despliegue. No se cambia ni recorta silenciosamente la contraseña; el futuro inicio de sesión debe respetar sus bytes UTF-8.

**Transacción serializable:** una única operación guarda cuenta, credencial, roles, auditoría y marca de creación. Una fila de control se actualiza antes de consultar/crear: dos consolas no pueden obtener dos cuentas iniciales. La alternativa de consultar primero si existe un Maestro deja una carrera entre ambas consolas. Ante SQLSTATE 40001 se repite toda la transacción, hasta cinco intentos; una desconexión de resultado incierto no se reintenta automáticamente.

**Auditoría:** registra origen `consola_local`, operador del sistema operativo, conexión SQL y los dos roles. La cuenta nueva figura como actor técnico del aprovisionamiento con `aprovisionamiento_sin_sesion`; no se inventa una sesión previa ni una firma personal autenticada. No guarda contraseña ni hash en ese historial. Si falla la auditoría, también se revierte la cuenta.

**Módulo de Identidad:** dominio valida datos sin SQL/HTTP; aplicación coordina el caso de uso; adaptadores implementan Argon2id y CockroachDB. La fachada del módulo permite invocar la capacidad desde la consola. Las otras capacidades no consultan sus tablas directamente. La alternativa de poner todo en `main.go` mezclaría reglas y acceso SQL, dificultando probar los fallos.

La longitud es una política técnica inicial, no una nueva regla clínica confirmada. Lista de contraseñas comprometidas, autenticación, limitación de intentos, sesiones, revocación, autorización por operación/recurso, recuperación y despliegue seguro siguen **pendientes**; este incremento no afirma cumplimiento integral OWASP/NIST.

## Comprobaciones y alcance

Verificado con Go 1.26.1 en Windows: pruebas unitarias, vet, compilación de ambos comandos y verificación de módulos. Integración completa del backend en CockroachDB 26.2.0; los cinco casos de Identidad también pasan en 26.2.7. Comando completo probado en una terminal PTY con contraseña ficticia sin eco y resultado SQL comprobado sin mostrar hashes. Suite SQL en 26.2.0: 38 comprobaciones correctas con las nueve migraciones y SQL inicial. Evidencia: [resultado-cuenta-inicial.json](tests/resultado-cuenta-inicial.json). CI remota se registra en la entrega del PR.

La preparación inicial de los tests se corrigió para confirmar cada sentencia DDL como el migrador real y omitir comentarios al separar sentencias. Esos fallos de preparación no se presentan como verificaciones correctas; los resultados anteriores corresponden a la ejecución final.

Pruebas unitarias: Unicode/límites, rechazo de argumentos de contraseña y entrada no interactiva, confirmación, salida sin secretos, borrado de buffers, fallos de hashing sin escrituras, PHC/salt y derivación verificable. Adaptador SQL: reintentos acotados para 40001, sin repetir COMMIT incierto.

Integración con CockroachDB real: cuatro creaciones concurrentes (una sola ganadora), cuenta/credencial/dos roles/auditoría/marca consistentes, repetición sin cambios, nombre existente sin elevación, reversión si falla auditoría, Maestro suspendido preexistente y actualización 008→009. El test crea bases nuevas `danber_test_cuenta_*` en un nodo de pruebas local; no usa tu base `danber` ni la borra. Requiere las dos variables de integración de [README](README.md).

Desde `backend`:

```powershell
go test ./...
go vet ./...
go build ./...
go test -tags=integration -count=1 -v ./...
```

El workflow de GitHub verifica formato, vet, pruebas con detector de carreras y compilación de todos los comandos. Las pruebas SQL reales se ejecutan aparte; no se presume cobertura de integración por ese workflow.

## Siguiente incremento

Después de revisar e integrar este PR, crear tu cuenta con los pasos anteriores. Luego, una rama nueva para inicio/cierre de sesión, expiración y revocación, con autorización en el backend antes de exponer Pacientes y Agenda. No se debe añadir autenticación parcial a sus rutas.

Fuentes consultadas: [Argon2id en Go](https://pkg.go.dev/golang.org/x/crypto/argon2), [lectura sin eco](https://pkg.go.dev/golang.org/x/term), [OWASP: almacenamiento de contraseñas](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html), [NIST SP 800-63B-4](https://pages.nist.gov/800-63-4/sp800-63b.html). Revisar la política completa en la fase de autenticación; citar estas fuentes no acredita conformidad.
