
-- +migrate Up
ALTER TABLE schedule_on_call_users
    DROP CONSTRAINT schedule_on_call_users_schedule_id_user_id_end_time_key;

CREATE UNIQUE INDEX idx_schedule_on_call_once
ON schedule_on_call_users (schedule_id, user_id)
WHERE end_time ISNULL;

-- +migrate Down
DROP INDEX idx_schedule_on_call_once;
ALTER TABLE schedule_on_call_users
    ADD CONSTRAINT schedule_on_call_users_schedule_id_user_id_end_time_key UNIQUE(schedule_id, user_id, end_time);
