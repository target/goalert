-- +migrate Up

UPDATE engine_processing_versions
SET "version" = 3, state = DEFAULT
WHERE type_id = 'metrics';

-- +migrate Down

UPDATE engine_processing_versions
SET "version" = 2, state = DEFAULT
WHERE type_id = 'metrics';
