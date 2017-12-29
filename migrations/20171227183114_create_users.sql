
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE "users" (
  "id" uuid DEFAULT uuid_generate_v4(),
  "email" text NOT NULL,
  "is_admin" boolean NOT NULL,
  "allowed_apps" uuid[],
  "created_by" text NOT NULL,
  "created_at" bigint,
  "updated_at" bigint ,
  PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX uix_users_email ON "users"(email);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE "users";
