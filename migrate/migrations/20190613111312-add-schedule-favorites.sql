-- +migrate Up
ALTER TABLE user_favorites
    ADD COLUMN tgt_schedule_id UUID REFERENCES schedules(id),
    ALTER COLUMN tgt_service_id DROP NOT NULL,
    ADD CONSTRAINT user_favorites_user_id_tgt_schedule_id_key UNIQUE(user_id, tgt_schedule_id);

-- +migrate Down
ALTER TABLE user_favorites
    DROP CONSTRAINT user_favorites_user_id_tgt_schedule_id_key,
    DROP COLUMN tgt_schedule_id,
    ALTER COLUMN tgt_service_id SET NOT NULL;