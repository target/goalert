-- +migrate Up

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

CREATE INDEX alert_metrics_closed_date_idx ON alert_metrics (DATE(closed_at AT TIME ZONE 'UTC' ) ASC);

-- +migrate Down

DROP INDEX alert_metrics_closed_date_idx;

DROP TABLE daily_alert_metrics;

UPDATE engine_processing_versions 
SET state = DEFAULT
WHERE type_id = 'metrics';
