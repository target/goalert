-- +migrate Up
UPDATE engine_processing_versions
SET version = 3
WHERE type_id = 'schedule';

-- +migrate Down
UPDATE engine_processing_versions
SET version = 2
WHERE type_id = 'schedule';
