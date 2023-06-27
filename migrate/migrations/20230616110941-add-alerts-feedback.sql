-- +migrate Up
CREATE TABLE alert_feedback
(
    alert_id BIGINT PRIMARY KEY REFERENCES alerts (id) ON DELETE CASCADE,
    note TEXT
);

-- +migrate Down
DROP TABLE alert_feedback;
