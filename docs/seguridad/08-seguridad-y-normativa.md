# Seguridad y normativa peruana: revisión inicial

**CHG-007 · 2026-10-16:** diseño técnico del primer recorrido elaborado por instrucción del usuario. Diccionario con tipos/obligatoriedad, relaciones y restricciones; flujo/transiciones; matriz de permisos y casos de aceptación. Estado del diseño: **Propuesto; revisión de impacto pendiente**. No equivale a migraciones verificadas ni aplicación implementada. Los antecedentes siguientes conservan su historia; para concretar implementación consultar el diseño vigente.

[Leer el diseño técnico vigente](../base-datos/06-diseno-tecnico-pacientes-agenda.md).


Fecha de consulta: **14 de octubre de 2026**. Estado: controles **propuestos** y obligaciones **pendientes de determinar y validar** para el establecimiento concreto. La investigación identifica referencias oficiales pertinentes; no acredita cumplimiento legal, certificación de historia clínica electrónica ni habilitación para uso con pacientes reales.

## 1. Datos y permisos por necesidad

La identificación administrativa también es dato personal. Separarla de los antecedentes reduce exposición, pero no la convierte en información pública. Un cargo puede revelar indirectamente un tratamiento: los permisos económicos también requieren cuidado.

| Operación | Recepción | Odontólogo | Técnico dental | Administrador |
|---|---|---|---|---|
| Buscar/registrar paciente | Sí, datos mínimos | Según necesidad asistencial | Solo identificación mínima en trabajo asignado | Según función autorizada |
| Reservar/reprogramar | Sí | Si se le asigna esa capacidad | No por defecto | Solo con permiso operativo |
| Leer antecedentes/evolución | No por defecto | Con permiso y relación asistencial validada | Solo información específica necesaria para trabajo autorizado | No por ser administrador |
| Finalizar atención/receta | No | Según atribución profesional y reglas aplicables | No por defecto | No por ser administrador |
| Registrar pago | Solo si se asigna permiso | No por ser odontólogo | No por defecto | Según función económica |
| Cambiar roles y cuentas | No | No por defecto | No | Permiso específico y auditoría |
| Exportar historias/documentos | No por defecto | Operación adicional con alcance | No por defecto | Solo función expresamente autorizada |

Esta matriz es una propuesta para validar, no una política aprobada. La autorización combina **rol/capacidad** con **recurso y contexto**: ser odontólogo no concede acceso automático a todos los pacientes. Debe definirse relación asistencial, sustituciones, derivaciones y cualquier acceso excepcional. Cada excepción tendrá motivo, alcance, registro y revisión; no se implementa un acceso de emergencia genérico sin esa política.

## 2. Autenticación y sesiones

Comparación: identidad administrada por un proveedor reduce mantenimiento de credenciales y puede facilitar MFA, pero añade costos, dependencia e intercambio de datos. Autenticación local evita esa dependencia inicial, pero exige recuperación, sesiones, bloqueo y actualización segura. Propuesta provisional: cuentas internas del personal con sesiones del servidor; revisar proveedor antes de implementar si ya existe una identidad institucional.

Si se guardan contraseñas, usar una implementación mantenida de **Argon2id**, con sal individual y parámetros calibrados/versionados. No implementar criptografía propia ni usar SHA-256 solo para contraseñas. OWASP recomienda algoritmos resistentes a ataques con hardware y ofrece parámetros de referencia que deben adaptarse al entorno. [OWASP Password Storage](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html).

Sesión propuesta: identificador aleatorio opaco, copia como hash en servidor, cookie `Secure`, `HttpOnly` y `SameSite` apropiado. Rotar al autenticar o elevar privilegios; caducidad absoluta e inactividad; revocar al salir, suspender cuenta o recuperar acceso. Definir límites de intentos y mensajes de acceso que no revelen si existe una cuenta. Proponer MFA para acceso privilegiado y remoto, sujeto al diseño operativo.

Servir interfaz y API bajo el mismo origen simplifica cookies. Para operaciones de escritura, aplicar protección CSRF y validar origen; `SameSite` es una defensa complementaria. Evitar tokens de acceso persistidos en `localStorage`. Los límites temporales se acordarán con el consultorio, no se inventan como requisito. [OWASP Session Management](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html).

Recuperación: tokens aleatorios de uso único, expiración corta definida por política, almacenamiento como hash, revocación de sesiones pertinentes, auditoría y canal verificado. Ningún administrador puede consultar la contraseña de otra persona.

## 3. Integridad, confidencialidad y operación

- **Autorización:** comprobar en cada caso de uso y recurso; denegar por defecto; no confiar en IDs, roles o precios enviados por el navegador. Las consultas compuestas y exportaciones aplican las mismas restricciones.
- **Transporte/reposo:** HTTPS, validación de certificados de base, redes restringidas, cifrado de almacenamiento y copias según capacidades del despliegue. Revisar claves y restauración; usar CockroachDB no demuestra por sí solo que todo esté cifrado correctamente.
- **Secretos:** fuera del código, documentación, navegador y logs; identidades separadas de ejecución/migración; rotación y acceso mínimo. No usar cuenta administradora de base en la API.
- **Auditoría:** quién, cuándo, operación, recurso, resultado y correlación; registrar lecturas/exportaciones clínicas además de escrituras. Evitar contenido íntegro de consultas, contraseñas, tokens, DNI y notas en logs operativos. Restringir acceso y alteración de registros de auditoría; sincronizar tiempos y prever revisión.
- **Historia clínica:** borrador editable con control de versión; finalización con autor responsable; correcciones por adenda y motivo. Revisar requisitos profesionales y normativos para autoría/firma; una imagen de firma no los sustituye.
- **Archivos:** almacenamiento privado; autorización en cada descarga, enlaces temporales si se utilizan; límites de tipo/tamaño, nombres seguros y análisis de archivos según exposición. No guardar documentos en rutas públicas del frontend.
- **Navegador:** no persistir expedientes completos por defecto, vaciar caché al cerrar sesión, evitar datos clínicos en URL, telemetría o herramientas de reproducción de sesiones; respuestas sensibles sin almacenamiento compartido.
- **Dependencias:** versiones fijadas, actualizaciones con verificación, análisis de vulnerabilidades y procedencia; retirar librerías que no se usan.

Copias: frecuencia, retención, acceso, cifrado y ubicación por definir con responsables. Proponer una restauración de ensayo antes de producción y ensayos periódicos documentados. **RPO** es la pérdida máxima de datos que se acepta y **RTO** el tiempo objetivo para restaurar servicio: resuelven cuánto respaldo y recuperación necesitamos; sin acordarlos, «tenemos backup» no demuestra que el consultorio pueda continuar. Ejemplo ilustrativo para negociar, no compromiso: tolerar hasta una hora de pérdida y recuperar en cuatro horas exigiría una estrategia y pruebas acordes. Una copia diaria sola no satisface ese RPO.

Incidente: detectar, contener, preservar evidencia, evaluar alcance y obligaciones de comunicación, recuperar y revisar causas. Definir responsables y canal alternativo para el consultorio. Retención clínica, económica, de accesos y de copias tiene finalidades distintas; no aplicar un único borrado automático a todas.

## 4. Referencias peruanas y validaciones necesarias

Las siguientes filas identifican temas a validar, sin afirmar que todas las obligaciones se apliquen del mismo modo a Danber. La revisión especializada debe confirmar textos vigentes, modificatorias y alcance según establecimiento, responsable y forma de operación.

| Tema y referencia oficial | Hallazgo verificable | Qué debe resolverse antes del uso correspondiente |
|---|---|---|
| [Ley 29733](https://www.gob.pe/institucion/congreso-de-la-republica/normas-legales/243470) | Establece el marco de protección de datos personales | Identificar responsable/encargados, finalidades, base de tratamiento, datos sensibles, avisos y ejercicio de derechos |
| [DS 016-2024-JUS, texto publicado](https://www3.congreso.gob.pe/Docs/DGP/DIDP/files/ds_016-2024-jus.pdf) | Aprueba el nuevo reglamento de la ley; contiene disposiciones de seguridad, incidentes y aplicación progresiva de ciertas obligaciones | Revisar registro de bancos, medidas/documentación de seguridad, encargados, transferencias, incidentes y aplicabilidad/calendario de oficial de datos personales |
| [NTS 139 y modificatoria, publicación oficial del INO](https://www.ino.gob.pe/wp-content/uploads/2023/05/NTS-GESTION-HISTORIA-CLINICA-MINSA-2018.pdf) | Compila la norma de gestión de historia clínica aprobada por RM 214-2018 y modificada por RM 265-2018 | Contenido, identificación/numeración, integridad, custodia, acceso, conservación y procedimiento de corrección/eliminación |
| [RM 559-2022-MINSA](https://www.gob.pe/institucion/minsa/normas-legales/3304261-559-2022-minsa) | Aprueba la NTS 188 para uso del odontograma | Validar piezas, dentición, superficies, simbología, colores, trazado e impresión con el anexo y el profesional; no certificar los PNG aportados |
| [Ley 30024, publicación oficial](https://spij.minjus.gob.pe/Normas/textos/220513T.pdf) y [DS 020-2025-SA](https://www.gob.pe/institucion/minsa/normas-legales/7479922-020-2025-sa) | La ley crea RENHICE; el decreto de 2025 modifica el reglamento aprobado por DS 009-2017-SA para adecuarlo a la Ley 31750 | Determinar requisitos aplicables al sistema de historia electrónica, interoperabilidad, acreditación/autorización, acceso y firma; no basarse únicamente en el texto de 2017 |
| [DIGEMID, comunicado 007-2024](https://www.digemid.minsa.gob.pe/Archivos/Comunicados/2024/COM_007-2024.pdf) | Recuerda límites profesionales de prescripción y requisitos de receta; distingue receta electrónica y firma digital | Validar habilitación del prescriptor, campos, firma, impresión, conservación y reglas especiales de medicamentos sujetos a control |
| [SUNAT, tipos de comprobantes](https://cpe.sunat.gob.pe/informacion_general/tipos_comprobantes_pago) | Diferencia factura, boleta, recibo por honorarios y documentos complementarios | Determinar quién emite, RUC/régimen, tipo aplicable, oportunidad de emisión, proveedor/sistema, tratamiento de impuestos, notas y contingencias |

La noticia oficial sobre el nuevo reglamento destaca la protección de datos sensibles y la introducción progresiva del oficial de datos personales. Su alcance y fechas se deben contrastar con la situación de Danber; no se asigna aquí una obligación universal inmediata. [MINJUSDH, explicación del reglamento](https://www.gob.pe/institucion/minjus/noticias/1067368-ejecutivo-aprueba-nuevo-reglamento-de-la-ley-de-proteccion-de-datos-personales).

No se han cotejado todos los símbolos de las imágenes con el anexo del odontograma; por ello su conformidad queda abierta. La consulta de la resolución confirma la norma de referencia, no valida el dibujo concreto.

Consentimiento clínico, aceptación del presupuesto, consentimiento para tratamiento de datos y autorización para mensajes comerciales son decisiones distintas. No crear una casilla que pretenda cubrirlas todas. La base jurídica y las excepciones de cada tratamiento deben validarse; no se afirma que todo uso clínico requiera siempre el mismo consentimiento.

## 5. Decisiones con responsable y evidencia de cierre

| Pendiente | Responsable sugerido | Evidencia para cerrar |
|---|---|---|
| Contenido clínico, odontograma y prescripción | Odontólogo responsable + revisión normativa | Formularios y reglas aprobados, convenciones versionadas |
| Roles, sustituciones, derivaciones y menores | Dirección clínica y administración | Matriz por operación/recurso y ejemplos de acceso permitido/denegado |
| Privacidad, retención, incidentes y transferencias | Responsable del tratamiento + asesoría especializada | Política aplicable, contratos, registro de decisiones y procedimientos |
| Historia electrónica y firma | Dirección clínica + asesoría especializada | Requisitos de validez/autoría y condiciones de uso documentados |
| Tributación y comprobantes | Responsable contable/tributario | Flujo de emisión y corrección validado para el emisor real |
| Recuperación y operación | Responsable técnico + dirección | Objetivos acordados y restauración probada |

El primer recorrido usa datos ficticios hasta disponer de controles técnicos probados y validaciones aplicables al alcance. La validación de un MVP administrativo no autoriza automáticamente a sustituir la historia clínica oficial por un módulo clínico futuro.

