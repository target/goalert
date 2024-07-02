-- +migrate Up notransaction
ALTER TYPE enum_limit_type
    ADD VALUE IF NOT EXISTS 'pending_signals_per_service';

ALTER TYPE enum_limit_type
    ADD VALUE IF NOT EXISTS 'pending_signals_per_dest_per_service';

-- +migrate Down
