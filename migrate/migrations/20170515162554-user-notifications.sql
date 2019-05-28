
-- +migrate Up

CREATE TABLE notifications (
    user_id UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    started_at TIMESTAMP NOT NULL DEFAULT now()
);


CREATE TABLE sent_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_id BIGINT NOT NULL REFERENCES alerts (id) ON DELETE CASCADE,
    contact_method_id UUID NOT NULL REFERENCES user_contact_methods (id) ON DELETE CASCADE,
    sent_at TIMESTAMP,
    UNIQUE(alert_id,contact_method_id)
);

CREATE VIEW active_contact_methods AS
    SELECT users.id as user_id, m.id as contact_method_id
        FROM users, user_contact_methods m, user_notification_rules r, notifications n
        WHERE m.user_id = users.id
            AND r.user_id = users.id
            AND n.user_id = users.id
            AND r.contact_method_id = m.id
            AND ((r.delay_minutes::text||' minutes')::interval + n.started_at) < now();

CREATE VIEW triggered_alert_users AS
    SELECT action.user_id as user_id, a.id as alert_id
        FROM escalation_policy_actions action, escalation_policy_step step, service s, alerts a, alert_escalation_levels lvl
        WHERE action.escalation_policy_step_id = step.id
            AND step.escalation_policy_id = s.escalation_policy_id
            AND step.step_number = lvl.relative_level
            AND s.id = a.service_id
            AND a.status = 'triggered'::enum_alert_status
        GROUP BY a.id, action.user_id;


-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION update_notifications() RETURNS VOID AS
    $$
        BEGIN
        INSERT INTO notifications (user_id)
            SELECT user_id FROM triggered_alert_users
            GROUP BY user_id
            ON CONFLICT DO NOTHING;

        DELETE FROM notifications WHERE user_id NOT IN (SELECT user_id FROM triggered_alert_users WHERE user_id = user_id);
        END;
    $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION add_notifications() RETURNS TRIGGER AS
    $$
        BEGIN
            INSERT INTO notifications (user_id)
                SELECT user_id FROM triggered_alert_users
                WHERE alert_id = NEW.id
                LIMIT 1
                ON CONFLICT DO NOTHING;

            DELETE FROM notifications WHERE user_id NOT IN (SELECT user_id FROM triggered_alert_users WHERE user_id = user_id);
            RETURN NEW;
        END;
    $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

CREATE TRIGGER add_notifications_alert_changed
    AFTER UPDATE OR INSERT ON alerts
    FOR EACH ROW
    EXECUTE PROCEDURE add_notifications();

SELECT update_notifications();


CREATE VIEW needs_notification_sent AS
    SELECT trig.alert_id, acm.contact_method_id, cm.type, cm.value, a.description, s.name as service_name
    FROM active_contact_methods acm, triggered_alert_users trig, user_contact_methods cm, alerts a, service s
    WHERE acm.user_id = trig.user_id
        AND acm.user_id = trig.user_id
        AND cm.id = acm.contact_method_id
        AND cm.disabled = FALSE
        AND a.id = trig.alert_id
        AND s.id = a.service_id
        AND NOT EXISTS (
            SELECT id
            FROM sent_notifications
            WHERE alert_id = trig.alert_id
                AND contact_method_id = acm.contact_method_id
                AND sent_at IS NOT NULL
        );


CREATE TABLE user_contact_method_locks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL,
    alert_id BIGINT NOT NULL REFERENCES alerts (id) ON DELETE CASCADE,
    contact_method_id UUID NOT NULL REFERENCES user_contact_methods (id) ON DELETE CASCADE,
    timestamp TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE (alert_id, contact_method_id)
);

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION aquire_user_contact_method_lock(_client_id UUID, _alert_id BIGINT, _contact_method_id UUID) RETURNS UUID AS
    $$
        DECLARE
            lock_id UUID = gen_random_uuid();
        BEGIN
            DELETE FROM user_contact_method_locks WHERE alert_id = _alert_id
                AND contact_method_id = _contact_method_id
                AND (timestamp + '5 minutes'::interval) < now();

            INSERT INTO user_contact_method_locks (id, alert_id, contact_method_id, client_id) 
                VALUES (lock_id, _alert_id, _contact_method_id, _client_id)
                RETURNING id INTO lock_id;

            INSERT INTO sent_notifications (id, alert_id, contact_method_id) VALUES (lock_id, _alert_id, _contact_method_id);

            RETURN lock_id;
        END;
    $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION release_user_contact_method_lock(_client_id UUID, _id UUID, success BOOLEAN) RETURNS VOID AS
    $$
        BEGIN
            DELETE FROM user_contact_method_locks WHERE id = _id AND client_id = _client_id;
            IF success
            THEN
                UPDATE sent_notifications SET sent_at = now() WHERE id = _id;
            ELSE
                DELETE FROM sent_notifications WHERE id = _id;
            END IF;
        END;
    $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd


-- +migrate Down

DROP FUNCTION update_notifications();
DROP VIEW active_contact_methods;
DROP VIEW triggered_alert_users;
DROP TABLE sent_notifications;
DROP TABLE notifications;
