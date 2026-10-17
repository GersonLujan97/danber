# Guía de conceptos aplicada al consultorio

Esta guía acompaña las fases A–C y las propuestas preliminares. Cada ficha explica significado, problema, utilidad, alternativa, consecuencia y ejemplo. Los patrones son herramientas; no son carpetas obligatorias.

## Entender el negocio

### 1. Dominio y DDD

**Qué significa:** dominio es la actividad que el software ayuda a realizar; DDD organiza el diseño alrededor de su conocimiento y lenguaje. **Problema:** un modelo construido solo desde pantallas o tablas suele perder reglas. **Relevancia:** Danber combina clínica, agenda y dinero. **Alternativa:** una aplicación CRUD, centrada en crear, leer, actualizar y borrar registros, basta para catálogos simples. **Sin este criterio:** un campo «estado» puede mezclar aceptación, ejecución y pago. **Ejemplo:** conocer cuándo un tratamiento termina antes de diseñar `finalizarTratamiento`.

### 2. Subdominio y clasificación estratégica

**Qué significa:** una parte del problema; core concentra conocimiento diferenciador, supporting lo apoya y generic resuelve necesidades comunes. **Problema:** dedicar el mismo esfuerzo de diseño a todo. **Relevancia:** prioriza continuidad clínica frente a inventar un mecanismo de contraseñas. **Alternativa:** tratar todo como CRUD uniforme. **Sin clasificar:** se puede invertir mucho en infraestructura y poco en reglas clínicas. **Ejemplo:** diseñar cuidadosamente seguimiento de ortodoncia y usar algoritmos criptográficos establecidos.

### 3. Bounded context, módulo y context map

**Qué significa:** un contexto delimita un modelo; un módulo es su organización de código; un mapa muestra relaciones entre modelos. **Problema:** palabras y datos con significados incompatibles. **Relevancia:** «estado» de Cita y de Pago son cosas diferentes. **Alternativa:** un modelo único. **Sin límites:** todos cambian las mismas estructuras. **Ejemplo:** Clínica publica una referencia a un procedimiento y Economía decide si origina cargo. Un contexto puede tener varios paquetes; no es obligatoriamente un servicio de red.

### 4. Lenguaje ubicuo

**Qué significa:** vocabulario acordado entre personas y software. **Problema:** sinónimos ambiguos. **Relevancia:** «Costo» ya se aclaró como importe del procedimiento. **Alternativa:** conservar nombres distintos en cada pantalla. **Sin acuerdo:** un informe podría tratar ese importe como dinero recibido. **Ejemplo:** usar `Pago` para el abono y `ProcedimientoRealizado` para el hecho clínico.

## Modelar reglas

### 5. Entidad

**Qué significa:** objeto cuya identidad continúa aunque cambien atributos. **Problema:** distinguir continuidad de igualdad de datos. **Relevancia:** el paciente sigue siendo el mismo al actualizar el teléfono. **Alternativa:** identificar por nombre o documento. **Sin identidad estable:** correcciones generan duplicados o referencias rotas. **Ejemplo:** `PacienteID` estable y documento corregible; dos pacientes pueden compartir teléfono.

### 6. Objeto de valor

**Qué significa:** valor definido por su contenido, normalmente inmutable. **Problema:** repetir validaciones y combinar datos incompatibles. **Relevancia:** dinero necesita moneda e importe; un intervalo necesita inicio y fin. **Alternativa:** números y cadenas sueltos. **Sin encapsular:** se puede sumar monedas distintas o aceptar fin anterior al inicio. **Ejemplo:** `Dinero(100.00, PEN)` y `Intervalo(09:00, 09:30)`. No crear un tipo complejo para cada texto si no protege una regla.

### 7. Invariante

**Qué significa:** condición que debe mantenerse tras una operación válida. **Problema:** cambios individualmente plausibles que dejan un estado incorrecto. **Relevancia:** el importe aplicado no puede superar lo disponible en un pago. **Alternativa:** comprobaciones dispersas en formularios. **Sin regla central:** otra API puede saltársela. **Ejemplo:** toda reserva ordinaria de un odontólogo debe quedar sin solapamientos, incluso con dos recepcionistas.

### 8. Agregado y raíz de agregado

**Qué significa:** grupo pequeño de objetos cuya consistencia se protege a través de una entrada, la raíz. **Problema:** modificar partes sin aplicar reglas del conjunto. **Relevancia:** Cita controla transiciones e historial; Pago controla su estado. **Alternativa:** editar cada fila libremente. **Sin este límite:** una cita cancelada puede confirmarse sin una regla definida. **Ejemplo:** `Cita.Reprogramar` valida su transición, pero la exclusión con otras citas requiere además coordinación transaccional. Historia completa no será un agregado gigantesco que haya que cargar para cada nota.

### 9. Repositorio

**Qué significa:** contrato de persistencia para un agregado o necesidad de aplicación. **Problema:** reglas mezcladas con consultas. **Relevancia:** guardar Cita sin introducir el driver SQL en su lógica. **Alternativa:** SQL en el caso de uso; aceptable en un registro muy simple si se controla el acoplamiento, menos adecuado para la reserva concurrente. **Sin separación cuando hace falta:** probar reglas exige una base. **Ejemplo:** guardar una cita y su historial dentro de la transacción del caso de uso. No crear un repositorio genérico CRUD por cada tabla.

### 10. Servicio de dominio y servicio de aplicación

**Qué significa:** el primero expresa una regla que no encaja en una entidad; el segundo coordina un objetivo del usuario. **Problema:** poner todo en una entidad enorme o en el controlador. **Relevancia:** calcular disponibilidad difiere de autorizar, consultar y reservar. **Alternativa:** funciones simples; a menudo son precisamente la implementación adecuada en Go. **Sin distinguir:** una regla acaba enviando HTTP o gestionando sesiones. **Ejemplo:** `EvaluarSolapamiento` puede ser una función pura; `ReservarCita` coordina validación, transacción e historial.

## Organizar código y comunicación

### 11. Monolito modular y microservicio

**Qué significa:** un backend desplegado como unidad con límites internos, frente a servicios desplegados por separado. **Problema:** mantener independencia funcional sin costo operativo excesivo. **Relevancia:** equipo pequeño y dominio en aprendizaje. **Alternativa:** microservicios desde el inicio. **Sin modularidad:** una aplicación única puede convertirse en código fuertemente acoplado; sin microservicios no existe por sí mismo un problema inicial. **Ejemplo:** Agenda y Clínica son paquetes distintos del mismo binario Go.

### 12. Dependencia, puerto y adaptador

**Qué significa:** depender es necesitar un contrato o detalle; un puerto declara una necesidad; un adaptador la satisface. **Problema:** reglas ligadas al framework o la base. **Relevancia:** reservar sin conocer pgx ni React. **Alternativa:** usar dependencias concretas directamente donde no hay reglas o necesidad de sustitución. **Sin invertir las dependencias importantes:** cambiar persistencia obliga a tocar dominio. **Ejemplo:** Agenda pide `ObtenerPacienteParaCita`; un adaptador llama Pacientes y otro devuelve un paciente ficticio en pruebas. En Go una interfaz pequeña puede cumplirse sin una declaración explícita de herencia.

### 13. Composición e inyección de dependencias

**Qué significa:** construir piezas y entregarles lo que necesitan desde el arranque. **Problema:** componentes que crean conexiones o acceden a variables globales por su cuenta. **Relevancia:** sustituir un lector de pacientes durante una prueba. **Alternativa:** localizadores globales o crear dependencias dentro de cada función. **Sin composición explícita:** resulta difícil conocer el conjunto necesario para ejecutar un caso. **Ejemplo:** `main` conecta almacenamiento, reloj y servicio de Agenda. No hace falta un contenedor de inyección ni reflexión.

### 14. DTO, contrato y referencia por identificador

**Qué significa:** un DTO transporta datos; un contrato define forma, significado, errores y garantías; una referencia enlaza identidades sin copiar objetos completos. **Problema:** filtrar datos y depender de estructuras ajenas. **Relevancia:** Agenda necesita identificar a alguien, no conocer alergias. **Alternativa:** compartir la entidad Paciente completa. **Sin contrato limitado:** agregar o cambiar un dato clínico afecta consumidores administrativos. **Ejemplo:** `PacienteParaCita{id, nombreVisible, estado}` y `paciente_id` en la reserva. Un DTO no contiene la lógica de la entidad.

### 15. Comando, consulta y CQRS

**Qué significa:** un comando intenta cambiar estado; una consulta devuelve información; CQRS separa modelos de escritura y lectura en distintos grados. **Problema:** tratar toda operación como editar una fila. **Relevancia:** `ReprogramarCita` debe proteger reglas; `ConsultarAgenda` combina información. **Alternativa:** CRUD sencillo para catálogos. **Sin distinguir intención:** un `PATCH` genérico puede saltarse historial. **Ejemplo:** funciones distintas para reservar y listar usando la misma base. No se propone un framework CQRS, una base de lectura separada ni mensajería por este motivo.

### 16. Evento de dominio, evento de integración y outbox

**Qué significa:** un evento describe un hecho; el de dominio sirve dentro del módulo, el de integración tiene un contrato para otros; outbox guarda tareas de publicación junto al cambio de negocio. **Problema:** perder efectos posteriores cuando el proceso cae. **Relevancia:** una atención finalizada debe actualizar seguimiento aunque el proceso se reinicie. **Alternativa:** llamada directa si se necesita resultado inmediato o no hay trabajo posterior durable. **Sin persistencia del trabajo necesario:** el segundo módulo puede no enterarse. **Ejemplo:** guardar atención y tarea, después comunicar `AtencionFinalizada.v1`. No significa ejecutar todo mediante eventos.

### 17. Proyección, copia histórica y capa anticorrupción

**Qué significa:** proyección es información derivada para leer; copia histórica conserva datos como eran; capa anticorrupción traduce un modelo externo. **Problema:** consultas costosas, documentos que cambian retrospectivamente y dependencia del proveedor. **Relevancia:** son necesidades distintas que no conviene mezclar. **Alternativa:** leer siempre datos actuales y usar directamente estados externos. **Sin diferenciarlas:** una factura antigua podría cambiar con el nombre del catálogo. **Ejemplo:** proyección del saldo actual, descripción congelada en presupuesto y traducción del estado de emisión del proveedor. Ninguna copia autoriza modificar el original de otro módulo.

## Consistencia y calidad

### 18. Transacción y aislamiento serializable

**Qué significa:** transacción agrupa cambios que se confirman o abortan juntos; serializable busca un resultado equivalente a algún orden secuencial de transacciones. **Problema:** observar disponibilidad y reservar sobre datos que otro usuario cambió. **Relevancia:** dos recepcionistas compiten por un odontólogo. **Alternativa:** menor aislamiento con una estrategia de bloqueo correcta y específica. **Sin un protocolo completo:** ambas reservas pueden parecer válidas. **Ejemplo:** validar intervalo y guardar cita en una transacción; si CockroachDB requiere reintento, repetir la unidad completa. La propiedad técnica y el manejo de errores se verifican en la [referencia oficial](https://docs.cockroachlabs.com/docs/v26.2/transaction-retry-error-reference).

### 19. Concurrencia optimista e idempotencia

**Qué significa:** versión optimista detecta que alguien cambió un registro; idempotencia evita duplicar el efecto de la misma solicitud. **Problema:** edición obsoleta y repetición tras problemas de red. **Relevancia:** son riesgos distintos. **Alternativa:** bloquear edición o reconciliar manualmente duplicados. **Sin protección:** una reprogramación pisa otra o un abono se registra dos veces. **Ejemplo:** `versionEsperada=3` al editar Cita y una clave estable al repetir `RegistrarPago`. La clave no reemplaza validar disponibilidad ni permisos.

### 20. Migración y expansión/contracción

**Qué significa:** cambio versionado del esquema; expansión añade estructura compatible y contracción retira lo viejo después de migrar consumidores y datos. **Problema:** cambios de base que rompen código existente. **Relevancia:** un monolito modular sigue compartiendo despliegue y base. **Alternativa:** editar manualmente tablas. **Sin historial:** ambientes difieren y no se reproduce un fallo. **Ejemplo:** añadir un dato opcional de cita, poblarlo, adaptar lectura y solo después exigirlo. Un script inverso no recupera datos ya eliminados.

### 21. Prueba unitaria, integración, contrato y extremo a extremo

**Qué significa:** prueban respectivamente una regla aislada, colaboración real con infraestructura, promesas entre módulos y un recorrido completo. **Problema:** confundir compilar con funcionar. **Relevancia:** un sustituto en memoria no reproduce concurrencia de CockroachDB. **Alternativa:** pruebas manuales, útiles como complemento. **Sin niveles:** una suite lenta puede omitir reglas críticas. **Ejemplo:** función de solapamiento unitaria; dos conexiones reales compitiendo; contrato Pacientes–Agenda; navegador registrando y reservando.

### 22. Autenticación, autorización, auditoría y trazabilidad

**Qué significa:** verificar identidad, decidir acceso, registrar acciones/accesos y reconstruir cambios. **Problema:** creer que estar conectado permite leer todo. **Relevancia:** recepción y odontólogo necesitan información diferente. **Alternativa:** un acceso general para personal; no es adecuado aquí. **Sin separación:** datos sensibles quedan disponibles sin necesidad. **Ejemplo:** una sesión válida no basta para leer cualquier historia; se verifica permiso y relación asistencial, y se registra quién accedió. Auditoría de lectura y evolución clínica son registros diferentes.

### 23. Estado local, caché del servidor y validación de formulario

**Qué significa:** estado local pertenece a la interacción; caché conserva temporalmente datos obtenidos; validación comprueba formato y reglas. **Problema:** duplicar la agenda en varios almacenes del navegador. **Relevancia:** una disponibilidad en pantalla puede estar desactualizada. **Alternativa:** cargar manualmente todo con efectos y variables; funciona al principio pero multiplica manejo de errores. **Sin distinguir:** la interfaz puede mostrar una reserva no confirmada como real. **Ejemplo:** selección de fecha local, agenda consultada con TanStack Query y reserva confirmada exclusivamente por backend.

### 24. ADR y MVP vertical

**Qué significa:** ADR registra una decisión, motivos y consecuencias; MVP es el alcance mínimo útil; un recorrido vertical atraviesa interfaz, caso de uso y persistencia. **Problema:** perder razones y construir capas enteras que todavía no entregan valor. **Relevancia:** permite aprender con una operación real. **Alternativa:** implementar toda la base, después todo el backend y al final la interfaz. **Sin una entrega vertical:** se descubren tarde problemas de contratos y permisos. **Ejemplo:** registrar paciente, reservar y consultar agenda antes de construir odontograma y facturación completos.

## Patrones que no se incorporan por ahora

**Event sourcing** reconstruye el estado desde eventos como fuente primaria: puede aportar reconstrucción completa, pero añade complejidad de evolución y consultas. Para Danber se proponen tablas actuales, historiales y adendas; no tener event sourcing no impide auditar.

**Saga** coordina pasos distribuidos y compensaciones: sería útil con varios servicios autónomos; por ahora un coordinador local y estados recuperables son suficientes. Compensar no significa borrar un acto clínico ya ocurrido.

**Repositorio genérico, jerarquías de clases y contenedor de inyección** no resuelven una necesidad demostrada en el primer recorrido. Las interfaces concretas se introducen en los límites que sí se probarán o sustituirán.
