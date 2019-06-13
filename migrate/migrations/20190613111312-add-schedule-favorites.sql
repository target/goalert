-- +migrate Up
ALTER TABLE user_favorites 
ADD COLUMN tgt_schedule_id UUID 
REFERENCES schedules(id);

-- +migrate Down
ALTER TABLE user_favorites 
DROP COLUMN tgt_schedule_id;