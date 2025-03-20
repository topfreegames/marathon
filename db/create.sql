CREATE ROLE marathon_user LOGIN
  SUPERUSER INHERIT CREATEDB CREATEROLE;

CREATE DATABASE marathon
  WITH OWNER = marathon_user
       ENCODING = 'UTF8'
       TABLESPACE = pg_default
       TEMPLATE = template0;

CREATE DATABASE push
  WITH OWNER = marathon_user
       ENCODING = 'UTF8'
       TABLESPACE = pg_default
       TEMPLATE = template0;
