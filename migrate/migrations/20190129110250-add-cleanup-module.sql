-- +migrate Up notransaction
ALTER TYPE engine_processing_type ADD VALUE IF NOT EXISTS 'cleanup';
INSERT INTO engine_processing_versions (type_id) VALUES ('cleanup');

-- +migrate Down
DELETE FROM engine_processing_versions WHERE type_id = 'cleanup';
