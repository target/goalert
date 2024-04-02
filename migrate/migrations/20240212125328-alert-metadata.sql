-- +migrate Up
CREATE TABLE alert_data(
    alert_id bigint PRIMARY KEY REFERENCES alerts(id) ON DELETE CASCADE,
    metadata jsonb,
    id bigserial UNIQUE NOT NULL
);

-- +migrate Down
DROP TABLE alert_data;

