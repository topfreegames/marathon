-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE organizations (
    id varchar(36) PRIMARY KEY,
    name varchar(200) NOT NULL,
    created_at bigint NOT NULL,
    updated_at bigint NULL,

    CONSTRAINT unique_organization_name UNIQUE(name)
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE organizations;
