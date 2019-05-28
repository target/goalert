
-- +migrate Up

CREATE TABLE alert_assignments (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    alert_id BIGINT NOT NULL REFERENCES alerts (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, alert_id)
);

-- +migrate Down

DROP TABLE alert_assignments;
