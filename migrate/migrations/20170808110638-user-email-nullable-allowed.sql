
-- +migrate Up
ALTER TABLE users
    DROP CONSTRAINT goalert_user_email_key,
    ALTER COLUMN EMAIL SET DEFAULT '';

-- +migrate Down
ALTER TABLE users 
    ADD CONSTRAINT goalert_user_email_key UNIQUE (email),
    ALTER COLUMN EMAIL DROP DEFAULT;