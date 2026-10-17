# ADR 0009 — repositorio personal y saltos de línea

Fecha: 2026-10-17. Estado: **Confirmado por el usuario** para cuenta personal y remoto; **Implementado** para política de archivos. Revisión de impacto: **Pendiente de revisión**. CHG-009.

## Contexto y decisión

El usuario inicializó main, configuró autor Isai con correo personal y cuenta GitHub GersonLujan97, y estableció origin en https://github.com/GersonLujan97/danber.git. El primer push falló con «src refspec main does not match any» porque la rama todavía no tenía commits. La configuración global de trabajo no se modifica.

Preparar un commit inicial de documentación, formatos y migraciones; comprobar el envío al remoto por separado. No incluir .local, .env, binarios ni datos del nodo.

Agregar .gitattributes con SQL, Python y PowerShell en LF y PNG binario. Git no debe convertir los SQL a CRLF en un checkout de Windows, porque el ejecutor compara SHA256 de bytes y una conversión alteraría migraciones ya registradas.

## Alternativas y consecuencias

Cambiar la identidad global afectaría otros proyectos y contradice el alcance personal de Danber. Usar identidad local conserva ambas cuentas.

Dejar la conversión de fin de línea dependiente de cada equipo rompería checksums; recalcularlos sobre texto normalizado cambiaría el contrato del ejecutor. Se conservan los bytes originales y se fija LF para archivos ejecutables versionados.

## Impacto y comprobaciones

Afectados: .gitattributes, respaldo local, portada y capítulo 07 de Notion. Las migraciones mantienen su contenido y versión. Comparar bytes de cada SQL en disco y en el índice de Git antes del commit. No requiere repetir pruebas SQL por una política de checkout sin cambios de instrucciones.

Revisión de impacto pendiente para futuros clones/checkout; verificar autor/remoto antes de publicar y no interpretar credencial guardada como prueba de permisos. No se registra contraseña ni token. Estado del envío remoto se informa tras su ejecución; este ADR no lo presume completado.

