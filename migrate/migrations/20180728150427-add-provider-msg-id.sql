
-- +migrate Up

UPDATE engine_processing_versions
SET "version" = 2
WHERE type_id = 'message';

ALTER TABLE outgoing_messages
    ADD COLUMN provider_msg_id TEXT,
    ADD COLUMN provider_seq INT NOT NULL DEFAULT 0;

CREATE UNIQUE INDEX idx_outgoing_messages_provider_msg_id ON outgoing_messages (provider_msg_id);

-- +migrate Down

-- Lower version first when migrating down, to stop processing
UPDATE engine_processing_versions
SET "version" = 1
WHERE type_id = 'message';

ALTER TABLE outgoing_messages
    DROP COLUMN provider_msg_id,
    DROP COLUMN provider_seq;
