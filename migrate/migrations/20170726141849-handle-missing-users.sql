
-- +migrate Up


-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION update_notification_cycles() RETURNS VOID AS
    $$
        BEGIN
			INSERT INTO user_notification_cycles (user_id, alert_id, escalation_level)
			SELECT user_id, alert_id, escalation_level
			FROM on_call_alert_users
			WHERE status = 'triggered'
                AND user_id IS NOT NULL
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

-- +migrate Down

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
