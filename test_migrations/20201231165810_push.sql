-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE "testapp_apns" (
  "id" uuid DEFAULT uuid_generate_v4(),
  "created_at" timestamp DEFAULT now(),
  "user_id" text NOT NULL,
  "token" text NOT NULL,
  "region" text NOT NULL,
  "locale" text NOT NULL,
  "tz" text NOT NULL,
  "adid" text NOT NULL,
  "fiu" text NOT NULL,
  "vendor_id" text NOT NULL,
  PRIMARY KEY ("id")
);

CREATE TABLE "testapp_gcm" (
  "id" uuid DEFAULT uuid_generate_v4(),
  "created_at" timestamp DEFAULT now(),
  "user_id" text NOT NULL,
  "token" text NOT NULL,
  "region" text NOT NULL,
  "locale" text NOT NULL,
  "tz" text NOT NULL,
  "adid" text NOT NULL,
  "fiu" text NOT NULL,
  "vendor_id" text NOT NULL,
  PRIMARY KEY ("id")
);

INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('9e558649-9c23-469d-a11c-59b05813e3d5', '1234', 'BR', 'pt', '-0300','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('57be9009-e616-42c6-9cfe-505508ede2d0', '1235', 'US', 'en', '-0300','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('a8e8d2d5-f178-4d90-9b31-683ad3aae920', '1236', 'BR', 'pt', '-0300','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('5c3033c0-24ad-487a-a80d-68432464c8de', '1237', 'US', 'en', '-0500','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('4223171e-c665-4612-9edd-485f229240bf', '1238', 'BR', 'pt', '-0300','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('2df5bb01-15d1-4569-bc56-49fa0a33c4c3', '1239', 'US', 'en', '-0300','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('67b872de-8ae4-4763-aef8-7c87a7f928a7', '1244', 'BR', 'pt', '-0500','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('3f8732a1-8642-4f22-8d77-a9688dd6a5ae', '1245', 'BR', 'pt', '-0300','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('21854bbf-ea7e-43e3-8f79-9ab2c121b941', '1246', 'US', 'en', '-0300','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('843a61f8-45b3-44f9-9ab7-8becb2765653', '1247', 'BR', 'pt', '-0500','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('843a61f8-45b3-44f9-9ab7-8becb3365653', '1247', 'AU', 'au', '-0500','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('843a61f8-45b3-44f9-aaaa-8becb3365653', '1247', 'FR', 'fr', '-0800','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('843a61f8-45b3-44f9-bbbb-8becb3365653', '1247', 'FR', 'fr', '-0800','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('e78431ca-69a8-4326-af1f-48f817a4a669', '1247', 'ES', 'es', '-0800','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('d9b42bb8-78ca-44d0-ae50-a472d9fbad92', '1247', 'ES', 'es', '-0800','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('ee4455fe-8ff6-4878-8d7c-aec096bd68b4', '1247', 'ES', 'es', '-0800','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('e78431ca-69a8-4326-af1f-48f817a4a669', '1247', 'ES', 'es', '-0800','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('e78431ca-69a8-4326-af1f-48f817a4a669', '1248', 'ES', 'es', '-0800','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('7ed725ce-e516-4386-bc6a-0b16bbbac678', '1222', 'CN', 'cn', '-4440','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('6ec06ad1-0416-4e0a-9c2c-0b4381976091', '1223', 'CN', 'cn', '-4440','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('a04087d6-4d95-4d99-901f-a1ff8578a2bf', '1224', 'CN', 'cn', '-4440','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('5146be6c-ffda-401c-8721-3c43c7370872', '1225', 'CN', 'cn', '-4440','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('dc2be5c1-2b6d-47d6-9a45-c188fd96d124', '1226', 'CN', 'cn', '-4440','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('7ae62ce6-94fb-4636-9484-05bae4398505', '1227', 'CN', 'cn', '-4440','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('9e3dfdf8-5991-4609-82ba-258ed2a78504', '1228', 'CN', 'cn', '-4440','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('f57a0010-1318-4997-9a92-dcfb8ca0f24a', '1229', 'CN', 'cn', '-4440','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('6be7b349-6034-4f99-847c-dab3ee4576d0', '1244', 'CN', 'cn', '-4440','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_apns (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('830a4cbf-c95f-40de-ab20-fef493899944', '1233', 'CN', 'cn', '-4440','some adid', 'some_fiu', 'some_vendor_id');

INSERT INTO testapp_gcm (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('9e558649-9c23-469d-a11c-59b05000e3d5', '1234', 'br', 'PT', '-0300','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_gcm (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('57be9009-e616-42c6-9cfe-505508ede2d0', '1235', 'us', 'EN', '-0300','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_gcm (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('a8e8d2d5-f178-4d90-9b31-683ad3aae920', '1236', 'br', 'PT', '-0300','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_gcm (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('5c3033c0-24ad-487a-a80d-68432464c8de', '1237', 'us', 'EN', '-0500','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_gcm (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('4223171e-c665-4612-9edd-485f229240bf', '1238', 'br', 'PT', '-0300','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_gcm (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('2df5bb01-15d1-4569-bc56-49fa0a33c4c3', '1239', 'us', 'EN', '-0300','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_gcm (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('67b872de-8ae4-4763-aef8-7c87a7f928a7', '1244', 'br', 'PT', '-0500','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_gcm (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('3f8732a1-8642-4f22-8d77-a9688dd6a5ae', '1245', 'br', 'PT', '-0300','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_gcm (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('21854bbf-ea7e-43e3-8f79-9ab2c121b941', '1246', 'us', 'EN', '-0300','some adid', 'some_fiu', 'some_vendor_id');
INSERT INTO testapp_gcm (user_id, token, region, locale, tz, adid, fiu, vendor_id) VALUES ('843a61f8-45b3-44f9-9ab7-8becb2765653', '1247', 'br', 'PT', '-0500','some adid', 'some_fiu', 'some_vendor_id');

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE "testapp_apns";
DROP TABLE "testapp_gcm";
