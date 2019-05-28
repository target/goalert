
-- +migrate Up

ALTER TABLE users DROP COLUMN email_verified;

-- +migrate Down

ALTER TABLE users ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT false;
