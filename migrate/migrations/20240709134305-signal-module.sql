-- +migrate Up notransaction
ALTER TYPE engine_processing_type
    ADD VALUE IF NOT EXISTS 'signals';

INSERT INTO engine_processing_versions(type_id, version)
    VALUES ('signals', 1)
ON CONFLICT
    DO NOTHING;

-- +migrate Down
DELETE FROM engine_processing_versions
WHERE type_id = 'signals';

