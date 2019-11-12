-- +migrate Up

UPDATE engine_processing_versions
SET "version" = 7
WHERE type_id = 'message';

-- +migrate Down

UPDATE engine_processing_versions
SET "version" = 6
WHERE type_id = 'message';
