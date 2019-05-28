
-- +migrate Up

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
        AND nr.created_at <= c.started_at + (nr.delay_minutes::TEXT||' minutes')::INTERVAL
		AND nr.id NOT IN (
			SELECT notification_rule_id
			FROM sent_notifications
			WHERE alert_id = c.alert_id
				AND cycle_id = c.id
				AND contact_method_id = nr.contact_method_id);

-- +migrate Down

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
