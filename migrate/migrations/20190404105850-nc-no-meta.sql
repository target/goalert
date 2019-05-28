
-- +migrate Up

UPDATE notification_channels SET meta = '{}' WHERE meta isnull;
ALTER TABLE notification_channels
    ALTER COLUMN meta SET DEFAULT '{}',
    ALTER COLUMN meta SET NOT NULL;

-- +migrate Down
ALTER TABLE notification_channels
    ALTER COLUMN meta DROP NOT NULL,
    ALTER COLUMN meta DROP DEFAULT;
