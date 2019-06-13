-- +migrate Up
ALTER TABLE user_favorites ALTER COLUMN tgt_service_id DROP NOT NULL;

-- +migrate Down
ALTER TABLE user_favorites ALTER COLUMN tgt_service_id SET NOT NULL;


