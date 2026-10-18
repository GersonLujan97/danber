# ADR 0013 — cuenta Maestro inicial por consola

Fecha: 2026-10-17. CHG-013. Origen: Isai solicita actualizar su copia tras integrar PR #1 y continuar el siguiente incremento. Autor técnico: Codex. Estado: **Implementado y verificado localmente**, alcance autorizado; política técnica inicial revisable. Revisión de impacto: **Pendiente de revisión**. Rama `feat/cuenta-inicial`, desde `main` 583343b.

## Antes y después

Antes había catálogos y tablas de Identidad, conexión SQL y health, sin cuentas de aplicación. Ahora se incorpora `cmd/crear-maestro`: creación local única de una cuenta activa Maestro + Recepción, hash Argon2id y auditoría atómicos. El usuario escoge su contraseña en su terminal; no se crea su cuenta real durante esta implementación. No se inventan datos profesionales ni se concede el rol odontólogo sin perfil.

Complementa ADR 0005 (roles) y 0011 (backend); no sustituye identificación de pacientes, reglas de agenda ni permisos profesionales. Confirma el alcance autorizado, sin convertir otras propuestas de seguridad en aprobaciones del consultorio.

## Decisión y alternativas

- Consola interactiva local; ninguna ruta HTTP de bootstrap. El control del equipo y la conexión SQL son la frontera de esta operación; no se presenta como autenticación fuerte en un nodo `--insecure`.
- Contraseña 15–128 puntos de código Unicode válidos, sin recortar ni normalizar silenciosamente. Argon2id con m=19456 KiB,t=2,p=1, salt aleatorio 16 bytes y salida 32 bytes, PHC versionado. Dependencias fijadas x/crypto 0.58.0 y x/term 0.47.0; Go 1.26.1. Longitud/parámetros son política técnica inicial; medir en despliegue y revisar autenticación completa.
- Nueva migración 009 con una fila de control del aprovisionamiento. Actualizarla dentro de una transacción serializable antes de comprobar/insertar. No modificar bytes de 001–008 ni reiniciar bases existentes. Regenerar el SQL inicial solo para bases nuevas.
- Una transacción guarda usuario, credencial, Maestro y Recepción, cambio_acceso y marca. Rechazar Maestro previo incluso suspendido, marca consumida o nombre existente; no elevar ni restablecer contraseñas. Reintentar toda transacción solo ante 40001, máximo cinco; ante COMMIT incierto verificar mediante nueva ejecución segura.
- Auditoría de origen, operador OS y conexión SQL. El actor obligatorio referencia la cuenta nueva y se etiqueta expresamente como aprovisionamiento sin sesión; no atribuir una autorización anterior inexistente.
- Dominio independiente de SQL/HTTP, aplicación con contratos pequeños y adaptadores privados del módulo. La base solo se comparte a través del punto de composición; no añadir repositorio genérico.

Alternativas descartadas: contraseña fija en semillas o argumentos (filtración), alta pública sin sesión (riesgo de apropiación inicial), SQL manual de hashes/roles (difícil verificar atomicidad y repetir), comprobación previa sin guardia transaccional (carrera), modificar migraciones aplicadas (rompe historia). Un segundo Maestro o recuperación se harán con un proceso explícito futuro; no un interruptor para saltar el control inicial.

## Verificación

Go 1.26.1 en Windows: pruebas unitarias, vet, compilación y verificación de módulos correctas. Integración completa del backend en CockroachDB 26.2.0; los cinco casos de Identidad también pasan en 26.2.7. Incluye cuatro llamadas concurrentes (una ganadora), repetición sin cambios, nombre previo sin elevación, fallo de auditoría con rollback, actualización 008→009 y Maestro suspendido previo. Prueba del comando completo en terminal PTY con datos ficticios, sin eco y con comprobación del resultado SQL sin mostrar el hash. Suite SQL en 26.2.0: 38 comprobaciones correctas. Las ocho migraciones anteriores mantienen bytes y checksums.

Evidencia: [informe de cuenta inicial](../../backend/tests/resultado-cuenta-inicial.json), [informe SQL](../../db/tests/resultado-verificacion.json) y guía [Cuenta inicial](../../backend/CUENTA-INICIAL.md). Se corrigió la preparación inicial de tests para respetar commits DDL del migrador y comentarios; la ejecución final pasa. CI remota se registra en el PR, sin presumir integración SQL en ese workflow.

## Consecuencias e impacto

Capítulos 02 (límite Identidad), 03 (nueva tabla/migración y datos), 04 (árbol, consola y proceso), 05 (hash, auditoría, concurrencia/pruebas), 07 (ADR/CHG) y portada. Glosario: conservar cuenta/usuario SQL/persona/profesional/rol separados; sin cambio semántico clínico. Contrato HTTP health sin cambios; no nuevo endpoint ni contrato de login todavía.

Permisos: 009 concede SELECT/UPDATE del control a danber_runtime, sin CREATE ni DELETE. Se usan privilegios existentes de Identidad; los roles de aplicación no son roles SQL. Datos: nueva tabla de control, sin datos personales de pacientes ni contraseñas iniciales. Código: fachada/caso de uso, validaciones, Argon2id, adaptador SQL, CLI y BeginTx de pool. CI compila ambos comandos. SQL exportado y guías local/Notion deben quedar sincronizados y leídos tras editar.

**Qué debe revisarse:** política de contraseña completa (lista de comprometidas, normalización, medición del hash), recuperación, actor técnico del bootstrap, identidad OS frente a autenticación real, permisos por recurso, sesiones y entorno seguro. Los tests no cierran esta revisión: registrar responsable/fecha/evidencia al validarla. Próximo incremento: sesiones, desde main una vez integrado este PR.
