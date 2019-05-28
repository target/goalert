
-- +migrate Up

ALTER TABLE users ADD COLUMN avatar_url TEXT NOT NULL DEFAULT '';

-- +migrate Down

ALTER TABLE users DROP COLUMN avatar_url;
