
-- +migrate Up

-- future work, for now we are using the user_id as the policy id

-- CREATE TABLE notification_policies (
--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     repeat_count INT NOT NULL DEFAULT 0,
--     repeat_delay_minutes INT NOT NULL DEFAULT 1
-- );

-- CREATE TABLE notification_policy_rules (
--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     notification_policy_id UUID NOT NULL REFERENCES notification_policies (id) ON DELETE CASCADE,
--     delay_minutes INT NOT NULL DEFAULT 0,
--     created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
-- );

CREATE TABLE notification_policy_cycles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    alert_id INT NOT NULL REFERENCES alerts (id) ON DELETE CASCADE,
    repeat_count INT NOT NULL DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    checked BOOLEAN NOT NULL DEFAULT TRUE,

    UNIQUE(user_id, alert_id)
);

INSERT INTO notification_policy_cycles (
    id,
    user_id,
    alert_id,
    started_at
)
SELECT
    c.id,
    c.user_id,
    c.alert_id,
    c.started_at
FROM
    user_notification_cycles c,
    alerts a
WHERE
    a.id = c.alert_id AND
    a.escalation_level = c.escalation_level;


-- +migrate Down
DROP TABLE notification_policy_cycles;
