# ADR 0011 — backend inicial y conexión de ejecución

Fecha: 2026-10-17. Estado: **Implementado para desarrollo local** por instrucción del usuario; decisiones operativas configurables, sin cerrar reglas de negocio pendientes. Revisión de impacto: **Pendiente de revisión**. CHG-011. Rama feat/backend-inicial; el usuario adopta el flujo de ADR 0010 al autorizar esta etapa.

## Resultado y límites

Un módulo Go github.com/GersonLujan97/danber/backend, Go 1.26.1 instalado/verificado, net/http y pgx 5.11.0 fijado con go.sum. cmd/api arranca y cierra el proceso; bootstrap compone dependencias; platform/config valida entorno; platform/database administra pool; health expone GET /health/live y GET /health/ready.

Aún no se crean carpetas vacías de Identidad/Personal/Pacientes/Agenda. Se crearán al implementar cada capacidad. Migraciones permanecen en db/. No existen rutas de negocio, autenticación ni usuario Maestro todavía.

## Decisiones y alternativas

- Usar biblioteca HTTP estándar evita un framework innecesario en este alcance. Driver pgx queda en adaptador SQL; entidades futuras no importan HTTP ni drivers.
- Configuración explícita del entorno; sin carga mágica de .env. Bind HTTP 127.0.0.1:8090 para evitar consola Cockroach 8081 del usuario y puerto estándar 8080. Este incremento exige SQL y HTTP en loopback.
- Pool máximo cuatro, plazos de conexión/arranque cinco segundos y readiness dos segundos, límites HTTP y cierre ordenado hasta cinco segundos. Son valores técnicos iniciales revisables, no reglas clínicas.
- Rechazar root/admin/node y pertenencia efectiva al rol SQL admin; comprobar motor/versión 26.2.7 antes de servir. Ejecución con danber_api miembro de danber_runtime, separado del migrador y del futuro Maestro.
- Mensajes SQL/configuración genéricos sin secretos. Logs JSON solo de ciclo de vida; health no devuelve host, contraseña ni datos clínicos. Solo se admiten opciones TLS de conexión; no query params que reescriban usuario/host/base.
- Liveness no depende de SQL; readiness sí, con 503 ante fallo. El proceso exige conexión inicial; readiness posterior no demuestra migraciones completas ni autorización.
- Nueva migración 008 retira CREATE de public al rol público: una prueba real detectó que el runtime podía crear tablas en public. No se editan las siete migraciones previas; regenerar SQL inicial y actualizar suite al total de versiones.
- Contrato health en contracts/openapi.yaml. CI en .github/workflows/backend.yml para formato, vet, tests con race y compilación en Linux. No incluye integración SQL automática ni configura protecciones de main.

Alternativas: usar cuenta root simplificaría el arranque pero mezclaría migrador y ejecución; responder siempre 200 ocultaría una base caída; añadir todos los módulos vacíos o capas genéricas anticiparía necesidades. Una rama por capacidad completa acumularía cambios: se entrega solo infraestructura inicial revisable.

## Verificación

go test ./..., go vet ./... y go build ejecutados con Go 1.26.1 en Windows. Integración con CockroachDB 26.2.7 aislado en memoria: conectividad con usuario de ejecución, rechazo de administrador, liveness/readiness ante pool cerrado, servidor HTTP real y cierre ordenado, y denegación de DDL por SQLSTATE 42501 tras migración 008. El test exige ese código específico, no cualquier error.

El suite SQL se repite por la nueva migración; resultado actualizado en db/tests/resultado-verificacion.json. La prueba inicial que permitió DDL motivó la corrección; su fallo queda explicado sin presentarlo como resultado final. CI remota se informa al consultar su ejecución; no se presume exitosa por crear YAML.

No se probó autenticación de aplicación ni RBAC por recurso porque aún no existen. No se usa la base danber del usuario. Tests SQL no sustituyen futuros casos de uso y viceversa.

## Impacto y revisión

Afecta backend, contratos health, CI, .gitattributes, db/migrations/008, SQL exportado, suite SQL y guías; capítulos 03, 04, 05, 07 y portada. Definir usuario SQL local y aplicar 008 en instalaciones existentes. La API no lo ejecuta ni migra silenciosamente.

README/AGENTS y Notion deben señalar alcance vigente; ADR 0010 pasa a adoptado, sin activar automáticamente protecciones GitHub ni exigir aprobación externa. Tipo documental, apoderados, permisos clínicos y demás reglas pendientes permanecen abiertos.

Guía de ejecución: [backend/README](../../backend/README.md). Próxima entrega: cuenta Maestro controlada y sesiones, en rama nueva desde main integrado. Esta rama se envía y propone mediante PR, sin fusionarla automáticamente.

Fuentes: [net/http](https://pkg.go.dev/net/http), [pgxpool](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool), [GitHub REST PR](https://docs.github.com/en/rest/pulls/pulls).

