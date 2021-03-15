-- +migrate Up

UPDATE engine_processing_versions
SET version = 2
WHERE type_id = 'rotation';

-- Start with a DEFAULT of 1 so that existing rows are
-- marked as the old version when the column is added.
ALTER TABLE rotation_state
ADD COLUMN version INT DEFAULT 1 NOT NULL;

-- Change the DEFAULT for all future inserted rows to be 2.
ALTER TABLE rotation_state
ALTER COLUMN version SET DEFAULT 2;

-- +migrate Down

UPDATE engine_processing_versions
SET version = 1
WHERE type_id = 'rotation';

ALTER TABLE rotation_state
DROP COLUMN processing_version;
