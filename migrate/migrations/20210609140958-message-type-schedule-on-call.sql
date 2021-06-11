-- +migrate Up notransaction

ALTER TYPE enum_outgoing_messages_type ADD VALUE IF NOT EXISTS 'schedule_on_call_status';

-- +migrate Down
