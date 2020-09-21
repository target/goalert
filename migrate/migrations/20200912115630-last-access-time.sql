-- +migrate Up
ALTER TABLE auth_user_sessions
    ALTER COLUMN last_access_at SET NOT NULL,
    ALTER COLUMN last_access_at SET DEFAULT now();

-- +migrate Down
ALTER TABLE auth_user_sessions
    ALTER COLUMN last_access_at DROP NOT NULL,
    ALTER COLUMN last_access_at DROP DEFAULT;