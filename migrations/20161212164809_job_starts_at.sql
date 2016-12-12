
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE "jobs" ADD COLUMN starts_at bigint;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE "jobs" DROP COLUMN starts_at;
