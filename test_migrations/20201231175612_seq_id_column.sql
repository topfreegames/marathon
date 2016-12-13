
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

ALTER TABLE "testapp_apns" ADD COLUMN seq_id bigserial;
ALTER TABLE "testapp_gcm" ADD COLUMN seq_id bigserial;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

ALTER TABLE "testapp_apns" DROP COLUMN seq_id;
ALTER TABLE "testapp_gcm" DROP COLUMN seq_id;
