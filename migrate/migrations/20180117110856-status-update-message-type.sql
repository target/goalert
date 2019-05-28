
-- +migrate Up notransaction

ALTER TYPE enum_outgoing_messages_type
    ADD VALUE IF NOT EXISTS 'alert_status_update';

-- +migrate Down
