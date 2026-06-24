-- +migrate Up
ALTER TABLE user_contact_methods
    ADD COLUMN private boolean NOT NULL DEFAULT FALSE;

-- +migrate Down
ALTER TABLE user_contact_methods
    DROP COLUMN private;

