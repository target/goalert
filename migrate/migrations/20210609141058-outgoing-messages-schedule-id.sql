-- +migrate Up
UPDATE engine_processing_versions
SET version = 8
WHERE type_id = 'message';

ALTER TABLE
  outgoing_messages
ADD
  COLUMN schedule_id UUID REFERENCES schedules(id) ON DELETE CASCADE;

-- +migrate Down
UPDATE engine_processing_versions
SET version = 7
WHERE type_id = 'message';

DELETE FROM outgoing_messages
WHERE message_type = 'schedule_on_call_status';

ALTER TABLE
  outgoing_messages DROP COLUMN schedule_id;
