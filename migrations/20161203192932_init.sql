
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE "apps" (
  "id" uuid DEFAULT uuid_generate_v4(),
  "name" text NOT NULL,
  "bundle_id" text NOT NULL,
  "created_by" text NOT NULL,
  "created_at" timestamp with time zone,
  "updated_at" timestamp with time zone , 
  PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX uix_apps_bundle_id ON "apps"(bundle_id);

CREATE TABLE "templates" (
  "id" uuid DEFAULT uuid_generate_v4(),
  "name" text NOT NULL,
  "locale" text NOT NULL,
  "defaults" JSONB NOT NULL DEFAULT '{}'::JSONB,
  "body" JSONB NOT NULL DEFAULT '{}'::JSONB,
  "compiled_body" text NOT NULL,
  "created_by" text NOT NULL,
  "app_id" uuid NOT NULL,
  "created_at" timestamp with time zone,
  "updated_at" timestamp with time zone , 
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
  "completed_at" timestamp with time zone,
  "expires_at" timestamp with time zone,
  "context" JSONB NOT NULL DEFAULT '{}'::JSONB,
  "service" text,
  "filters" JSONB NOT NULL DEFAULT '{}'::JSONB,
  "csv_url" text,
  "created_by" text,
  "app_id" uuid NOT NULL,
  "template_id" uuid NOT NULL,
  "created_at" timestamp with time zone,
  "updated_at" timestamp with time zone, 
  PRIMARY KEY ("id")
);

ALTER TABLE "jobs" 
ADD CONSTRAINT jobs_app_id_apps_id_foreign 
FOREIGN KEY (app_id) 
REFERENCES apps(id) 
ON DELETE CASCADE 
ON UPDATE CASCADE;

ALTER TABLE "jobs" 
ADD CONSTRAINT jobs_template_id_templates_id_foreign 
FOREIGN KEY (template_id) 
REFERENCES templates(id) 
ON DELETE CASCADE 
ON UPDATE CASCADE;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE "jobs";
DROP TABLE "templates";
DROP TABLE "apps";
