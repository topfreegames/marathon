-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE templates (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name varchar(255) NOT NULL CHECK (name <> ''),
  locale varchar(2) NOT NULL CHECK (locale <> ''),
  defaults JSONB NOT NULL DEFAULT '{}'::JSONB,
  body JSONB NOT NULL DEFAULT '{}'::JSONB,
  created_at bigint NOT NULL,
  updated_at bigint NULL
);
CREATE UNIQUE INDEX unique_template_name_locale ON templates (lower(name),(lower(locale)));

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE templates;
