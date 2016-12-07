
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE "apps" (
  "id" uuid DEFAULT uuid_generate_v4(),
  "name" text NOT NULL,
  "bundle_id" text NOT NULL,
  "created_by" text NOT NULL,
  "created_at" bigint,
  "updated_at" bigint ,
  PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX uix_apps_bundle_id ON "apps"(bundle_id);

CREATE TABLE "templates" (
  "id" uuid DEFAULT uuid_generate_v4(),
  "name" text NOT NULL,
  "locale" text NOT NULL,
  "defaults" JSONB NOT NULL DEFAULT '{}'::JSONB,
  "body" JSONB NOT NULL DEFAULT '{}'::JSONB,
  "created_by" text NOT NULL,
  "app_id" uuid NOT NULL,
  "created_at" bigint,
  "updated_at" bigint ,
  PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX name_locale_app ON "templates"("name", "locale", app_id);

ALTER TABLE "templates"
ADD CONSTRAINT templates_app_id_apps_id_foreign
FOREIGN KEY (app_id)
REFERENCES apps(id)
ON DELETE CASCADE
ON UPDATE CASCADE;

CREATE TABLE "jobs" (
  "id" uuid DEFAULT uuid_generate_v4(),
  "total_batches" integer,
  "completed_batches" integer NOT NULL DEFAULT 0,
  "completed_at" bigint,
  "expires_at" bigint,
  "context" JSONB NOT NULL DEFAULT '{}'::JSONB,
  "service" text,
  "filters" JSONB NOT NULL DEFAULT '{}'::JSONB,
  "metadata" JSONB NOT NULL DEFAULT '{}'::JSONB,
  "csv_url" text,
  "created_by" text,
  "app_id" uuid NOT NULL,
  "template_name" text NOT NULL,
  "created_at" bigint,
  "updated_at" bigint,
  PRIMARY KEY ("id")
);

ALTER TABLE "jobs"
ADD CONSTRAINT jobs_app_id_apps_id_foreign
FOREIGN KEY (app_id)
REFERENCES apps(id)
ON DELETE CASCADE
ON UPDATE CASCADE;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE "jobs";
DROP TABLE "templates";
DROP TABLE "apps";
