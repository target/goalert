-- +migrate Up notransaction
ALTER TYPE engine_processing_type ADD VALUE IF NOT EXISTS 'metrics';

-- +migrate Down
