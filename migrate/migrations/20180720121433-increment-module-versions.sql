
-- +migrate Up

UPDATE engine_processing_versions
SET "version" = 2
WHERE type_id = 'escalation';

UPDATE engine_processing_versions
SET "version" = 2
WHERE type_id = 'np_cycle';

-- +migrate Down

UPDATE engine_processing_versions
SET "version" = 1
WHERE type_id = 'escalation';

UPDATE engine_processing_versions
SET "version" = 1
WHERE type_id = 'np_cycle';
