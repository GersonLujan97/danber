# ADR 0012 — compatibilidad local con CockroachDB 26.2.0

Fecha: 2026-10-17. Estado: **Confirmado por instrucción del usuario e implementado**. Revisión de impacto: **Pendiente de revisión**. CHG-012. Complementa ADR 0008 y 0011 en selección de versión.

## Problema y decisión

La instalación real del usuario es CockroachDB 26.2.0. El backend y el ejecutor de migraciones exigían exactamente 26.2.7, bloqueando el uso antes de comprobar compatibilidad.

Aceptar una lista explícita de versiones: **26.2.0 y 26.2.7**. Reconocer el motor y el token completo de versión; rechazar PostgreSQL, parches no verificados, sufijos beta y coincidencias parciales como 26.2.70. Mantener el rechazo de conexión SQL administradora.

## Alternativas y consecuencias

Quitar todo chequeo aceptaría versiones no probadas; admitir toda la serie sin pruebas prometería compatibilidad no demostrada. Obligar al usuario a cambiar su instalación no responde a su solicitud actual. La lista explícita amplía solo el alcance verificado.

No se actualiza ni se degrada el servidor o almacén del usuario. Instalador opcional y lanzador siguen ofreciendo 26.2.7 para instalaciones nuevas; un nodo 26.2.0 existente puede mantenerse.

No cambia ninguna instrucción ni checksum de las ocho migraciones aplicables. El SQL único solo cambia su comentario de versiones; no repetirlo sobre una base existente ni recrear danber por este ajuste.

## Verificación e impacto

Go: pruebas unitarias del filtro, vet, build e integración real contra un nodo 26.2.0 aislado en memoria. Conectividad no administradora, rechazo de administrador, probes y cierre HTTP correctos. Suite SQL completo contra 26.2.0, informe en db/tests/resultado-verificacion.json; evidencia anterior 26.2.7 conservada en db/tests/resultado-verificacion-v26.2.7.json.

Binario 26.2.0 oficial verificado con SHA256 596721d964283b6b833b721c510472ae3af0a58d1bd60540f197ef05a25948e8. No se deduce compatibilidad de SQL Server/PostgreSQL ni de otros parches.

Afecta filtro de pool, ejecutor Python, comentario del exportador/SQL único, guías backend/DB, README/AGENTS, portada y capítulos Notion 03/04/07. Sin reglas de negocio nuevas, tablas o API adicionales. Pruebas funcionales verifican el incremento, no uso productivo ni todas las funciones posibles del motor.

Usuario: detener y volver a iniciar la API para cargar el código actualizado. No cambiar permisos ni credenciales por este ajuste. Rama actual feat/backend-inicial y PR #1 aún abiertos; se actualiza ese PR sin merge automático. Estado CI remota informado tras comprobarlo.

