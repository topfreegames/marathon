
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE "status" (
  "id" uuid DEFAULT uuid_generate_v4() UNIQUE,
  "name" text NOT NULL,
  "job_id" uuid NOT NULL,
  "created_at" bigint,
  PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX unique_job_status ON "status"(name, job_id);
ALTER TABLE "status"
ADD CONSTRAINT jobs_id_foreign
FOREIGN KEY (job_id)
REFERENCES jobs(id)
ON DELETE CASCADE
ON UPDATE CASCADE;

CREATE TYPE events_status AS ENUM ('fail', 'running', 'success');

CREATE TABLE "events" (
  "id" uuid DEFAULT uuid_generate_v4() UNIQUE,
  "state" events_status NOT NULL,
  "message" text NOT NULL,
  "status_id" uuid NOT NULL,
  "created_at" bigint,
  PRIMARY KEY ("id")
);

ALTER TABLE "events"
ADD CONSTRAINT status_id_foreign
FOREIGN KEY (status_id)
REFERENCES status(id)
ON DELETE CASCADE
ON UPDATE CASCADE;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE "status";
DROP TABLE "events";
