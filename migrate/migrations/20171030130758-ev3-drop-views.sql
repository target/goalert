
-- +migrate Up

drop view on_call_next_rotation;
drop view on_call_alert_users;
drop view on_call;
drop view alert_escalation_policies;
drop view needs_notification_sent;
drop view user_notification_cycle_state;

-- +migrate Down


CREATE OR REPLACE VIEW on_call AS
    WITH rotation_details AS (
        SELECT
            rotations.id,
            rotations.schedule_id,
            rotations.start_time,
            (((rotations.shift_length)::text ||
            CASE
                WHEN (rotations.type = 'hourly') THEN ' hours'
                WHEN (rotations.type = 'daily') THEN ' days'
                ELSE ' weeks'
            END))::interval AS shift,
            ((
            CASE
                WHEN (rotations.type = 'hourly') THEN (date_part('epoch', ((now() at time zone s.time_zone) - (rotations.start_time at time zone s.time_zone)))::bigint / 3600) -- number of hours
                WHEN (rotations.type = 'daily') THEN date_part('days', ((now() at time zone s.time_zone) - (rotations.start_time at time zone s.time_zone)))::bigint -- number of days
                ELSE (date_part('days'::text, ((now() at time zone s.time_zone) - (rotations.start_time at time zone s.time_zone)))::bigint / 7) -- number of weeks
            END / rotations.shift_length)) AS shift_number
        FROM
            rotations,
            schedules s
        WHERE s.id = rotations.schedule_id
            AND rotations.start_time <= now()
    ),
    p_count AS (
        SELECT
            rp.rotation_id,
            count(rp.id) AS count
        FROM
            rotation_participants rp,
            rotation_details d_1
        WHERE (rp.rotation_id = d_1.id)
        GROUP BY rp.rotation_id
    ),
    current_participant AS (
        SELECT
            rp.user_id,
            p.rotation_id
        FROM
            rotation_participants rp,
            rotation_details d_1,
            p_count p
        WHERE ((rp.rotation_id = d_1.id)
            AND (p.rotation_id = rp.rotation_id)
            AND (rp."position" = (d_1.shift_number % p.count)))
    ),
    next_participant AS (
        SELECT
            rp.user_id,
            p.rotation_id
        FROM
            rotation_participants rp,
            rotation_details d_1,
            p_count p
        WHERE ((rp.rotation_id = d_1.id)
            AND (p.rotation_id = rp.rotation_id)
            AND (rp."position" = ((d_1.shift_number + 1) % p.count)))
    )
    SELECT
        d.schedule_id,
        d.id AS rotation_id,
        c.user_id,
        n.user_id AS next_user_id,
        ((d.shift * (d.shift_number)::bigint) + d.start_time) AS start_time,
        ((d.shift * ((d.shift_number + 1))::bigint) + d.start_time) AS end_time,
        d.shift_number
    FROM
        rotation_details d,
        current_participant c,
        next_participant n
    WHERE ((d.id = c.rotation_id)
        AND (c.rotation_id = n.rotation_id));

CREATE OR REPLACE VIEW on_call_alert_users AS
    WITH alert_users AS (
        SELECT s.user_id,
            s.schedule_id,
            s.alert_id,
            a.status,
            a.escalation_level
        FROM
            alerts a,
            alert_escalation_policy_snapshots s
        WHERE s.alert_id = a.id
            AND s.step_number = (a.escalation_level % s.step_max)
            AND a.status <> 'closed'
    )
    SELECT DISTINCT au.alert_id,
        au.status,
        CASE
            WHEN au.user_id IS NULL THEN oc.user_id
            ELSE au.user_id
        END AS user_id,
        au.escalation_level
    FROM alert_users au
    LEFT JOIN on_call oc ON au.schedule_id = oc.schedule_id;

CREATE OR REPLACE VIEW on_call_next_rotation AS
WITH
	p_count AS (
		SELECT rotation_id, count(rp.position)
		FROM
			rotations r,
			rotation_participants rp
		WHERE r.id = rp.rotation_id
		GROUP BY rotation_id
	)
SELECT
	oc.schedule_id,
	rp.rotation_id,
	rp.user_id,
	oc.next_user_id,
	
	(
		CASE WHEN oc.shift_number % p.count < rp.position
			THEN rp.position-(oc.shift_number % p.count)
			ELSE rp.position-(oc.shift_number % p.count)+p.count
		END
	) * (oc.end_time-oc.start_time) + oc.start_time start_time,

	(
		CASE WHEN oc.shift_number % p.count < rp.position
			THEN rp.position-(oc.shift_number % p.count)
			ELSE rp.position-(oc.shift_number % p.count)+p.count
		END
	) * (oc.end_time-oc.start_time) + oc.end_time end_time,

	(
		CASE WHEN oc.shift_number % p.count < rp.position
			THEN rp.position-(oc.shift_number % p.count)
			ELSE rp.position-(oc.shift_number % p.count)+p.count
		END
	) + oc.shift_number shift_number

FROM
	rotations r,
	rotation_participants rp,
	p_count p,
	on_call oc
WHERE p.rotation_id = r.id
	AND rp.rotation_id = r.id
	AND oc.rotation_id = r.id
GROUP BY
	rp.user_id,
	rp.rotation_id,
	oc.shift_number,
	p.count,
	shift_length,
	type,
	oc.start_time,
	oc.end_time,
	rp.position,
	oc.schedule_id,
	oc.next_user_id;

CREATE VIEW alert_escalation_policies AS
 WITH step_max AS (
         SELECT escalation_policy_steps.escalation_policy_id,
            count(escalation_policy_steps.step_number) AS step_max
           FROM escalation_policy_steps
          GROUP BY escalation_policy_steps.escalation_policy_id
        )
 SELECT a.id AS alert_id,
    step.step_number,
    m.step_max,
    (step.delay::text || ' minutes'::text)::interval AS step_delay,
    e.repeat,
    act.user_id,
    act.schedule_id
   FROM alerts a,
    escalation_policies e,
    escalation_policy_steps step,
    step_max m,
    escalation_policy_actions act,
    services svc
  WHERE a.service_id = svc.id AND e.id = svc.escalation_policy_id AND step.escalation_policy_id = m.escalation_policy_id AND step.escalation_policy_id = svc.escalation_policy_id AND act.escalation_policy_step_id = step.id;


CREATE VIEW user_notification_cycle_state AS
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
        AND NOT cs.future
        AND cm.disabled = FALSE;
