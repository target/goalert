-- +migrate Up
ALTER TABLE user_favorites_user_id_tgt_rotation_id_key
    ADD COLUMN tgt_rotation_id UUID REFERENCES rotations (id),
    ALTER COLUMN tgt_service_id DROP NOT NULL,
    ADD CONSTRAINT user_favorite_rotations UNIQUE (user_id, tgt_rotation_id);

-- +migrate Down
ALTER TABLE user_favorites_user_id_tgt_rotation_id_key
    DROP CONSTRAINT user_favorite_rotations,
    DROP COLUMN tgt_rotation_id,
    ALTER COLUMN tgt_service_id SET NOT NULL;


