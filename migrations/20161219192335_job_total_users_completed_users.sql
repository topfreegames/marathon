
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE "jobs" ADD COLUMN completed_users integer NOT NULL DEFAULT 0;
ALTER TABLE "jobs" ADD COLUMN total_users integer NOT NULL DEFAULT 0;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE "jobs" DROP COLUMN completed_users;
ALTER TABLE "jobs" DROP COLUMN total_users;
