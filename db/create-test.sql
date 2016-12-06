CREATE ROLE marathon_test_user LOGIN
  SUPERUSER INHERIT CREATEDB CREATEROLE;

CREATE DATABASE marathon_test
  WITH OWNER = marathon_test_user
       ENCODING = 'UTF8'
       TABLESPACE = pg_default
       TEMPLATE = template0;
