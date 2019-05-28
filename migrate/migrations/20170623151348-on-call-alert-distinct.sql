
-- +migrate Up
CREATE OR REPLACE VIEW on_call_alert_users AS
  WITH alert_users AS (
			SELECT act.user_id, act.schedule_id, a.id as alert_id, a.status
			FROM
				alerts a,
				service s,
				alert_escalation_levels lvl,
				escalation_policy_step step,
				escalation_policy_actions act
			WHERE s.id = a.service_id
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
			END
		FROM alert_users au
		LEFT JOIN on_call oc ON au.schedule_id = oc.schedule_id;

-- +migrate Down

SELECT 1;
