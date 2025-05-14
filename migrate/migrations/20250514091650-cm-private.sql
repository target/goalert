-- +migrate Up
ALTER TABLE user_contact_methods
    ADD COLUMN private boolean DEFAULT FALSE;

-- +migrate Down
ALTER TABLE user_contact_methods
    DROP COLUMN private;

