
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE "jobs" RENAME COLUMN completed_users TO completed_tokens;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE "jobs" RENAME COLUMN completed_tokens TO completed_users;
