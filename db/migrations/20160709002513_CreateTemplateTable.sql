-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE templates (
  "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  "name" varchar(255) NOT NULL CHECK (name <> ''),
  "locale" varchar(2) NOT NULL CHECK (locale <> ''),
  "defaults" JSONB NOT NULL DEFAULT '{}'::JSONB,
  "body" JSONB NOT NULL DEFAULT '{}'::JSONB
);
CREATE UNIQUE INDEX "index_templates_on_name_and_locale" ON templates (name,locale);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS templates;
DROP INDEX IF EXISTS index_templates_on_name_and_locale;
