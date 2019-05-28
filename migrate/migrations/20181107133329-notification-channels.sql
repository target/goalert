
-- +migrate Up

CREATE TYPE enum_notif_channel_type AS ENUM (
    'SLACK'
);

CREATE TABLE notification_channels (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    type enum_notif_channel_type NOT NULL,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    meta jsonb
);

-- +migrate Down

DROP TABLE notification_channels;
DROP TYPE enum_notif_channel_type;
