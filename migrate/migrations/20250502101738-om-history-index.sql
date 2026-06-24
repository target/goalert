-- +migrate Up
CREATE INDEX ON message_status_history(message_id);

-- +migrate Down
DROP INDEX message_status_history_message_id_idx;

