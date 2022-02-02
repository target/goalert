-- +migrate Up

CREATE TABLE alert_metrics (
    alert_id INT PRIMARY KEY REFERENCES alerts (id) ON DELETE CASCADE,
    service_id UUID NOT NULL REFERENCES services (id) ON DELETE CASCADE,
    closed_at TIMESTAMPTZ NOT NULL,
    time_to_ack INTERVAL NOT NULL,
    time_to_close INTERVAL NOT NULL,
    escalated BOOLEAN DEFAULT FALSE NOT NULL
);

-- +migrate Down

DROP TABLE alert_metrics;