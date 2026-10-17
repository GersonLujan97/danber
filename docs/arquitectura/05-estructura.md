# Estructura preliminar del repositorio, Go y React

Estado: **Propuesto**. Estos árboles son diseños, no carpetas de aplicación creadas. Se implementará solo lo necesario para el primer recorrido después de revisar A–C. Un repositorio común facilita revisar conjuntamente contrato HTTP, backend, frontend y migraciones.

## 1. Organización general

```text
CODE/
  README.md
  AGENTS.md
  Imagenes/                           # referencias originales
  docs/
    dominio/
    glosario/
    arquitectura/
    base-datos/
    decisiones/
    modulos/
    pruebas/
    seguridad/
  backend/                            # primera implementación vertical
    go.mod
    go.sum
    cmd/api/main.go
    internal/
    migrations/
  web/
    package.json
    package-lock.json                 # si se confirma npm
    src/
    e2e/
  contracts/openapi.yaml              # crear con la primera API
  .github/workflows/ci.yml            # solo si se elige GitHub
```

Propuesta de gestión: un módulo Go para el backend y un proyecto frontend. No un `go.mod` por contexto ni un monorepo con herramientas adicionales. El gestor frontend propuesto es npm por disponibilidad habitual; se fija uno y un archivo de bloqueo. Git es la tecnología de versionado prevista; la plataforma remota no está definida.

## 2. Backend orientado a capacidades

```text
backend/
  cmd/api/main.go                     # configuración, arranque, cierre ordenado
  internal/
    bootstrap/app.go                 # compone módulos, rutas y dependencias
    platform/
      config/config.go               # configuración validada, sin secretos en logs
      database/pool.go               # conexión y límites del pool
      database/retry.go              # repetición transaccional del adaptador
      httpx/errors.go                # representación común de errores HTTP
    identidad/
      module.go                      # fachada e inicialización del módulo
      internal/                      # implementación privada a este módulo
        app/
        httpapi/
        sqlstore/
    personal/
      module.go
      internal/
        app/
        httpapi/
        sqlstore/
    pacientes/
      module.go
      api/identificacion.go           # DTO publicados para consumidores
      internal/
        domain/paciente.go           # identidad, estado y reglas necesarias
        app/registrar.go
        app/buscar.go
        app/ports.go
        httpapi/handlers.go
        sqlstore/pacientes.go
    agenda/
      module.go                      # fachada pública, no expone entidades
      api/contracts.go               # comandos/resultados públicos mínimos
      internal/
        domain/cita.go
        domain/intervalo.go
        domain/cita_test.go
        app/reservar.go
        app/reprogramar.go            # cuando se incluya esa operación
        app/consultar_agenda.go
        app/ports.go
        app/reservar_test.go
        httpapi/handlers.go
        httpapi/handlers_test.go
        sqlstore/agenda.go           # transacción, consultas e historial
        sqlstore/agenda_integration_test.go
    integraciones/
      agendapacientes/adapter.go      # traduce contrato Pacientes al puerto Agenda
      agendapersonal/adapter.go
  migrations/                        # secuencia global, propiedad explícita
```

`cmd` agrupa ejecutables y `internal` restringe importación por ubicación. Una segunda carpeta `internal` dentro de Agenda permite que el compilador impida a otros módulos importar sus implementaciones. El `internal` superior por sí solo impide acceso desde fuera del backend, pero no impide acceso indebido entre todos sus subpaquetes. La organización oficial de Go presenta `cmd` e `internal` como herramientas habituales, sin exigir una arquitectura empresarial de capas. [Organización de módulos Go](https://go.dev/doc/modules/layout).

**Fachada** significa aquí una entrada pública pequeña que delega al caso de uso; resuelve el acceso controlado. La alternativa de exportar cada paquete interno simplifica imports pero expone detalles. Ejemplo: `agenda.Module.Reservar` recibe un contrato publicado y convierte hacia los tipos internos. No es otra capa de reglas. `api` contiene datos del contrato entre módulos, no controladores HTTP, entidades del dominio ni dependencias SQL.

Las interfaces de Pacientes y Personal requeridas por Agenda se declaran del lado consumidor. Si un tipo de contrato se necesita para construir un adaptador externo, se expone en `agenda/api`; `app` puede utilizar ese contrato sin que `api` importe la implementación. Los adaptadores entre módulos los conecta `bootstrap`. Ningún adaptador se incorpora al dominio.

En Identidad o Personal, una operación sencilla puede usar `app` sin carpeta `domain` hasta que existan reglas que merezcan un modelo propio. Dentro de un paquete Go los archivos organizan lectura, no agregan niveles de acceso: no crear carpetas `entities`, `valueobjects`, `validators`, `commands` y `queries` con un archivo cada una por obligación.

Carpetas futuras `clinica`, `tratamientos`, `economia` y `facturacion` se crean cuando se desarrolle su primer caso de uso. `workflows/seguimiento` solo aparece cuando se necesita el coordinador real. No se crean directorios vacíos para todo el mapa.

## 3. Dónde vive cada responsabilidad

| Pieza | Ubicación y criterio |
|---|---|
| Entidad, objeto de valor y agregado | `domain` si protegen reglas; métodos o funciones pequeñas, campos modificables mediante operaciones controladas |
| Servicio de dominio | Función de `domain` cuando una regla no encaja en una entidad; sin conexiones ni HTTP |
| Servicio de aplicación, comando y consulta | `app`, agrupados por caso de uso; un archivo no implica una clase ni un patrón CQRS completo |
| Puerto o repositorio | Necesidad definida por consumidor, cerca del caso de uso; semántica concreta, no `Repository[T]` genérico |
| Validación | HTTP comprueba forma; aplicación permisos y coordinación; dominio invariantes; base restricciones estructurales |
| Error de negocio | Tipo/código estable del módulo; transporte lo traduce a HTTP sin filtrar SQL ni datos sensibles |
| DTO | En transporte para HTTP o en API pública para otros módulos; no serializar directamente entidades privadas |
| Transacción | Caso de uso define alcance; adaptador SQL ejecuta y reintenta; no pasar driver al dominio |
| Evento | Tipo local del módulo; versión pública diferente si hay consumidor externo al contexto |
| Prueba | Junto al archivo/regla; integración marcada y ejecutada contra CockroachDB separado |
| Configuración y composición | `platform/config`, `bootstrap` y `cmd`; no un singleton global consultado por todas las reglas |

Propuesta técnica: `net/http` para transporte inicial; un router adicional solo si reduce complejidad demostrada. `pgx`/pool para SQL explícito; evitar ORM al principio para comprender transacciones e índices. `sqlc` puede evaluarse si mantener conversiones manuales se vuelve costoso; no se añade sin ese problema. Documentación primaria: [net/http](https://pkg.go.dev/net/http), [pgx](https://github.com/jackc/pgx). Que pgx hable el protocolo PostgreSQL no sustituye las pruebas con CockroachDB.

## 4. Frontend por funcionalidades

```text
web/src/
  main.tsx
  app/
    App.tsx
    router.tsx                       # compone rutas y páginas públicas de features
    providers.tsx                    # sesión y cliente de consultas
    layouts/ConsultorioLayout.tsx
    config.ts                        # configuración pública, nunca secretos
  features/
    identidad/
      pages/IniciarSesionPage.tsx
      api.ts
      session.ts                     # estado mínimo de sesión y capacidades
      index.ts
    pacientes/
      pages/PacientesPage.tsx
      pages/RegistrarPacientePage.tsx
      components/PacienteForm.tsx
      schemas.ts
      api.ts
      queries.ts
      types.ts
      PacienteForm.test.tsx
      index.ts
    agenda/
      pages/AgendaPage.tsx
      components/ReservarCitaForm.tsx
      components/AgendaDiaria.tsx
      schemas.ts
      api.ts
      queries.ts
      types.ts
      AgendaPage.test.tsx
      index.ts
  shared/
    api/http.ts                      # fetch, sesión, errores y correlación
    ui/Button.tsx                    # presentación sin reglas del consultorio
    ui/Field.tsx
    ui/Dialog.tsx
    time/format.ts                   # formato con zona explícita
web/e2e/reservar-cita.spec.ts
```

La estructura es orientativa: no se crea un archivo por cada nombre si aún no tiene contenido. `api.ts` llama al servidor; `queries.ts` define claves y actualización de caché; `schemas.ts` expresa validaciones de formulario; `types.ts` modela datos del contrato, no replica toda la base.

`app` coordina capacidades. Una feature no importa archivos internos de otra; usa su exportación pública si necesita una pieza reutilizable, o recibe identificadores y callbacks. Un selector de paciente puede publicarse desde Pacientes; no trasladar reglas de pacientes a `shared/ui` solo porque Agenda lo usa.

Las rutas propuestas son `/iniciar-sesion`, `/pacientes`, `/pacientes/nuevo` y `/agenda`. La agenda comienza como lista diaria accesible y reserva con fecha/hora; un calendario con arrastrar y soltar se evalúa después de validar duración y reprogramación. Una pantalla puede reunir datos de varios contextos sin convertirse en propietaria de sus reglas.

## 5. Bibliotecas: comparación y recomendación

| Necesidad | Alternativas razonables | Recomendación inicial y consecuencia |
|---|---|---|
| Construcción de interfaz | React con Vite / framework con renderizado de servidor | React + TypeScript + Vite para aplicación interna con backend Go; el framework integrado aporta más convenciones y despliegue adicional si no se necesita renderizado de servidor |
| Rutas | React Router / TanStack Router | React Router en modo declarativo por claridad; evaluar tipado de rutas más estricto si se vuelve necesario |
| Datos del servidor | `fetch` y efectos propios / TanStack Query | TanStack Query para consultas, reintentos de lecturas e invalidación; evita escribir un segundo almacén de pacientes y citas |
| Formularios | Estado React/HTML / React Hook Form | React Hook Form cuando haya varios campos; HTML y estado local bastan para filtros simples |
| Esquemas de validación | Funciones propias / Zod | Zod para formularios y datos externos que requieran comprobación en ejecución; backend sigue siendo autoridad |
| Estado de interfaz | `useState`/Context / Zustand / Redux | Estado local y contexto de sesión; no añadir un almacén global para duplicar la caché remota |
| Estilo | CSS Modules / Tailwind / biblioteca de componentes | CSS Modules y pocos componentes accesibles inicialmente; revisar si una biblioteca acelera formularios complejos sin imponer el dominio |
| HTTP | `fetch` / Axios | `fetch` con adaptador pequeño; una dependencia extra no es necesaria para estos contratos |
| Pruebas | Vitest + Testing Library / herramientas equivalentes | Vitest y Testing Library para componentes; Playwright para el recorrido completo |

Fuentes primarias consultadas para las capacidades de estas herramientas: [React, creación desde cero](https://react.dev/learn/build-a-react-app-from-scratch), [React Router](https://reactrouter.com/start/declarative/installation), [TanStack Query](https://tanstack.com/query/latest/docs/framework/react/overview), [React Hook Form](https://github.com/react-hook-form/react-hook-form), [Zod](https://zod.dev/), [Vitest](https://vitest.dev/guide/), [Testing Library](https://testing-library.com/docs/react-testing-library/intro/), [Playwright](https://playwright.dev/docs/intro). La recomendación de conjunto es una decisión propuesta del proyecto, no una obligación de esas herramientas. Las versiones exactas y compatibilidades se fijan y prueban en H.

## 6. Estado, errores y seguridad de la interfaz

- Fecha seleccionada, texto de búsqueda y diálogo abierto son estado local; pacientes y agenda son datos del servidor.
- Claves de caché incluyen filtros y contexto de sesión. Al salir o cambiar usuario se vacía caché; no persistir historias en almacenamiento del navegador por defecto.
- Reservar requiere confirmación del servidor antes de mostrar éxito. Un `409` por ocupación actualiza disponibilidad y mantiene los datos del formulario.
- Errores de campo se muestran al lado del campo; fallos temporales permiten reintento con la misma clave; `401` solicita sesión y `403` explica falta de acceso sin revelar datos ajenos.
- No reintentar automáticamente comandos de pago o emisión sin contrato idempotente.
- Navegación y botones respetan capacidades, pero el servidor comprueba nuevamente operación y recurso.
- Configuración enviada al navegador es pública. Claves de base, firma, correo o facturación nunca se exponen mediante variables de compilación frontend.
- Formularios con etiquetas, teclado, foco y errores anunciables; fecha, moneda y zona visibles. No usar solamente colores para estado clínico o económico.
