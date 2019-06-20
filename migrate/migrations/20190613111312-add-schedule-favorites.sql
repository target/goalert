-- +migrate Up
ALTER TABLE user_favorites 
ADD COLUMN tgt_schedule_id UUID 
REFERENCES schedules(id);

ALTER TABLE user_favorites 
ALTER COLUMN tgt_service_id 
DROP NOT NULL;

ALTER TABLE user_favorites 
ADD CONSTRAINT user_fav_schedules UNIQUE (user_id, tgt_schedule_id);

-- +migrate Down
ALTER TABLE user_favorites
DROP CONSTRAINT user_fav_schedules;

ALTER TABLE user_favorites 
DROP COLUMN tgt_schedule_id;

ALTER TABLE user_favorites 
ALTER COLUMN tgt_service_id 
SET NOT NULL;