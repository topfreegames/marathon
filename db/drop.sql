-- marathon
-- https://github.com/topfreegames/marathon
-- Licensed under the MIT license:
-- http://www.opensource.org/licenses/mit-license
-- Copyright © 2016 Top Free Games <backend@tfgco.com>

REVOKE ALL ON SCHEMA public FROM marathon;
DROP DATABASE IF EXISTS marathon;
DROP ROLE marathon;

CREATE ROLE marathon LOGIN
  SUPERUSER INHERIT CREATEDB CREATEROLE;

CREATE DATABASE marathon
  WITH OWNER = marathon
       ENCODING = 'UTF8'
       TABLESPACE = pg_default
       TEMPLATE = template0;

GRANT ALL ON SCHEMA public TO marathon;
