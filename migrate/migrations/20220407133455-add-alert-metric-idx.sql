-- +migrate Up

CREATE INDEX alert_metrics_closed_date_idx ON alert_metrics (DATE(closed_at AT TIME ZONE 'UTC' ) ASC);

-- +migrate Down

DROP INDEX alert_metrics_closed_date_idx;