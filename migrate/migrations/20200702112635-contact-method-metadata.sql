-- +migrate Up
ALTER TABLE user_contact_methods
    ADD COLUMN metadata JSONB;

-- +migrate Down

ALTER TABLE user_contact_methods
    DROP COLUMN metadata JSONB;
