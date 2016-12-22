-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE "jobs" ADD COLUMN past_time_strategy TEXT;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE "jobs" DROP COLUMN past_time_strategy;
