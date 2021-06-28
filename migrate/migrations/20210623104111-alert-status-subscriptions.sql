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

DROP TABLE user_last_alert_log;

LOCK outgoing_messages; 

DELETE FROM outgoing_messages
WHERE
    message_type = 'alert_status_update_bundle' AND (
        last_status = 'pending' OR next_retry_at notnull
    );

ALTER TABLE outgoing_messages 
    ADD CONSTRAINT om_no_status_bundles CHECK(
        message_type != 'alert_status_update_bundle' OR
        last_status != 'pending'
    );

-- +migrate Down
UPDATE engine_processing_versions SET version = 2 WHERE type_id = 'status_update';

ALTER TABLE outgoing_messages DROP CONSTRAINT om_no_status_bundles;

 CREATE TABLE user_last_alert_log (
    alert_id BIGINT NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    id BIGSERIAL NOT NULL,
    log_id BIGINT NOT NULL REFERENCES alert_logs(id) ON DELETE CASCADE,
    next_log_id BIGINT NOT NULL REFERENCES alert_logs(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT user_last_alert_log_uniq_id UNIQUE(id),
    PRIMARY KEY (user_id, alert_id)
);

CREATE INDEX idx_ulal_log_id on user_last_alert_log (log_id);
CREATE INDEX idx_ulal_next_log_id on user_last_alert_log (next_log_id);
CREATE INDEX idx_ulal_alert_id on user_last_alert_log (alert_id);

INSERT INTO user_last_alert_log(user_id, alert_id, log_id, next_log_id)
SELECT 
    cm.user_id,
    s.alert_id,

    (
        SELECT id from alert_logs
        WHERE alert_id = s.alert_id AND
        event = ANY(
            CASE 
                WHEN s.last_alert_status = 'triggered' THEN ARRAY['created', 'escalated']::enum_alert_log_event[]
                WHEN s.last_alert_status = 'active' THEN ARRAY['acknowledged']::enum_alert_log_event[]
                ELSE ARRAY['closed']::enum_alert_log_event[]
            END
        )
        ORDER BY id DESC LIMIT 1
    ),

    (
        SELECT id from alert_logs
        WHERE alert_id = s.alert_id AND
        event IN ('created', 'acknowledged', 'closed', 'escalated')
        ORDER BY id DESC LIMIT 1
    ) 
     
FROM alert_status_subscriptions s
JOIN user_contact_methods cm ON cm.id = s.contact_method_id;

DROP TABLE alert_status_subscriptions;

ALTER TABLE outgoing_messages 
    DROP CONSTRAINT om_no_status_bundles;
