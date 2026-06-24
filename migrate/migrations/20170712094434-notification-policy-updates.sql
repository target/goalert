
-- +migrate Up

-- Update to return alert escalation_level
CREATE OR REPLACE VIEW on_call_alert_users AS
  WITH alert_users AS (
			SELECT act.user_id, act.schedule_id, a.id as alert_id, a.status, a.escalation_level
			FROM
				alerts a,
				service s,
				alert_escalation_levels lvl,
				escalation_policy_step step,
				escalation_policy_actions act
			WHERE s.id = a.service_id
				AND lvl.alert_id = a.id
				AND step.escalation_policy_id = s.escalation_policy_id
				AND step.step_number = lvl.relative_level
				AND a.status != 'closed'::enum_alert_status
				AND act.escalation_policy_step_id = step.id
			GROUP BY user_id, schedule_id, a.id
		)
		
		SELECT DISTINCT
			au.alert_id,
			au.status,
			CASE WHEN au.user_id IS NULL THEN oc.user_id
			ELSE au.user_id
			END,
            au.escalation_level
		FROM alert_users au
		LEFT JOIN on_call oc ON au.schedule_id = oc.schedule_id;

-- new notification tracking table
CREATE TABLE user_notification_cycles (
	id UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    alert_id BIGINT NOT NULL REFERENCES alerts (id) ON DELETE CASCADE,
    escalation_level INT NOT NULL,
    started_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, alert_id)
);

-- Add new throttle type, so new notification cycle can run during deployment
ALTER TYPE enum_throttle_type RENAME TO enum_throttle_type_old;
CREATE TYPE enum_throttle_type AS ENUM (
	'notifications',
	'notifications_2'
);
ALTER TABLE throttle ALTER COLUMN action TYPE enum_throttle_type USING action::TEXT::enum_throttle_type;
DROP TYPE enum_throttle_type_old;


DROP TRIGGER add_notifications_alert_changed ON alerts;
DROP FUNCTION add_notifications();
DROP FUNCTION update_notifications();


ALTER TABLE twilio_sms_callbacks DROP CONSTRAINT twilio_sms_callbacks_callback_id_fkey;
ALTER TABLE twilio_voice_callbacks DROP CONSTRAINT twilio_voice_callbacks_callback_id_fkey;
ALTER TABLE sent_notifications DROP CONSTRAINT sent_notifications_alert_id_contact_method_id_key;
ALTER TABLE sent_notifications DROP CONSTRAINT sent_notifications_pkey;
CREATE INDEX ON sent_notifications (id);

ALTER TABLE sent_notifications ADD COLUMN cycle_id UUID;
ALTER TABLE sent_notifications ADD COLUMN notification_rule_id UUID REFERENCES user_notification_rules (id) ON DELETE CASCADE;

DELETE FROM sent_notifications s
WHERE NOT EXISTS 
	(
		SELECT 1 FROM user_notification_rules r
		WHERE s.contact_method_id = r.contact_method_id
	);


WITH sent_users AS (
		SELECT DISTINCT alert_id, user_id
		FROM
			sent_notifications s,
			user_contact_methods c
		WHERE c.id = s.contact_method_id
	),
	cycles AS (
		SELECT alert_id, user_id, gen_random_uuid() as cycle_id
		FROM sent_users
	)
UPDATE sent_notifications n
SET cycle_id = c.cycle_id
FROM
	cycles c,
	user_contact_methods m
WHERE n.cycle_id IS NULL
	AND m.id = n.contact_method_id
	AND m.user_id = c.user_id
	AND n.alert_id = c.alert_id
;

-- +migrate StatementBegin
DO
$do$
BEGIN
	IF EXISTS (SELECT 1 FROM sent_notifications WHERE sent_at IS NULL) THEN
	RAISE EXCEPTION 'found in-flight notifications (sent_at was NULL)';
	END IF;
	IF EXISTS (SELECT 1 FROM user_contact_method_locks) THEN
	RAISE EXCEPTION 'found active contact method locks';
	END IF;
END
$do$;
-- +migrate StatementEnd
WITH sent_times AS
	(
		SELECT s.alert_id, c.user_id, min(s.sent_at) AS sent_at
		FROM
			sent_notifications s,
			user_contact_methods c
		WHERE c.id = s.contact_method_id
		GROUP BY s.alert_id, c.user_id
	),
	start_times AS
	(
		SELECT
			s.alert_id,
			s.user_id,
			s.sent_at - (max(n.delay_minutes)::TEXT||' minutes')::INTERVAL AS sent_at
		FROM
			sent_times s,
			user_notification_rules n
		WHERE n.user_id = s.user_id
		GROUP BY s.alert_id, s.user_id, s.sent_at
	)
INSERT INTO user_notification_cycles (id, user_id, alert_id, escalation_level, started_at)
SELECT DISTINCT
	s.cycle_id,
	c.user_id,
	s.alert_id,
	a.escalation_level,
	t.sent_at
FROM
	sent_notifications s,
	alerts a,
	start_times t,
	user_contact_methods c
WHERE a.id = s.alert_id
	AND c.id = s.contact_method_id
	AND t.alert_id = s.alert_id
	AND t.user_id = c.user_id
ORDER BY sent_at DESC
ON CONFLICT (user_id, alert_id) DO NOTHING;

INSERT INTO sent_notifications
	(id, alert_id, contact_method_id, sent_at, cycle_id, notification_rule_id)
SELECT s.id, s.alert_id, s.contact_method_id, s.sent_at, s.cycle_id, n.id
FROM
	sent_notifications s,
	user_notification_rules n
WHERE n.contact_method_id = s.contact_method_id;

DELETE FROM sent_notifications WHERE notification_rule_id IS NULL;

ALTER TABLE sent_notifications ALTER COLUMN cycle_id SET NOT NULL;
ALTER TABLE sent_notifications ALTER COLUMN notification_rule_id SET NOT NULL;
ALTER TABLE sent_notifications ADD UNIQUE(notification_rule_id, cycle_id);

ALTER TABLE user_notification_rules ADD COLUMN created_at TIMESTAMP DEFAULT now();

CREATE OR REPLACE VIEW user_notification_cycle_state AS
	SELECT DISTINCT
		c.alert_id,
		nr.id AS notification_rule_id,
		nr.user_id,
		c.id AS cycle_id,
		(nr.delay_minutes::TEXT||' minutes')::INTERVAL > now()-c.started_at AS future,
		nr.id NOT IN (
			SELECT notification_rule_id
			FROM sent_notifications
			WHERE alert_id = c.alert_id
				AND cycle_id = c.id
				AND contact_method_id = nr.contact_method_id) AS pending,
		c.escalation_level
	FROM
		user_notification_cycles c,
		alerts a,
		user_notification_rules nr
	WHERE a.id = c.alert_id
		AND a.status = 'triggered'
		AND nr.user_id = c.user_id
		AND nr.id NOT IN (
			SELECT notification_rule_id
			FROM sent_notifications
			WHERE alert_id = c.alert_id
				AND cycle_id = c.id
				AND contact_method_id = nr.contact_method_id);

CREATE OR REPLACE VIEW needs_notification_sent AS
	SELECT DISTINCT cs.alert_id, nr.contact_method_id, cm.type, cm.value, a.description, s.name as service_name, nr.id as notification_rule_id, cs.escalation_level, cs.cycle_id FROM
		user_notification_cycle_state cs,
		alerts a,
		user_contact_methods cm,
		user_notification_rules nr,
		service s
	WHERE a.id = cs.alert_id
		AND a.status = 'triggered'
		AND cs.escalation_level = a.escalation_level
		AND cm.id = nr.contact_method_id
		AND nr.id = cs.notification_rule_id
		AND s.id = a.service_id
		AND cs.pending
		AND NOT cs.future;

DROP VIEW active_contact_methods;
DROP TABLE notifications;

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

            INSERT INTO sent_notifications (id, alert_id, contact_method_id, cycle_id, notification_rule_id)
			SELECT lock_id, _alert_id, _contact_method_id, cycle_id, notification_rule_id
			FROM needs_notification_sent n
			WHERE n.alert_id = _alert_id AND n.contact_method_id = _contact_method_id
			ON CONFLICT DO NOTHING;

            RETURN lock_id;
        END;
    $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION update_notification_cycles() RETURNS VOID AS
    $$
        BEGIN
			INSERT INTO user_notification_cycles (user_id, alert_id, escalation_level)
			SELECT user_id, alert_id, escalation_level
			FROM on_call_alert_users
			WHERE status = 'triggered'
			ON CONFLICT DO NOTHING;

			UPDATE user_notification_cycles c
			SET escalation_level = a.escalation_level
			FROM
				alerts a,
				user_notification_cycle_state s
			WHERE a.id = c.alert_id
				AND s.user_id = c.user_id
				AND s.alert_id = c.alert_id;

			DELETE FROM user_notification_cycles c
			WHERE (
				SELECT count(notification_rule_id)
				FROM user_notification_cycle_state s
				WHERE s.alert_id = c.alert_id AND s.user_id = c.user_id
				LIMIT 1
			) = 0
				AND c.escalation_level != (SELECT escalation_level FROM alerts WHERE id = c.alert_id);

        END;
    $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

ALTER TABLE user_notification_rules ALTER COLUMN id SET DEFAULT gen_random_uuid();


-- +migrate Down


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

CREATE TABLE notifications (
    user_id UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    started_at TIMESTAMP NOT NULL DEFAULT now()
);

ALTER TYPE enum_throttle_type RENAME TO enum_throttle_type_old;
CREATE TYPE enum_throttle_type AS ENUM (
	'notifications'
);

ALTER TABLE throttle ALTER COLUMN action TYPE enum_throttle_type USING action::TEXT::enum_throttle_type;
DROP TYPE enum_throttle_type_old;

CREATE VIEW active_contact_methods AS
    SELECT users.id as user_id, m.id as contact_method_id
        FROM users, user_contact_methods m, user_notification_rules r, notifications n
        WHERE m.user_id = users.id
            AND r.user_id = users.id
            AND n.user_id = users.id
            AND r.contact_method_id = m.id
            AND ((r.delay_minutes::text||' minutes')::interval + n.started_at) < now();

DROP VIEW needs_notification_sent;



DROP VIEW on_call_alert_users;
-- Old version
CREATE OR REPLACE VIEW on_call_alert_users AS
WITH alert_users AS (
		    SELECT act.user_id,
		    act.schedule_id,
		    a.id AS alert_id,
		    a.status
		    FROM alerts a,
		    service s,
		    alert_escalation_levels lvl,
		    escalation_policy_step step,
		    escalation_policy_actions act
		    WHERE ((s.id = a.service_id) AND (step.escalation_policy_id = s.escalation_policy_id) AND (step.step_number = lvl.relative_level) AND (a.status <> 'closed'::enum_alert_status) AND (act.escalation_policy_step_id = step.id))
		    GROUP BY act.user_id, act.schedule_id, a.id
		    )
		    SELECT DISTINCT au.alert_id,
		    au.status,
		    CASE
		    WHEN (au.user_id IS NULL) THEN oc.user_id
		    ELSE au.user_id
		    END AS user_id
		    FROM (alert_users au
		    LEFT JOIN on_call oc ON ((au.schedule_id = oc.schedule_id)));

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION add_notifications() RETURNS TRIGGER AS
    $$
        BEGIN
            INSERT INTO notifications (user_id)
                SELECT user_id FROM on_call_alert_users
                WHERE alert_id = NEW.id AND status = 'triggered'::enum_alert_status
                LIMIT 1
                ON CONFLICT DO NOTHING;

            DELETE FROM notifications WHERE user_id NOT IN (SELECT user_id FROM on_call_alert_users WHERE status = 'triggered'::enum_alert_status AND user_id = user_id);
            RETURN NEW;
        END;
    $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

DROP VIEW user_notification_cycle_state;
DROP TABLE user_notification_cycles;

CREATE TRIGGER add_notifications_alert_changed
    AFTER UPDATE OR INSERT ON alerts
    FOR EACH ROW
    EXECUTE PROCEDURE add_notifications();

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION update_notifications() RETURNS VOID AS
    $$
        BEGIN
        INSERT INTO notifications (user_id)
            SELECT user_id FROM on_call_alert_users
            WHERE status = 'triggered'::enum_alert_status
            GROUP BY user_id
            ON CONFLICT DO NOTHING;

        DELETE FROM notifications WHERE user_id NOT IN (SELECT user_id FROM on_call_alert_users WHERE status = 'triggered'::enum_alert_status AND user_id = user_id);
        END;
    $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

SELECT update_notifications();

ALTER TABLE sent_notifications DROP COLUMN cycle_id;
ALTER TABLE sent_notifications DROP COLUMN notification_rule_id;
ALTER TABLE sent_notifications ADD UNIQUE(alert_id, contact_method_id);
ALTER TABLE user_notification_rules DROP COLUMN created_at;
ALTER TABLE user_notification_rules ALTER id DROP DEFAULT;

ALTER TABLE escalation_policy ALTER description DROP DEFAULT;
	-- ALTER COLUMN repeat DROP DEFAULT;

ALTER TABLE service ALTER description DROP DEFAULT;

DROP FUNCTION update_notification_cycles();
ALTER TABLE sent_notifications ADD CONSTRAINT sent_notifications_pkey PRIMARY KEY (id);


ALTER TABLE ONLY twilio_voice_callbacks
		    ADD CONSTRAINT twilio_voice_callbacks_callback_id_fkey FOREIGN KEY (callback_id) REFERENCES sent_notifications(id) ON DELETE CASCADE;
ALTER TABLE ONLY twilio_sms_callbacks
		    ADD CONSTRAINT twilio_sms_callbacks_callback_id_fkey FOREIGN KEY (callback_id) REFERENCES sent_notifications(id) ON DELETE CASCADE;
DROP INDEX sent_notifications_id_idx;

CREATE VIEW needs_notification_sent AS SELECT trig.alert_id,
		    acm.contact_method_id,
		    cm.type,
		    cm.value,
		    a.description,
		    s.name AS service_name
		    FROM active_contact_methods acm,
		    on_call_alert_users trig,
		    user_contact_methods cm,
		    alerts a,
		    service s
		    WHERE ((acm.user_id = trig.user_id) AND (acm.user_id = trig.user_id) AND (cm.id = acm.contact_method_id) AND (cm.disabled = false) AND (a.id = trig.alert_id) AND (trig.status = 'triggered'::enum_alert_status) AND (s.id = a.service_id) AND (NOT (EXISTS ( SELECT sent_notifications.id
		    FROM sent_notifications
		    WHERE ((sent_notifications.alert_id = trig.alert_id) AND (sent_notifications.contact_method_id = acm.contact_method_id) AND (sent_notifications.sent_at IS NOT NULL))))));
