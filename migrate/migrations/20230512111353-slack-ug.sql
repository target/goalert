-- +migrate Up notransaction
ALTER TYPE enum_notif_channel_type
    ADD VALUE IF NOT EXISTS 'SLACK_USER_GROUP';

-- +migrate Down
