-- +migrate Up
CREATE TABLE alert_data (
	alert_id BIGINT PRIMARY KEY REFERENCES alerts (id) ON DELETE CASCADE,
	metadata jsonb
);

-- +migrate Down
DROP TABLE alert_data;
