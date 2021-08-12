-- +migrate Up
UPDATE engine_processing_versions
SET version = 9
WHERE type_id = 'message';

ALTER TABLE outgoing_messages
    ADD COLUMN src_value TEXT;

-- +migrate Down

UPDATE engine_processing_versions
SET version = 8
WHERE type_id = 'message';

ALTER TABLE outgoing_messages
    DROP COLUMN src_value;
