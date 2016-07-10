-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE apps (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  organization_id UUID NOT NULL REFERENCES organizations (id),
  name varchar(200) NOT NULL CHECK (name <> ''),
  app_group varchar(200) NOT NULL CHECK (name <> ''),
  created_at bigint NOT NULL,
  updated_at bigint NULL
);
CREATE UNIQUE INDEX unique_app_name ON apps ((lower(name)));
CREATE INDEX app_group ON apps (lower('group'));

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE apps;
