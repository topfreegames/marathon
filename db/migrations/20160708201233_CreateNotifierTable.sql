-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE notifiers (
    id varchar(36) PRIMARY KEY,
    app_id varchar(36) NOT NULL REFERENCES apps (id),
    service varchar(5) NOT NULL,
    created_at bigint NOT NULL,
    updated_at bigint NULL,

    CONSTRAINT unique_notifier_app_service UNIQUE(app_id, service)
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE notifiers;
