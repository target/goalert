-- +migrate Up
ALTER TABLE auth_user_sessions
    ADD COLUMN last_access_at TIMESTAMPTZ;

-- +migrate Down

ALTER TABLE auth_user_sessions
    DROP COLUMN last_access_at;
