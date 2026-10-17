# Estrategia preliminar de pruebas e independencia

**CHG-008 · 2026-10-17:** migraciones iniciales de CockroachDB 26.2.7 generadas y verificadas: siete versiones, 28 tablas más control de migración, SQL único y scripts de ejecución local. **37 comprobaciones correctas** en nodo aislado con datos ficticios. Autorización/API aún pendientes; las reglas de negocio no confirmadas conservan ese estado. Revisión de impacto pendiente.

[Guía para crear la base local](../../db/README.md). Los antecedentes siguientes conservan su estado histórico.


**CHG-007 · 2026-10-16:** diseño técnico del primer recorrido elaborado por instrucción del usuario. Diccionario con tipos/obligatoriedad, relaciones y restricciones; flujo/transiciones; matriz de permisos y casos de aceptación. Estado del diseño: **Propuesto; revisión de impacto pendiente**. No equivale a migraciones verificadas ni aplicación implementada. Los antecedentes siguientes conservan su historia; para concretar implementación consultar el diseño vigente.

[Leer el diseño técnico vigente](../base-datos/06-diseno-tecnico-pacientes-agenda.md).


Estado: **Propuesto**. Esta entrega contiene documentación; no hay código de aplicación sobre el que ejecutar estas pruebas. Aquí se especifican evidencias que deben producirse al construir cada recorrido.

## 1. Capas de comprobación

| Nivel | Qué prueba | Infraestructura |
|---|---|---|
| Dominio | Intervalos, transiciones, valores monetarios y reglas puras | `testing` de Go, reloj/valores controlados |
| Caso de uso | Permisos, coordinación, idempotencia y respuestas ante fallos | Puertos sustituidos por implementaciones pequeñas |
| Repositorio/adaptador SQL | Consultas, restricciones, mapeo y transacción real | CockroachDB con versión fijada y base efímera |
| Contrato entre módulos | Datos mínimos, estados, errores y garantías | Proveedor real con entorno aislado y consumidores de referencia |
| API | Autenticación, validación, códigos y ausencia de filtraciones | Servidor de prueba y, según escenario, base aislada |
| Componente React | Interacción, etiquetas, foco, errores y conflictos | Vitest y Testing Library |
| Extremo a extremo | Registrar paciente → reservar → consultar agenda | Navegador automatizado, frontend, backend y base de prueba |
| Arquitectura | Imports, propiedad SQL y contratos públicos | Análisis de dependencias y reglas de revisión/CI |

No medir calidad únicamente por cobertura. La prioridad es proteger decisiones que pueden perder datos, duplicar pagos, reservar dos veces o revelar información clínica.

## 2. Cinco ejemplos solicitados

### A. Probar Agenda sin Facturación

Construir el caso de uso con un lector de pacientes, un lector profesional, un reloj y el puerto de almacenamiento de Agenda. Facturación no aparece entre sus dependencias. Probar una reserva válida y una reserva rechazada. Si necesita iniciar emisión tributaria para funcionar, el límite está mal diseñado.

### B. Sustituir dependencias

El sustituto de pacientes devuelve un paciente activo o un error programado. El sustituto profesional informa no elegibilidad. El almacenamiento de prueba captura la intención guardada o simula un conflicto de versión. Comprobar que ante paciente inválido no se intenta persistir y ante falla de persistencia no se comunica éxito.

No usar estos sustitutos para «demostrar» ausencia de dobles reservas: su comportamiento no reproduce aislamiento, restricciones ni reintentos de la base.

### C. Detectar contratos incompatibles

Si `ObtenerPacienteParaCita` deja de devolver estado o cambia el significado de «activo», una prueba de consumidor comprueba que sigue pudiendo decidir reserva y mostrar identificación mínima. Una firma incompatible rompe compilación; una incompatibilidad semántica debe romper una prueba de comportamiento. Para HTTP, comparar OpenAPI y ejecutar los ejemplos del consumidor. Revisar también si una respuesta nueva expone datos clínicos a recepción.

### D. Evitar que una migración rompa otros módulos

En una base efímera: crear desde cero con toda la secuencia; en otra, aplicar la versión anterior, insertar datos ficticios representativos y migrar a la nueva. Ejecutar consultas y contratos consumidores, comprobar historiales y referencias, y esperar terminación efectiva de cambios de esquema. Para despliegues compatibles, probar la versión anterior de aplicación frente al esquema expandido.

Una eliminación o cambio de tipo compartido requiere estrategia de transición. Reversión solo si preserva datos; no probar recuperación destruyendo una base compartida. Ensayar restauración de una copia de prueba en una instancia nueva.

### E. Dos recepcionistas reservan el mismo intervalo

1. Preparar un odontólogo, horario, fila de control y dos pacientes ficticios.
2. Abrir conexiones independientes a CockroachDB; no compartir una única conexión que serialice artificialmente la prueba.
3. Sincronizar el inicio de dos solicitudes distintas para 09:00–09:30. Mantener claves idempotentes diferentes: son dos intenciones en competencia.
4. Con el mecanismo de reintento habilitado, exigir una reserva confirmada y un conflicto de horario; consultar que exista exactamente una reserva activa y su historial correcto.
5. Repetir con 09:00–10:00 y 09:30–10:30 para descubrir el error de confiar solo en una unicidad de hora inicial.
6. Probar intervalos adyacentes 09:00–09:30 y 09:30–10:00: ambos deben poder confirmarse.
7. Competir reserva contra bloqueo y reserva contra reprogramación. Ninguna escritura de disponibilidad debe evadir la fila de control.
8. Probar cambio entre profesionales en orden inverso desde dos solicitudes: el orden de actualización y los reintentos deben resolver la contención sin pérdida del intervalo anterior.

Ejecutar además varias repeticiones acotadas para detectar carreras; conservar resultados finales y códigos. No hacer depender el éxito de que aparezca siempre exactamente un `40001`: ese detalle depende del orden real de ejecución. La evidencia principal es el estado persistido y el resultado de negocio.

## 3. Casos adicionales que justifican pruebas

- Repetir misma solicitud y clave devuelve el mismo ID; cambiar contenido con esa clave devuelve conflicto.
- Cortar la respuesta después de commit y repetir con la misma clave no genera otra cita.
- Forzar fallos serializables en un entorno aislado verifica repetición completa, límite y ausencia de efectos externos duplicados.
- Reprogramar con versión vieja conserva la edición más reciente.
- Usuario de recepción no obtiene antecedentes por endpoint, exportación o consulta compuesta.
- Odontólogo sin relación/autorización sobre el recurso no lee una historia ajena según la política acordada.
- Cierre de sesión, cuenta suspendida y cambio de permisos invalidan acceso efectivo.
- Acceso y corrección clínica generan auditoría y no sobrescriben finalizados, cuando se incorpore Clínica.
- Aplicaciones concurrentes no exceden pago ni cargo; un evento repetido no crea dos deudas, cuando se incorpore Economía.
- En frontend, un conflicto conserva el formulario, actualiza disponibilidad y no muestra éxito optimista engañoso.

## 4. Separación de ambientes y protección de producción

No basta llamar a una variable `TEST_DATABASE_URL`. Propuesta de controles combinados:

1. Instancia/cluster efímero local o entorno de CI dedicado, credenciales distintas y base por ejecución con identificador único.
2. CI y ejecutores de prueba sin secretos ni acceso de red a producción; repositorios de backups y almacenamiento también separados.
3. El arnés valida host/cluster permitido, base esperada y una marca de entorno antes de migrar o limpiar; falla si falta configuración y jamás usa por defecto la conexión ordinaria.
4. La identidad SQL de prueba no tiene permisos sobre otras bases. Limpieza restringida a recursos creados por esa ejecución y verificados por nombre/ID.
5. Datos sintéticos con identidades reconocibles. No copiar historias reales a snapshots, fixtures o logs. Los formatos vacíos sirven de referencia visual, no de datos clínicos de prueba.

Estos controles deben fallar de manera cerrada y probarse también con una configuración incorrecta. Un filtro textual de nombre de base, por sí solo, no garantiza aislamiento.

## 5. CI y ritmo de trabajo

Por cambio: formato, compilación, análisis estático, tipos de TypeScript, pruebas unitarias y componentes, verificación de imports/contratos. Si afecta persistencia: integración con la imagen exacta de CockroachDB y migración desde estado anterior. Si afecta reserva o autenticación: pruebas de concurrencia o seguridad pertinentes. En el recorrido vertical: un conjunto pequeño de pruebas de navegador.

Comandos de referencia para cuando exista código: `go test ./...`, `go vet ./...` y pruebas frontend mediante scripts versionados. Las pruebas de carrera de Go se ejecutarán en un entorno compatible; detectan carreras de memoria, no sustituyen la prueba de concurrencia SQL. Mantener un reporte de versión de motor, migraciones aplicadas y pruebas ejecutadas.

Las pruebas se incorporan junto al caso de uso, no al terminar toda la aplicación. Evitar pruebas que solo confirman nombres de carpetas o repiten la implementación. Para este cambio documental se revisan enlaces, consistencia y diagramas, sin inventar resultados de pruebas funcionales.


