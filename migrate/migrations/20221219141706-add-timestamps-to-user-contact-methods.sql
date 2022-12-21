-- +migrate Up
ALTER TABLE user_contact_methods
    ADD COLUMN created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ADD COLUMN disabled_since TIMESTAMP WITH TIME ZONE DEFAULT NOW();

-- +migrate Down
ALTER TABLE user_contact_methods
    DROP COLUMN created_at,
    DROP COLUMN disabled_since;

