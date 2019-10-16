-- +migrate Up

CREATE INDEX idx_om_last_status_sent ON outgoing_messages (last_status, sent_at);

-- +migrate Down

DROP INDEX idx_om_last_status_sent;
