-- +migrate Up

UPDATE engine_processing_versions
SET version = 2
WHERE type_id = 'rotation';

ALTER TABLE rotation_state
ADD COLUMN version INT DEFAULT 1 NOT NULL;

ALTER TABLE rotation_state
ALTER COLUMN version SET DEFAULT 2;

-- +migrate Down

UPDATE engine_processing_versions
SET version = 1
WHERE type_id = 'rotation';

ALTER TABLE rotation_state
DROP COLUMN processing_version;
