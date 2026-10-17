-- Backend inicial: el rol public no debe permitir DDL al usuario de ejecución.
-- No modifica tablas ni datos; propietario y administrador conservan sus permisos.
REVOKE CREATE ON SCHEMA public FROM public;
