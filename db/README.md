# Crear Danber en CockroachDB local

Entrega CHG-008 · 2026-10-17. Versión fijada: **CockroachDB 26.2.7**. Base inicial del primer recorrido: Identidad, Personal, Pacientes y Agenda. Incluye espera, horarios, bloqueos e historial; todavía no hay API ni pantallas.

Versión principal: [Crear la base local en Notion](https://app.notion.com/p/3fc2756ebb44819a8a6df741dc1f3e50).

## Qué ejecutar ahora

Desde PowerShell, situado en C:\GERSON\ISAI\DANBER\CODE:

```powershell
# 1. Instalar el binario oficial si aún no está disponible.
.\db\scripts\Install-Cockroach.ps1

# 2. Iniciar un nodo local persistente. Mantener abierta esta terminal.
.\db\scripts\Start-Local.ps1
```

En **otra terminal**, en la misma carpeta:

```powershell
# 3. Crear danber y aplicar sus siete migraciones.
.\db\scripts\Initialize-Database.ps1

# 4. Consultar el historial de migraciones.
$env:DANBER_DATABASE_URL = 'postgresql://root@127.0.0.1:26257/defaultdb?sslmode=disable'
python .\db\scripts\migrate.py status --database danber
```

Se requiere Python 3.8 o posterior en PATH; no instalar paquetes con pip. El script usa la CLI oficial. La instalación queda en .local, sin modificar el PATH ni instalar servicios. Si ya se descargó y verificó durante este trabajo, el primer paso solo muestra la versión instalada. No reinstalar innecesariamente.

El nodo escucha únicamente en localhost; SQL 26257, RPC 26258 y consola http://127.0.0.1:8080. Datos persistentes en .local/cockroach-data-dev. Ctrl+C detiene el proceso y se vuelve a iniciar con el mismo script. No borrar esa carpeta para aplicar una actualización.

Este inicio sin TLS es para desarrollo local con datos ficticios. El binario oficial Windows está publicado como experimental; un despliegue real requiere definir plataforma y conexión segura. No se crearon cuentas ni contraseñas de acceso de la aplicación.

Si PowerShell bloquea archivos descargados, revisar primero su origen y la política del equipo. No es necesario cambiar la política global ni ejecutar todo como administrador para generar SQL.

## Si ya tienes CockroachDB local

Omitir instalación/inicio y configurar la conexión real en el entorno de esa terminal. No pegar contraseñas en Notion o en el repositorio.

```powershell
$env:COCKROACH_BIN = 'C:\ruta\cockroach.exe'
$env:DANBER_DATABASE_URL = 'postgresql://root@127.0.0.1:26257/defaultdb?sslmode=disable'
python .\db\scripts\migrate.py init --database danber
```

Cambiar host, puerto, usuario y TLS según tu instancia. El ejecutor comprueba que el servidor es 26.2.7 antes de migrar. Si usas otra versión, detenerse y verificarla con el suite antes de cambiar la versión fijada; no quitar el chequeo para forzar compatibilidad.

Initialize-Database respeta una DANBER_DATABASE_URL ya definida. El parámetro --database selecciona explícitamente la base destino; el usuario necesita crear base/esquemas/tablas/roles y conceder privilegios. Usar una cuenta migradora, separada del futuro login SQL del backend.

## Alternativa: SQL único

Archivo: [danber-inicial.sql](danber-inicial.sql). Crea **danber** y todos los objetos, catálogos y permisos de esta entrega.

En una instancia 26.2.7, comprobar versión y ejecutar desde PowerShell:

```powershell
& '.\.local\cockroach\cockroach-v26.2.7.windows-6.2-amd64\cockroach.exe' sql --url=$env:DANBER_DATABASE_URL --execute='SELECT version();'
& '.\.local\cockroach\cockroach-v26.2.7.windows-6.2-amd64\cockroach.exe' sql --url=$env:DANBER_DATABASE_URL --set=errexit=true --file='.\db\danber-inicial.sql'
```

Elegir **el ejecutor o el SQL único** para el alta inicial. El SQL único exige una base danber nueva: si existe, se detiene; no incluye DROP ni borra datos. Registra los mismos checksums, por lo que las siguientes actualizaciones pueden gestionarse con el ejecutor.

El runner se recomienda para instalaciones nuevas y actualizaciones: comprueba versión, hashes e historial. La comprobación de versión del SQL único se hace expresamente antes de ejecutarlo. Si tu herramienta SQL no detiene la ejecución al primer error o no conserva la sesión después de SET DATABASE, usar la CLI.

Para regenerar el SQL único desde las migraciones originales:

```powershell
python .\db\scripts\migrate.py export
```

No editar el archivo generado a mano. Para usar un nombre distinto, exportar con --database danber_dev --output .local/danber_dev.sql; nombre validado, sin SQL arbitrario.

## Archivos y secuencia

| Archivo | Contenido |
|---|---|
| 001_esquemas.sql | Esquemas de los cuatro propietarios |
| 002_identidad.sql | Usuario, credencial hash, sesión hash, roles/permisos y cambios de acceso |
| 003_personal.sql | Trabajador, profesional, especialidades y vínculos |
| 004_pacientes.sql | Paciente, tipo/documento, apoderado, parentesco e historial |
| 005_agenda.sql | Horario, excepciones, bloqueos, cita, espera, historial, guard e idempotencia |
| 006_catalogos_y_permisos.sql | DNI y roles/permisos iniciales, sin personas reales |
| 007_privilegios_runtime.sql | Rol SQL sin login danber_runtime con permisos restringidos |
| danber_meta.migracion | Historial de versión, nombre, SHA256 y estado |

Una migración es un cambio versionado, no un script que se repite libremente. Cada archivo aplicado queda congelado por su checksum; nuevos cambios se incorporan en un archivo con número siguiente. Una reejecución correcta no duplica tablas ni catálogos.

La DDL se ejecuta fuera de una transacción global porque CockroachDB tiene restricciones para mezclar cambios de esquema. Se marca cada archivo como aplicando antes de ejecutarlo y aplicada solo cuando termina. Dos ejecutores que compitan por el mismo número no pueden aplicar ambos ese archivo.

Si falla, queda **aplicando**: detenerse, revisar salida, SHOW JOBS y objetos realmente creados. No cambiar el estado a aplicada sin comprobar y completar las instrucciones faltantes. No borrar el historial ni reejecutar el archivo a ciegas. Se prefiere corrección hacia adelante; no se incluye rollback destructivo de pacientes o movimientos.

## Qué protege la base y qué requiere backend

| Regla | Cobertura de esta entrega |
|---|---|
| Campos obligatorios, no vacíos y tamaños | NOT NULL/CHECK/tipos; teléfono y correo no acreditan contacto real |
| Documento repetido por tipo | UNIQUE con número calculado upper/btrim, DNI de ocho dígitos |
| Un documento principal por paciente | Índice único parcial; al menos uno se garantiza en el caso de uso |
| Documento del apoderado y parentesco | Columnas requeridas y referencias; condición por edad necesita caso de uso |
| Mínimo de 30 minutos de reserva | CHECK; atención real puede durar menos o más |
| Citas adyacentes | Permitidas; la exclusión completa necesita protocolo transaccional |
| Dos atenciones abiertas del mismo profesional | Índice único parcial |
| Espera enlazada con otro paciente/profesional | FK compuesta lo rechaza |
| Historial, cuentas o pacientes referenciados | Sin cascada de borrado |
| Runtime modifica/borrra historial | Sin privilegios UPDATE/DELETE en tablas de cambios y recibos |
| No solapamiento, horarios, apoderado por edad | Ejemplo SQL transaccional verificado; backend debe implementarlo |
| Operación por rol y por profesional | Catálogo base; autorización de aplicación aún no implementada |

Un INSERT directo del propietario puede saltarse el protocolo de disponibilidad; la estructura SQL por sí sola no equivale a una agenda terminada. Todas las escrituras de disponibilidad deben actualizar control_profesional y revalidar bajo SERIALIZABLE. El ejemplo [reservar_referencia.sql](tests/reservar_referencia.sql) usa esa fila, intervalo completo, horarios/excepciones, bloqueos, identificación, historial e idempotencia.

El ejemplo contiene parámetros de plantilla para el suite, no valores listos para pegar. El backend futuro debe usar parámetros del driver, derivar actor de la sesión, comprobar permisos y devolver errores de negocio; no sustituir texto de usuarios ni considerar éxito solo porque SQL terminó sin error. El recibo compatible identifica una reserva confirmada. Reintentar la transacción completa ante 40001 con la misma intención.

No se incluyen triggers que conviertan cualquier INSERT en una reserva segura ni se afirma compatibilidad con SQL Server. La tabla de recibos v1 corresponde a operaciones que producen cita; idempotencia de cambios de horario/bloqueos debe concretarse con su caso de uso.

## Privilegios y carga inicial

Maestro/Recepción/Odontólogo son roles de la aplicación. danber_runtime es un rol de la conexión SQL del backend; no se entrega al navegador. El migrador conserva DDL. El runtime puede leer/escribir tablas operativas, insertar historial, pero no alterar/borrar esos historiales ni modificar la tabla de migraciones.

El login SQL, contraseña/certificado y GRANT danber_runtime TO login se provisionan fuera de los archivos. Conexión root sin TLS solo para desarrollo local. El usuario maestro de la aplicación se creará después mediante un flujo controlado con hash de contraseña; no se inserta una contraseña por defecto.

Antes de reservar con datos configurados: crear cuenta/personal/profesional, asociar especialidad, declarar elegibilidad operativa, crear control_profesional y definir horario. El alta profesional y su control deben coordinarse antes de ofrecer reservas. estado pendiente impide usar un profesional aún no revisado.

No sembrar odontólogo ficticio ni sábados hasta las 21:00 en una base de trabajo. Los fixtures del suite se ejecutan solamente en bases con nombre danber_test_ aleatorio. Solo se siembra DNI; otro documento requiere agregar un tipo concreto y su política de validación. Especialidad, teléfono de menores y responsables propuestos conservan su estado de revisión.

## Pruebas reproducibles

Con el servidor local activo:

```powershell
$env:DANBER_DATABASE_URL = 'postgresql://root@127.0.0.1:26257/defaultdb?sslmode=disable'
python -u .\db\tests\verify.py
```

El suite crea bases danber_test_ con sufijo aleatorio; no borra ni usa datos de danber. Prueba aplicación desde cero/actualización desde v3, reejecución, integridad, reservas competidoras con conexiones independientes, idempotencia, almuerzo, bloques, rollback y privilegios SQL. Guarda [resultado-verificacion.json](tests/resultado-verificacion.json). No es una prueba de interfaz, API ni autorización por recurso.

Las bases sintéticas se dejan para inspección en tu instancia; identificar únicamente esos nombres antes de cualquier limpieza manual. La verificación de esta entrega se ejecuta en un nodo en memoria separado, puertos 26379/18089; al detenerlo no afecta tu futuro nodo persistente de 26257.

## Pendientes antes de uso real

Validar tipos documentales admitidos; criterio operativo de menor/contacto/apoderado; catálogo de especialidades; cierre sabatino; responsables de estados/espera y política de reservas paralelas del paciente. El esquema admite las alternativas sin declarar esos pendientes confirmados.

Elegibilidad de profesional en referencia se comprueba en el orden serializable de la reserva; la baja debe revisar citas futuras explícitamente y no garantiza cancelación automática. La concurrencia con baja requiere prueba de ese futuro caso de uso, sin prometer atomicidad entre módulos por una consulta previa.

Definir retención de historial/recibos, permisos de despliegue, conexión segura y copia/restauración antes de datos reales. Siguiente entrega: backend mínimo con operaciones y autorización, pruebas de los contratos y frontend del primer recorrido.

## Fuentes técnicas

[Release 26.2.7](https://docs.cockroachlabs.com/docs/releases/v26.2#v26-2-7), [CREATE TABLE](https://docs.cockroachlabs.com/docs/v26.2/create-table), [índices parciales](https://docs.cockroachlabs.com/docs/v26.2/partial-indexes), [DDL y transacciones](https://docs.cockroachlabs.com/docs/v26.2/online-schema-changes#schema-changes-within-transactions), [reintentos](https://docs.cockroachlabs.com/docs/v26.2/transaction-retry-error-reference), [CLI SQL](https://docs.cockroachlabs.com/docs/v26.2/cockroach-sql).


