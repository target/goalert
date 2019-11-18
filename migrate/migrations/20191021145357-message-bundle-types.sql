-- +migrate Up notransaction

ALTER TYPE enum_outgoing_messages_status ADD VALUE IF NOT EXISTS 'bundled';
ALTER TYPE enum_outgoing_messages_type ADD VALUE IF NOT EXISTS 'alert_notification_bundle';
ALTER TYPE enum_outgoing_messages_type ADD VALUE IF NOT EXISTS 'alert_status_update_bundle';

-- +migrate Down
