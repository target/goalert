-- +migrate Up
ALTER TABLE notification_channels
    ADD COLUMN dest jsonb UNIQUE CHECK (type != 'DEST'
        OR dest IS NOT NULL);

-- +migrate Down
ALTER TABLE notification_channels
    DROP COLUMN dest;

