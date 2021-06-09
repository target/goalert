-- +migrate Up
UPDATE engine_processing_versions SET version = 3 WHERE type_id = 'status_update';

CREATE TABLE alert_status_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    channel_id UUID REFERENCES notification_channels (id) ON DELETE CASCADE,
    contact_method_id UUID REFERENCES user_contact_methods (id) ON DELETE CASCADE,
    alert_id BIGINT NOT NULL REFERENCES alerts (id) ON DELETE CASCADE,
    last_alert_status enum_alert_status NOT NULL,

  UNIQUE (channel_id, contact_method_id, alert_id),
    CHECK ((channel_id IS NULL) != (contact_method_id IS NULL))
);

INSERT INTO alert_status_subscriptions (contact_method_id, alert_id, last_alert_status)
SELECT
    u.alert_status_log_contact_method_id,
    l.alert_id,
    CASE
        WHEN l.event = 'acknowledged'
        THEN 'active'::enum_alert_status
        ELSE 'triggered'::enum_alert_status
    END
FROM user_last_alert_log
JOIN users u ON u.id = user_id
JOIN alert_logs l ON l.id = log_id AND l.event != 'closed';

-- DROP TABLE IF EXISTS user_last_alert_log;

-- +migrate Down
UPDATE engine_processing_versions SET version = 2 WHERE type_id = 'status_update';

DROP TABLE alert_status_subscriptions;
