# ADR 0010 — flujo de ramas para un desarrollador

Fecha: 2026-10-17. Estado vigente: **Adoptado el 2026-10-17**, CHG-011, al autorizar el usuario la siguiente etapa en feat/backend-inicial. La explicación inferior conserva la propuesta original; protecciones siguen pendientes de configuración. Revisión de impacto: **Pendiente de revisión**. CHG-010. Origen: el usuario solicita aprender buenas prácticas de ramas antes de crear el backend. No se crean ramas, PR, protecciones ni automatizaciones en esta entrega.

## Recomendación y contexto

Para Danber, un desarrollador y un repositorio para documentación, SQL y futura aplicación: main estable y una rama breve por cambio, con Pull Request hacia main. Es una adaptación de GitHub Flow. Estable significa que cumple las comprobaciones del incremento actual; todavía no existe backend ejecutable y main no significa «desplegado en producción».

Una rama es una línea de cambios aislada dentro del mismo repositorio. Crear feat/backend-inicial permite trabajar en el servidor sin incorporar avances incompletos a main. La alternativa es modificar main directamente; facilita empezar pero mezcla trabajo en curso con lo integrado. Una rama no crea otra base CockroachDB ni otro repositorio.

No crear develop permanente, ramas por persona, frontend o backend. Las capacidades organizan carpetas; las ramas organizan cambios. Una entrega puede incluir Go, migración, documentación y pruebas en la misma rama cuando persiguen un objetivo.

## Reglas y motivos propuestos

| Práctica | Qué resuelve y por qué importa | Riesgo de omitirla |
|---|---|---|
| main cumple comprobaciones y no recibe trabajo incompleto | Ofrece una base conocida desde donde continuar | No saber qué versión se puede ejecutar |
| Una rama por objetivo pequeño | Permite revisar, terminar o revertir por separado | Mezclar login y reservas dificulta descubrir errores |
| Ramas breves, integrar al terminar | Reduce divergencia y conflictos | Semanas de cambios acumulados son difíciles de revisar |
| Commits con intención concreta | Explican qué cambió y permiten recuperar avances | Mensajes «cambios» no ayudan a investigar |
| PR incluso trabajando solo | Centraliza problema, diff, comprobaciones y límites | Integrar sin revisar el conjunto |
| Revisar el diff y ejecutar checks proporcionales | Encuentra archivos accidentales y fallos reales | Código que solo funciona en un equipo |
| Actualizar Notion si cambia una decisión | Une implementación y razones | Tener dos explicaciones vigentes incompatibles |
| No reescribir main con force push | Conserva historia integrada | Perder referencias y dificultar recuperación |
| Retirar rama al integrarla | Señala que terminó y evita reutilizarla | Continuar sobre una rama obsoleta |
| Migraciones aplicadas inmutables | Conserva checksums e historia de datos | Cambiar una rama rompe instalaciones existentes |

Rama breve no significa un plazo rígido. Si una tarea crece, dividir en incrementos completos: servidor/conexión → cuenta inicial → sesiones → alta de pacientes → reserva. Ejemplo: feat/backend-inicial debe terminar con arranque, configuración, conexión y comprobaciones, sin incluir toda Agenda.

## Nombres orientativos

- feat/backend-inicial: nueva capacidad o incremento del backend.
- feat/registro-pacientes: alta y búsqueda.
- fix/solapamiento-citas: corrección de un fallo.
- docs/flujo-ramas: documentación.
- chore/configuracion-ci: mantenimiento de herramientas.

Los prefijos facilitan leer; no son una exigencia de Git. Mantener un vocabulario consistente y nombres breves sin espacios. Mensajes propuestos de commit: «Agrega comprobación de conexión a CockroachDB» o «Rechaza reserva que cruza el almuerzo». No exigir Conventional Commits ni automatizar releases hasta que exista una necesidad.

## Ciclo concreto (ejemplo para ejecutar después)

Antes de cambiar de rama, revisar git status. Si hay cambios ajenos o una entrega en curso, resolverlos explícitamente; no hacer reset ni stash automático.

```powershell
git status
git switch main
git pull --ff-only origin main
git switch -c feat/backend-inicial

# Implementar el objetivo, revisar diff y ejecutar sus comprobaciones.
git diff

# Agregar archivos concretos una vez revisados; no incluir secretos ni datos locales.
git add backend README.md
git commit -m "Agrega servidor inicial y conexion a CockroachDB"
git push -u origin feat/backend-inicial
```

Estos comandos son un ejemplo: backend todavía no existe. Agregar también documentación/Notion y pruebas afectadas según el cambio real. El primer commit inicial ya hecho directamente en main es una base histórica válida; no reescribirlo para simular este flujo.

Abrir PR en GitHub con base main y comparación feat/backend-inicial. Explicar problema, resultado, cómo se comprobó y limitaciones. Revisar Files changed, permisos, ausencia de secretos y estado de checks. Abrir Draft si aún está en curso. Propuesta inicial: Merge commit para conservar los commits del incremento y su integración de forma visible; squash es alternativa válida si después se prefiere resumir cada PR en un commit.

Después de integrar por Merge commit:

```powershell
git switch main
git pull --ff-only origin main
git branch -d feat/backend-inicial
```

Eliminar la rama remota integrada desde el PR. Git conserva cambios integrados y la discusión del PR. Si se eligió squash, -d puede rechazar la eliminación local porque la historia no contiene los mismos commits: verificar el resultado antes de decidir limpieza, no convertir automáticamente en -D.

Si main avanzó mientras trabajabas: git fetch origin y, estando en tu rama de trabajo, git merge origin/main; resolver conflictos y repetir comprobaciones afectadas. No hace falta enseñar rebase y force push como requisito del primer incremento.

## Pull Request, CI y revisión en solitario

Un PR propone integrar una rama y muestra todo el cambio. La autorrevisión es útil para aprender y descubrir errores; no equivale a revisión independiente ni exige que otro usuario apruebe un proyecto de una sola persona.

CI ejecuta comprobaciones en un entorno reproducible al subir cambios. Cuando exista Go: formato, go vet, pruebas y compilación; para persistencia, integración contra CockroachDB fijado. Frontend añadirá sus propios checks cuando exista. No crear tests de una aplicación inexistente ni declarar una comprobación obligatoria que no se ha configurado.

Proponer protección de main sin force push, con PR y checks disponibles cuando se implementen. No exigir una aprobación externa inexistente. Disponibilidad y alcance de protecciones dependen de la visibilidad y plan de GitHub; no se verificó el plan ni se activaron reglas. Hasta configurarlas, son una disciplina acordable, no una barrera técnica activa.

## Entornos, versiones y errores

Una rama comparte la base local que señale su configuración. Para probar migraciones de una rama usar una base aislada/sintética; git switch no deshace SQL ya aplicado. No modificar una migración aplicada; añadir la siguiente.

Si un cambio integrado falla, preparar rama fix y PR, o revertir el cambio de código con otro commit tras revisar impacto. Revertir Git no revierte datos automáticamente. Para SQL, analizar el estado y usar corrección hacia adelante o restauración verificada según el caso.

Cuando exista una versión que se quiera identificar, usar una etiqueta de versión sobre un commit integrado. No crear rama release o hotfix permanente por anticipado. Git Flow con develop/release/hotfix se reconsidera si se mantienen varias versiones o ciclos de liberación distintos; añade sincronización que hoy no aporta una necesidad concreta.

## Impacto y verificación

Afecta README/guía de trabajo, capítulo 04 y capítulo 07 de Notion, y futuras prácticas de commits/PR/CI. No modifica reglas de negocio, contratos, tablas ni pruebas de aplicación. Propuesta documentada; responsables de adopción y revisión: usuario y asistente en la próxima tarea autorizada.

Verificar enlaces y lectura posterior de Notion. No publicar ni configurar protecciones por este documento. La próxima rama sugerida es feat/backend-inicial una vez adoptado el flujo y autorizada esa implementación.

Fuentes: [GitHub Flow](https://docs.github.com/en/get-started/using-github/github-flow), [protección de ramas](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches).


