-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE apps (
    id varchar(36) PRIMARY KEY,
    organization_id varchar(36) NOT NULL REFERENCES organizations (id),
    name varchar(200) NOT NULL,
    app_group varchar(200) NOT NULL,
    created_at bigint NOT NULL,
    updated_at bigint NULL,
    deleted_at bigint NOT NULL,

    CONSTRAINT unique_name UNIQUE(name)
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE apps;
