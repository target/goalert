-- +migrate Up
ALTER TABLE user_favorites 
ALTER COLUMN tgt_service_id 
DROP NOT NULL;

ALTER TABLE user_favorites
ADD UNIQUE (user_id, tgt_schedule_id);

-- +migrate Down
ALTER TABLE user_favorites 
ALTER COLUMN tgt_service_id 
SET NOT NULL;

ALTER TABLE user_favorites
DROP CONSTRAINT UNIQUE (user_id, tgt_schedule_id);