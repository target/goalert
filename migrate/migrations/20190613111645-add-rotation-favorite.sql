-- +migrate Up
ALTER TABLE user_favorites ADD COLUMN tgt_rotation_id UUID REFERENCES rotations (id);


-- +migrate Down
ALTER TABLE user_favorites DROP COLUMN tgt_rotation_id;


