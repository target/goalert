
-- +migrate Up

DROP TABLE daily_alert_metrics;

-- +migrate Down

CREATE TABLE daily_alert_metrics (
    id BIGSERIAL PRIMARY KEY,
    service_id UUID NOT NULL,
    date DATE NOT NULL,
    alert_count INT DEFAULT 0 NOT NULL,
    avg_time_to_ack INTERVAL,
    avg_time_to_close INTERVAL,
    escalated_count INT DEFAULT 0 NOT NULL,

    UNIQUE(service_id, date)
);