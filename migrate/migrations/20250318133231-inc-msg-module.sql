-- +migrate Up
UPDATE
    engine_processing_versions
SET
    "version" = 11
WHERE
    type_id = 'message';

-- +migrate Down
UPDATE
    engine_processing_versions
SET
    "version" = 10
WHERE
    type_id = 'message';

