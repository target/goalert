-- +migrate Up notransaction
ALTER TYPE enum_outgoing_messages_status
    ADD VALUE IF NOT EXISTS 'read';

-- +migrate Down
