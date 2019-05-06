
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE "job_groups" (
  "id" uuid DEFAULT uuid_generate_v4() UNIQUE,
  "app_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);

ALTER TABLE "job_groups"
ADD CONSTRAINT jobs_group_app_id_apps_id_foreign
FOREIGN KEY (app_id)
REFERENCES apps(id)
ON DELETE CASCADE
ON UPDATE CASCADE;

ALTER TABLE "jobs" ADD COLUMN job_group_id uuid DEFAULT uuid_generate_v4();

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

ALTER TABLE "jobs" DROP COLUMN job_group_id;
DROP TABLE "job_group";
