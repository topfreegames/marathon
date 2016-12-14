-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE "testapp_apns" (
  "id" uuid DEFAULT uuid_generate_v4(),
  "user_id" text NOT NULL,
  "token" text NOT NULL,
  "locale" text NOT NULL,
  "tz" text NOT NULL,
  PRIMARY KEY ("id")
);

CREATE TABLE "testapp_gcm" (
  "id" uuid DEFAULT uuid_generate_v4(),
  "user_id" text NOT NULL,
  "token" text NOT NULL,
  "locale" text NOT NULL,
  "tz" text NOT NULL,
  PRIMARY KEY ("id")
);

INSERT INTO testapp_apns (user_id, token, locale, tz) VALUES ('9e558649-9c23-469d-a11c-59b05813e3d5', '1234', 'pt', '-0300');
INSERT INTO testapp_apns (user_id, token, locale, tz) VALUES ('57be9009-e616-42c6-9cfe-505508ede2d0', '1235', 'en', '-0300');
INSERT INTO testapp_apns (user_id, token, locale, tz) VALUES ('a8e8d2d5-f178-4d90-9b31-683ad3aae920', '1236', 'pt', '-0300');
INSERT INTO testapp_apns (user_id, token, locale, tz) VALUES ('5c3033c0-24ad-487a-a80d-68432464c8de', '1237', 'en', '-0500');
INSERT INTO testapp_apns (user_id, token, locale, tz) VALUES ('4223171e-c665-4612-9edd-485f229240bf', '1238', 'pt', '-0300');
INSERT INTO testapp_apns (user_id, token, locale, tz) VALUES ('2df5bb01-15d1-4569-bc56-49fa0a33c4c3', '1239', 'en', '-0300');
INSERT INTO testapp_apns (user_id, token, locale, tz) VALUES ('67b872de-8ae4-4763-aef8-7c87a7f928a7', '1244', 'pt', '-0500');
INSERT INTO testapp_apns (user_id, token, locale, tz) VALUES ('3f8732a1-8642-4f22-8d77-a9688dd6a5ae', '1245', 'pt', '-0300');
INSERT INTO testapp_apns (user_id, token, locale, tz) VALUES ('21854bbf-ea7e-43e3-8f79-9ab2c121b941', '1246', 'en', '-0300');
INSERT INTO testapp_apns (user_id, token, locale, tz) VALUES ('843a61f8-45b3-44f9-9ab7-8becb2765653', '1247', 'pt', '-0500');
INSERT INTO testapp_apns (user_id, token, locale, tz) VALUES ('843a61f8-45b3-44f9-9ab7-8becb3365653', '1247', 'au', '-0500');

INSERT INTO testapp_gcm (user_id, token, locale, tz) VALUES ('9e558649-9c23-469d-a11c-59b05000e3d5', '1234', 'pt', '-0300');
INSERT INTO testapp_gcm (user_id, token, locale, tz) VALUES ('57be9009-e616-42c6-9cfe-505508ede2d0', '1235', 'en', '-0300');
INSERT INTO testapp_gcm (user_id, token, locale, tz) VALUES ('a8e8d2d5-f178-4d90-9b31-683ad3aae920', '1236', 'pt', '-0300');
INSERT INTO testapp_gcm (user_id, token, locale, tz) VALUES ('5c3033c0-24ad-487a-a80d-68432464c8de', '1237', 'en', '-0500');
INSERT INTO testapp_gcm (user_id, token, locale, tz) VALUES ('4223171e-c665-4612-9edd-485f229240bf', '1238', 'pt', '-0300');
INSERT INTO testapp_gcm (user_id, token, locale, tz) VALUES ('2df5bb01-15d1-4569-bc56-49fa0a33c4c3', '1239', 'en', '-0300');
INSERT INTO testapp_gcm (user_id, token, locale, tz) VALUES ('67b872de-8ae4-4763-aef8-7c87a7f928a7', '1244', 'pt', '-0500');
INSERT INTO testapp_gcm (user_id, token, locale, tz) VALUES ('3f8732a1-8642-4f22-8d77-a9688dd6a5ae', '1245', 'pt', '-0300');
INSERT INTO testapp_gcm (user_id, token, locale, tz) VALUES ('21854bbf-ea7e-43e3-8f79-9ab2c121b941', '1246', 'en', '-0300');
INSERT INTO testapp_gcm (user_id, token, locale, tz) VALUES ('843a61f8-45b3-44f9-9ab7-8becb2765653', '1247', 'pt', '-0500');

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE "testapp_apns";
DROP TABLE "testapp_gcm";
