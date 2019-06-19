-- +migrate Up
ALTER TABLE user_favorites ADD COLUMN tgt_rotation_id UUID REFERENCES rotations (id);

ALTER TABLE user_favorites ALTER COLUMN tgt_service_id DROP NOT NULL;

ALTER TABLE user_favorites ADD CONSTRAINT user_fav_rotations UNIQUE (user_id, tgt_rotation_id);



-- +migrate Down
ALTER TABLE user_favorites DROP CONSTRAINT user_fav_rotations;

ALTER TABLE user_favorites DROP COLUMN tgt_rotation_id;

ALTER TABLE user_favorites ALTER COLUMN tgt_service_id SET NOT NULL;


