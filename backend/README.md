# Backend inicial de Danber

**CHG-015 · 2026-10-18:** alta/búsqueda administrativa de pacientes con documento, apoderado, permisos, auditoría e idempotencia. Nueva migración 011; conservar 001–010. Verificado localmente en CockroachDB 26.2.0 y 26.2.7; revisión de impacto pendiente. [Guía de pacientes](PACIENTES.md).

**Estado vigente · CHG-014:** autenticación HTTP con sesiones, controles de abuso/CSRF y revocación en `feat/autenticacion`, verificada localmente en las dos versiones admitidas. [Guía en Notion](https://app.notion.com/p/3fd2756ebb4481cfa6ddc7e41e798fe9) y [respaldo técnico](AUTENTICACION-Y-SEGURIDAD.md). Aplicar 010 y activar explícitamente la excepción HTTP solo para probar en loopback. Usar la cuenta ya creada; no repetir crear-maestro. Los apartados históricos siguientes conservan el desarrollo por etapas.

**Estado vigente · CHG-013:** cuenta Maestro inicial disponible por consola en `feat/cuenta-inicial`, verificada localmente en CockroachDB 26.2.0 y 26.2.7. Aplicar migración 009 antes de usarla. [Guía de creación en Notion](https://app.notion.com/p/3fc2756ebb448101ba65e6e54e1532ca) y [respaldo técnico](CUENTA-INICIAL.md). Inicio de sesión, recuperación y autorización por recurso siguen pendientes. El contenido CHG-011 inferior conserva el contexto del primer backend.

CHG-011 · 2026-10-17. Primer incremento en feat/backend-inicial. Go **1.26.1**, net/http y pgx **5.11.0**. Solo comprobaciones operativas; aún no se exponen cuentas, pacientes ni reservas.

Documentación principal: [Backend inicial en Notion](https://app.notion.com/p/3fc2756ebb4481068fb4eba26fb1228b).

## Iniciar en tu equipo

Puedes mantener CockroachDB 26.2.0 (tu instalación actual) o 26.2.7: SQL localhost:26257 y consola localhost:8081. La API usa localhost:8090, así no ocupa el puerto de tu consola.

Primero, con el administrador de tu nodo LOCAL, entra en SQL:

```powershell
.\cockroach sql --insecure --host=localhost:26257 --database=danber
```

Antes de usar el backend, aplica las migraciones pendientes desde la raíz del proyecto, usando la cuenta migradora (011 es la última en CHG-015):

```powershell
$env:DANBER_DATABASE_URL = 'postgresql://root@localhost:26257/defaultdb?sslmode=disable'
python db/scripts/migrate.py apply --database danber
```

La migración 008 retira CREATE en public al rol público; 009 agrega el control de cuenta inicial; 010 agrega actividad/versión e índice de sesiones. Las versiones anteriores conservan sus bytes. Luego, si aún no preparaste el usuario SQL, ejecuta una vez:

```sql
CREATE USER IF NOT EXISTS danber_api;
GRANT CONNECT ON DATABASE danber TO danber_api;
GRANT danber_runtime TO danber_api;
```

danber_api es la identidad de conexión SQL del proceso; todavía no es una cuenta Maestro de la aplicación. Tu nodo usa --insecure, por lo que estos permisos no sustituyen autenticación/TLS. Desarrollo local y datos ficticios; no exponer a red.

En una nueva terminal PowerShell:

```powershell
cd C:\GERSON\ISAI\DANBER\CODE\backend
$env:DANBER_DATABASE_URL = 'postgresql://danber_api@localhost:26257/danber?sslmode=disable'
$env:DANBER_HTTP_ADDR = '127.0.0.1:8090'
go run ./cmd/api
```

La primera ejecución descarga dependencias. Si ya tienes DANBER_DATABASE_URL con root para migraciones, reemplazarla en ESTA terminal por danber_api; no cambiar el usuario migrador ni guardar secretos en el código. La aplicación lee el entorno del proceso; no importa .env automáticamente. .env.example es referencia y .gitignore excluye tus .env privados.

Consultar en navegador o desde otra terminal:

```powershell
Invoke-RestMethod http://127.0.0.1:8090/health/live
Invoke-RestMethod http://127.0.0.1:8090/health/ready
```

| Ruta | 200 | Falla |
|---|---|---|
| GET /health/live | Proceso HTTP responde: status ok | No consulta la base |
| GET /health/ready | Conexión SQL operativa: status ready | 503, status unavailable, sin detalle interno |

Al arrancar comprueba conexión, versión CockroachDB 26.2.0 o 26.2.7 y que el usuario SQL no sea administrador. Si falla, termina con código 1 y mensaje genérico. Mientras corre, una pérdida de conexión hace fallar ready sin volver unhealthy al proceso solo por esa dependencia. ready verifica conectividad, no demuestra migraciones completas, autorización de negocio ni disponibilidad de toda una aplicación.

Ctrl+C cierra HTTP antes del pool, permitiendo finalizar solicitudes en curso durante hasta cinco segundos. También se atiende SIGTERM en plataformas compatibles. No migrar automáticamente al arrancar ni usar una cuenta administradora en la API.

## Árbol creado y responsabilidad

```text
backend/
  cmd/api/main.go
  internal/
    bootstrap/app.go
    health/handler.go
    platform/
      config/config.go
      database/pool.go
  .env.example
  go.mod
  go.sum
```

- cmd/api convierte configuración/ejecución en código de salida, configura logs y captura señales.
- bootstrap compone las piezas y controla su ciclo de vida.
- config valida entorno y falla con mensajes que no imprimen secretos. Este incremento solo admite loopback para HTTP y SQL; TLS local explícito disable o verify-full, sin modos de fallback silencioso ni opciones que reescriban host/usuario/base.
- database contiene el driver y un pool pequeño (máximo cuatro conexiones); no se exporta pgx a reglas de negocio.
- health define únicamente las comprobaciones, con una función de conexión pequeña para poder probar fallos sin simular persistencia.
- CHG-013 agrega Identidad y el comando crear-maestro con reglas y adaptadores privados del módulo. Personal, Pacientes y Agenda se crearán cuando se implemente su capacidad; no hay repositorios genéricos ni módulos vacíos. Las migraciones siguen en db/, sin duplicarlas en backend/migrations.

## Buenas prácticas aplicadas y por qué

Un pool reutiliza conexiones con límites: evita abrir una por cada petición y saturar el servidor. La alternativa es conexión directa por operación; su costo crece con el uso.

Un context con plazo limita cuánto puede esperar una conexión o comprobación; si CockroachDB deja de responder, no retiene indefinidamente la petición. Se usan cinco segundos de arranque y dos de readiness como valores operativos iniciales, revisables.

Liveness y readiness separadas permiten identificar proceso y dependencia. La alternativa de una única respuesta siempre positiva ocultaría una base caída; hacer depender liveness de SQL confundiría ambos fallos.

net/http reduce dependencias iniciales; pgx queda en el adaptador SQL. Logs JSON de ciclo de vida omiten URLs, credenciales y datos de pacientes. No hay logging de cuerpos ni cadenas de conexión. El transporte tiene límites de tiempo y de cabeceras.

Regla de versión y rechazo de administrador evitan arrancar con un entorno distinto del verificado o con privilegios del migrador. No son una auditoría completa del despliegue; la autorización por recurso llega con cada caso de uso.

## Comprobaciones

Desde backend:

```powershell
gofmt -l cmd internal
go test ./...
go vet ./...
go build -o ../.local/danber-api.exe ./cmd/api
```

Las pruebas HTTP usan httptest y las de configuración cubren rechazo de administrador, exposición de red, opciones ambiguas y mensajes sin secretos. Son pruebas de comportamiento; no simulan aislamiento SQL.

Integración: usar una base aislada y usuario sin administración en CockroachDB real, nunca credenciales/BD de producción:

```powershell
$env:DANBER_TEST_DATABASE_URL = 'postgresql://danber_api_test@127.0.0.1:26379/danber_backend_test?sslmode=disable'
$env:DANBER_TEST_ADMIN_URL = 'postgresql://root@127.0.0.1:26379/danber_backend_test?sslmode=disable'
go test -tags=integration -count=1 ./...
```

Preparar esa base con las migraciones existentes y otorgar danber_runtime a danber_api_test. Las variables de ejemplo corresponden al entorno aislado de verificación, no se crean al ejecutar la API.

## CI y contrato HTTP

El workflow .github/workflows/backend.yml comprueba formato, vet, tests con detector de carreras y compilación en Linux al enviar cambios del backend o abrir PR. La integración SQL se verifica localmente contra CockroachDB aislado; no se afirma cubierta por ese workflow. El contrato HTTP está en contracts/openapi.yaml.

## Flujo y alcance posterior

Rama feat/backend-inicial → comprobaciones → PR hacia main → revisión → integración por usuario. No se integra automáticamente. La adopción del flujo de ADR 0010 se confirma por la instrucción del usuario de continuar esa etapa; nuevas decisiones clínicas o de identificación permanecen pendientes.

CHG-013 incorpora aprovisionamiento controlado de cuenta Maestro como entrega separada. Próxima capacidad: inicio/cierre de sesión, expiración y revocación. Sin credenciales por defecto, ORM, entidades compartidas ni algoritmos de reserva adelantados.

CHG-014 implementa esa autenticación. Guía vigente de ejecución y controles: [Acceso y seguridad](AUTENTICACION-Y-SEGURIDAD.md). Pacientes/Agenda y autorización por recurso son siguientes incrementos, con pruebas propias; la seguridad no se pospone al final.

Fuentes: [net/http](https://pkg.go.dev/net/http), [pgxpool](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool). Tener protocolo PostgreSQL no sustituye probar CockroachDB.




