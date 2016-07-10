-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE organizations (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name VARCHAR(200) NOT NULL CHECK (name <> ''),
  created_at BIGINT NOT NULL,
  updated_at BIGINT NULL
);
CREATE UNIQUE INDEX unique_organization_name ON organizations ((lower(name)));

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE organizations;
