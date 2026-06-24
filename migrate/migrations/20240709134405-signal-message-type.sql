-- +migrate Up notransaction
ALTER TYPE enum_outgoing_messages_type
    ADD VALUE IF NOT EXISTS 'signal_message';

-- +migrate Down
