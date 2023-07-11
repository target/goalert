-- +migrate Up
CREATE TABLE alert_feedback
(
    alert_id BIGINT PRIMARY KEY REFERENCES alerts (id) ON DELETE CASCADE,
    noise_reason TEXT NOT NULL
);

-- +migrate Down
DROP TABLE alert_feedback;
