
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
		
		SELECT
			au.alert_id,
			au.status,
			CASE WHEN au.user_id IS NULL THEN oc.user_id
			ELSE au.user_id
			END
		FROM alert_users au
		LEFT JOIN on_call oc ON au.schedule_id = oc.schedule_id;


CREATE OR REPLACE VIEW on_call AS
	WITH rotation_details AS (
        SELECT 
            id,
            schedule_id,
            start_time,
            
            (shift_length::TEXT||CASE
                WHEN type='hourly'::enum_rotation_type THEN ' hours'
                WHEN type='daily'::enum_rotation_type THEN ' days'
                ELSE ' weeks'
            END)::interval shift,
            
            (CASE
                WHEN type='hourly'::enum_rotation_type THEN extract(epoch from now()-start_time)/3600
                WHEN type='daily'::enum_rotation_type THEN extract(days from now()-start_time)
                ELSE extract(days from now()-start_time)/7
            END/shift_length)::BIGINT shift_number
            
            FROM rotations
    ),
    p_count AS (
        SELECT rotation_id, count(rp.id)
        FROM 
            rotation_participants rp,
            rotation_details d
        WHERE rp.rotation_id = d.id
        GROUP BY rotation_id
    ),
    current_participant AS (
        SELECT user_id, p.rotation_id
        FROM 
            rotation_participants rp,
            rotation_details d,
            p_count p
        WHERE rp.rotation_id = d.id
        	AND p.rotation_id = rp.rotation_id
            AND rp.position = d.shift_number % p.count
    ),
    next_participant AS (
        SELECT user_id, p.rotation_id
        FROM 
            rotation_participants rp,
            rotation_details d,
            p_count p
        WHERE rp.rotation_id = d.id
        	AND p.rotation_id = rp.rotation_id
            AND rp.position = (d.shift_number+1) % p.count
    )
    SELECT 
        d.schedule_id,
        d.id rotation_id,
        c.user_id,
        n.user_id next_user_id,
        (d.shift*d.shift_number)+d.start_time start_time,
        (d.shift*(d.shift_number+1))+d.start_time end_time,
        d.shift_number
    FROM
        rotation_details d,
        current_participant c,
        next_participant n
    WHERE d.id = c.rotation_id
    	AND c.rotation_id = n.rotation_id;



-- +migrate Down

SELECT 1;
