
-- +migrate Up

UPDATE engine_processing_versions
SET "version" = 5
WHERE type_id = 'message';


UPDATE engine_processing_versions
SET "version" = 3
WHERE type_id = 'escalation';

-- +migrate Down

UPDATE engine_processing_versions
SET "version" = 2
WHERE type_id = 'escalation';

UPDATE engine_processing_versions
SET "version" = 4
WHERE type_id = 'message';
