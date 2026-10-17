# ADR 0001 — monolito modular y límites de negocio

Propuesto: 2026-10-14. Estado: **Aceptado provisionalmente el 2026-10-16**. Revisión de impacto: **Pendiente de revisión**.

CHG-002 en Notion: el usuario leyó la documentación y expresó «me parece por el momento correcto». Se registra conformidad con la orientación arquitectónica; las reglas abiertas y la validación técnica siguen pendientes. Se conserva abajo el razonamiento de la propuesta original.

## Contexto

Danber inicia un sistema para un consultorio odontológico, con un equipo pequeño y reglas por descubrir. Las imágenes reúnen identificación, clínica, seguimiento e importes en documentos compartidos. El usuario solicita aprender las decisiones antes de crear toda la aplicación.

## Decisión propuesta

Backend Go desplegado como una unidad, frontend React/TypeScript y CockroachDB. Ocho contextos: Identidad y Acceso, Personal, Pacientes, Agenda, Clínica, Tratamientos, Economía y Facturación. Implementar solo los necesarios para cada recorrido. Separar reglas de HTTP y SQL mediante puertos pequeños donde haya una necesidad real de aislamiento o prueba.

Recetas pertenece inicialmente a Clínica. Controles recomendados pertenecen a Tratamientos; reserva a Agenda; resultado a Clínica. Tarifario, Presupuestos y Cobranzas son capacidades diferentes dentro de Economía. Facturación tiene límite propio y se implementará después. Administración de cuentas pertenece a Identidad; datos laborales/profesionales a Personal.

## Alternativas

Monolito organizado solo por tablas/pantallas: menos estructura inicial, más mezcla de reglas y permisos. Microservicios: independencia de despliegue a cambio de red, consistencia distribuida y operación que hoy no se justifican. Unir Clínica y Tratamientos: viable si el trabajo real es simple, pero debilita la distinción entre intención longitudinal y hechos clínicos.

## Consecuencias

Facilita trabajo por capacidades y pruebas aisladas, mantiene llamadas locales y despliegue simple. Exige contratos, revisión de imports y disciplina SQL; la base y el proceso siguen siendo compartidos. No promete extracción gratuita a servicios. El dominio no depende de drivers, HTTP ni componentes de interfaz.

## Señales para revisar

Equipos autónomos, necesidad de despliegue independiente, carga medida concentrada en una capacidad, integración externa compleja o evidencia de que dos límites cambian siempre juntos. Un aumento de carpetas por sí solo no es evidencia.

Detalle: [arquitectura](../arquitectura/03-propuesta.md), [contextos](../dominio/02-contextos.md).

