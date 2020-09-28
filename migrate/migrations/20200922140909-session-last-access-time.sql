-- +migrate Up

UPDATE auth_user_sessions SET last_access_at = now() WHERE last_access_at ISNULL;

ALTER TABLE auth_user_sessions
    ALTER COLUMN last_access_at SET DEFAULT now(),
    ALTER COLUMN last_access_at SET NOT NULL;

-- +migrate Down
ALTER TABLE auth_user_sessions
    ALTER COLUMN last_access_at DROP NOT NULL,
    ALTER COLUMN last_access_at DROP DEFAULT;
