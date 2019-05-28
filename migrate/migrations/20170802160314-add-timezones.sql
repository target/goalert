
-- +migrate Up
ALTER TABLE alert_logs ALTER "timestamp" TYPE TIMESTAMP WITH TIME ZONE;
ALTER TABLE alerts ALTER last_escalation TYPE TIMESTAMP WITH TIME ZONE;
ALTER TABLE sent_notifications ALTER sent_at TYPE TIMESTAMP WITH TIME ZONE;
ALTER TABLE throttle ALTER last_action_time TYPE TIMESTAMP WITH TIME ZONE;
ALTER TABLE twilio_sms_errors ALTER occurred_at TYPE TIMESTAMP WITH TIME ZONE;
ALTER TABLE twilio_voice_errors ALTER occurred_at TYPE TIMESTAMP WITH TIME ZONE;
ALTER TABLE user_contact_method_locks ALTER "timestamp" TYPE TIMESTAMP WITH TIME ZONE;


DROP VIEW needs_notification_sent;
DROP VIEW user_notification_cycle_state;

ALTER TABLE user_notification_cycles ALTER started_at TYPE TIMESTAMP WITH TIME ZONE;
ALTER TABLE user_notification_rules ALTER created_at TYPE TIMESTAMP WITH TIME ZONE;

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
        AND nr.created_at < c.started_at + (nr.delay_minutes::TEXT||' minutes')::INTERVAL
		AND nr.id NOT IN (
			SELECT notification_rule_id
			FROM sent_notifications
			WHERE alert_id = c.alert_id
				AND cycle_id = c.id
				AND contact_method_id = nr.contact_method_id);

CREATE VIEW needs_notification_sent AS
    SELECT DISTINCT
        cs.alert_id,
        nr.contact_method_id,
        cm.type,
        cm.value,
        a.description,
        s.name AS service_name,
        nr.id AS notification_rule_id,
        cs.escalation_level,
        cs.cycle_id
    FROM
        user_notification_cycle_state cs,
        alerts a,
        user_contact_methods cm,
        user_notification_rules nr,
        services s
    WHERE a.id = cs.alert_id
        AND a.status = 'triggered'::enum_alert_status
        AND cs.escalation_level = a.escalation_level
        AND cm.id = nr.contact_method_id
        AND nr.id = cs.notification_rule_id
        AND s.id = a.service_id
        AND cs.pending
        AND NOT cs.future;


-- +migrate Down
ALTER TABLE alert_logs ALTER "timestamp" TYPE TIMESTAMP WITHOUT TIME ZONE;
ALTER TABLE alerts ALTER last_escalation TYPE TIMESTAMP WITHOUT TIME ZONE;
ALTER TABLE sent_notifications ALTER sent_at TYPE TIMESTAMP WITHOUT TIME ZONE;
ALTER TABLE throttle ALTER last_action_time TYPE TIMESTAMP WITHOUT TIME ZONE;
ALTER TABLE twilio_sms_errors ALTER occurred_at TYPE TIMESTAMP WITHOUT TIME ZONE;
ALTER TABLE twilio_voice_errors ALTER occurred_at TYPE TIMESTAMP WITHOUT TIME ZONE;
ALTER TABLE user_contact_method_locks ALTER "timestamp" TYPE TIMESTAMP WITHOUT TIME ZONE;


DROP VIEW needs_notification_sent;
DROP VIEW user_notification_cycle_state;

ALTER TABLE user_notification_cycles ALTER started_at TYPE TIMESTAMP WITHOUT TIME ZONE;
ALTER TABLE user_notification_rules ALTER created_at TYPE TIMESTAMP WITHOUT TIME ZONE;

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
        AND nr.created_at < c.started_at + (nr.delay_minutes::TEXT||' minutes')::INTERVAL
		AND nr.id NOT IN (
			SELECT notification_rule_id
			FROM sent_notifications
			WHERE alert_id = c.alert_id
				AND cycle_id = c.id
				AND contact_method_id = nr.contact_method_id);

CREATE VIEW needs_notification_sent AS
    SELECT DISTINCT
        cs.alert_id,
        nr.contact_method_id,
        cm.type,
        cm.value,
        a.description,
        s.name AS service_name,
        nr.id AS notification_rule_id,
        cs.escalation_level,
        cs.cycle_id
    FROM
        user_notification_cycle_state cs,
        alerts a,
        user_contact_methods cm,
        user_notification_rules nr,
        services s
    WHERE a.id = cs.alert_id
        AND a.status = 'triggered'::enum_alert_status
        AND cs.escalation_level = a.escalation_level
        AND cm.id = nr.contact_method_id
        AND nr.id = cs.notification_rule_id
        AND s.id = a.service_id
        AND cs.pending
        AND NOT cs.future;