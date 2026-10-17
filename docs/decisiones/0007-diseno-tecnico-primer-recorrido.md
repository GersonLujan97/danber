# ADR 0007 — diseño técnico del primer recorrido

Fecha: 2026-10-16. Estado: **Propuesto**; elaboración autorizada por el usuario mediante «Ok, implementalo» tras acordar diccionario técnico, flujo, matriz de permisos y casos de aceptación. Revisión de impacto: **Pendiente de revisión**. Registro CHG-007 en Notion.

## Decisión propuesta

Adoptar el [diseño técnico de Pacientes y Agenda](../base-datos/06-diseno-tecnico-pacientes-agenda.md) como especificación revisable para las siguientes migraciones y el primer recorrido.

Mantener propietario por capacidad, IDs internos independientes, paciente/documento/apoderado/vínculo separados, varios roles por cuenta y Maestro explícito. Agenda conserva resultado de cita, ocupación del intervalo, llegada, espera y tiempos reales diferenciados. La vista muestra estados comprensibles derivados de esos hechos.

Un espacio liberado por ausencia deja de competir por disponibilidad sin borrar la cita original; liberar y asignar otro paciente es una operación atómica. Reprogramación mantiene historia y revierte por completo ante conflicto. Toda escritura de disponibilidad participa del control por profesional del ADR 0002.

## Alternativas y consecuencias

Un único estado para cita, presencia y ocupación simplifica una columna, pero no representa al paciente tardío que ya llegó mientras su intervalo fue usado por otro. Campos separados y transiciones explícitas conservan ambos hechos; requieren más validaciones y una vista derivada.

Una Persona global o duplicar cuenta por rol mezclaría propiedad y permisos. Referencias estables y vínculos dentro de cada capacidad permiten crecer sin introducir tablas globales ni servicios adicionales.

Un diccionario físico asumido como definitivo permitiría empezar SQL más rápido, pero convertiría propuestas de formatos, responsables y catálogos en decisiones silenciosas. Esta entrega deja tipos/longitudes e índices como candidatos y marca lo pendiente antes de fijar SQL.

## Estado y alcance

No cambia confirmaciones del usuario; concreta propuestas de ADR 0002 y 0006. ADR 0003–0005 conservan los cambios parciales ya registrados. Las etiquetas de estados anteriores quedan como antecedentes: para implementación propuesta usar resultado, ocupación y hechos del diseño vigente.

La autorización de elaborar esta entrega no aprueba automáticamente todos los catálogos, restricciones o responsables. La fase D no queda verificada hasta ejecutar sus migraciones y pruebas en CockroachDB real con versión fijada. Esta entrega no incorpora aplicación ni migraciones ejecutables.

## Impacto y revisión

Capítulo 03: diseño completo. Capítulo 04: contratos y matriz de operación enlazados. Capítulo 05: casos de aceptación y alcance de verificación. Capítulo 07: ADR, bitácora y pendientes. Portada y respaldo local actualizados; glosario y contextos enlazan la terminología vigente.

Validar pendientes de sección 11, revisar contratos de liberar/asignar, restricciones de documentos/vínculos y permisos por profesional. Responsable de negocio: usuario; implementación/verificación técnica en fase posterior. Evidencia editorial: lectura posterior de Notion y verificación local de enlaces/consistencia; no equivale a pruebas SQL.


## Verificación editorial — 2026-10-16

El asistente leyó de vuelta la página técnica y cinco páginas enlazadas de Notion; confirmó CHG-007, 30 casos de aceptación y diagrama Mermaid. Revisión local de 11 archivos: 0 enlaces relativos rotos y delimitadores Mermaid completos. Revisión de impacto funcional y SQL pendiente; no se ejecutaron pruebas de aplicación.
