-- +migrate Up
ALTER TABLE
  outgoing_messages
ADD
  COLUMN schedule_id UUID REFERENCES schedules(id) ON DELETE CASCADE;

-- +migrate Down
ALTER TABLE
  outgoing_messages DROP COLUMN schedule_id;