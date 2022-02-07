-- +migrate Up

CREATE TABLE alert_metrics (
    alert_id INT PRIMARY KEY REFERENCES alerts (id) ON DELETE CASCADE,
    service_id UUID NOT NULL REFERENCES services (id) ON DELETE CASCADE,
    time_to_ack INTERVAL,
    time_to_close INTERVAL NOT NULL,
    escalated BOOLEAN DEFAULT FALSE NOT NULL
);

CREATE INDEX idx_closed_events ON alert_logs (timestamp) WHERE event = 'closed';

-- +migrate Down

DROP INDEX idx_closed_events;

DROP TABLE alert_metrics;
