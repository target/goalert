-- +migrate Up
ALTER TABLE user_favorites ADD COLUMN tgt_rotation_id UUID;


-- +migrate Down
ALTER TABLE user_favorites DROP COLUMN tgt_rotation_id;


