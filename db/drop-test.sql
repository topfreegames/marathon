REVOKE ALL ON SCHEMA public FROM marathon_test;
DROP DATABASE IF EXISTS marathon_test;
DROP ROLE marathon_test;

CREATE ROLE marathon_test LOGIN
  SUPERUSER INHERIT CREATEDB CREATEROLE;

CREATE DATABASE marathon_test
  WITH OWNER = marathon_test
       ENCODING = 'UTF8'
       TABLESPACE = pg_default
       TEMPLATE = template0;

GRANT ALL ON SCHEMA public TO marathon_test;
