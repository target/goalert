-- +migrate Up
CREATE TABLE alert_metadata
(
    alert_id BIGINT PRIMARY KEY REFERENCES alerts (id) ON DELETE CASCADE,
    sentiment INT NOT NULL DEFAULT 0,
    note TEXT
);

-- +migrate Down
DROP TABLE alert_metadata;
