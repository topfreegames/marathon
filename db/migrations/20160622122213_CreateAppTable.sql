-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE apps (
  "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  "organization_id" UUID NOT NULL,
  "name" varchar(200) NOT NULL CHECK (name <> ''),
  "group" varchar(200) NOT NULL CHECK (name <> ''),
  "created_at" bigint NOT NULL,
  "updated_at" bigint NULL
);
CREATE UNIQUE INDEX "index_apps_on_name" ON apps (name);
CREATE INDEX "index_apps_on_group" ON apps ("group");

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS apps;
DROP INDEX IF EXISTS unique_app_name;
DROP INDEX IF EXISTS group;
