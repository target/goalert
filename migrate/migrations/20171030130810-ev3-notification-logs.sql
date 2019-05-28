
-- +migrate Up

CREATE TABLE notification_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_id BIGINT NOT NULL REFERENCES alerts (id) ON DELETE CASCADE,
    contact_method_id UUID NOT NULL REFERENCES user_contact_methods (id) ON DELETE CASCADE,
    process_timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    completed BOOLEAN NOT NULL DEFAULT FALSE
);

INSERT INTO notification_logs (
    id,
    alert_id,
    contact_method_id,
    process_timestamp,
    completed
)
SELECT
    s.id,
    s.alert_id,
    s.contact_method_id,
    s.sent_at,
    TRUE
FROM
    sent_notifications s
JOIN notification_policy_cycles c ON s.cycle_id = c.id -- only record sent notifications for active cycles, to preserve old behavior
WHERE
    s.sent_at IS NOT NULL
ORDER BY s.sent_at DESC
ON CONFLICT DO NOTHING;

-- +migrate Down

DROP TABLE notification_logs;
