-- +migrate Up

ALTER TYPE enum_notif_channel_type ADD VALUE IF NOT EXISTS 'WEBHOOK';

-- +migrate Down
