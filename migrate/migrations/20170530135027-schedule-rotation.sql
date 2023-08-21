
-- +migrate Up
CREATE TABLE schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    time_zone TEXT NOT NULL
);

INSERT INTO schedules (id, name, description, time_zone)
    SELECT id::UUID, name, description, 'America/Chicago' FROM schedule;

CREATE TYPE enum_rotation_type AS ENUM (
    'monthly',
    'weekly',
    'daily',
    'hourly'
);

CREATE TABLE rotations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id UUID NOT NULL REFERENCES schedules (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    type enum_rotation_type NOT NULL,
    start_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    shift_length BIGINT NOT NULL DEFAULT 1,
    UNIQUE (schedule_id, name)
);

INSERT INTO rotations (id, start_time, name, type, description, shift_length, schedule_id)
    SELECT id::UUID, effective_date, name, rotation_type::enum_rotation_type, description, shift_length, schedule_id::UUID FROM schedule_layer;

CREATE TABLE rotation_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rotation_id UUID NOT NULL REFERENCES rotations (id) ON DELETE CASCADE,
    position INT NOT NULL,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    UNIQUE (rotation_id, position) DEFERRABLE INITIALLY DEFERRED
);

INSERT INTO rotation_participants (rotation_id, position, user_id)
    SELECT schedule_layer_id::UUID, step_number-1, user_id FROM schedule_layer_user;


-- Update escalation_policy_actions references
ALTER TABLE escalation_policy_actions RENAME COLUMN schedule_id TO old_schedule_id;
ALTER TABLE escalation_policy_actions ADD COLUMN schedule_id UUID REFERENCES schedules (id) ON DELETE CASCADE;
UPDATE escalation_policy_actions SET schedule_id = old_schedule_id::UUID WHERE old_schedule_id IS NOT NULL;
ALTER TABLE escalation_policy_actions DROP COLUMN old_schedule_id;
ALTER TABLE escalation_policy_actions ADD UNIQUE(escalation_policy_step_id, schedule_id, user_id);
ALTER TABLE escalation_policy_actions ADD CHECK((schedule_id IS NOT NULL AND user_id IS NULL) OR (user_id IS NOT NULL AND schedule_id IS NULL));

DROP TABLE schedule_layer_user;
DROP TABLE schedule_layer;
DROP TABLE schedule;


CREATE VIEW on_call AS
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
    ), p_count AS (
        SELECT count(rp.id)
        FROM 
            rotation_participants rp,
            rotation_details d
        WHERE rp.rotation_id = d.id
    ),
    current_participant AS (
        SELECT user_id
        FROM 
            rotation_participants rp,
            rotation_details d,
            p_count p
        WHERE rp.rotation_id = d.id
            AND rp.position = d.shift_number % p.count
        LIMIT 1
    ),
    next_participant AS (
        SELECT user_id
        FROM 
            rotation_participants rp,
            rotation_details d,
            p_count p
        WHERE rp.rotation_id = d.id
            AND rp.position = (d.shift_number+1) % p.count
        LIMIT 1
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
        next_participant n;

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION move_rotation_position(_id UUID, _new_pos INT) RETURNS VOID AS
    $$
    DECLARE
        _old_pos INT;
        _rid UUID;
    BEGIN
        SELECT position,rotation_id into _old_pos, _rid FROM rotation_participants WHERE id = _id;
        IF _old_pos > _new_pos THEN
            UPDATE rotation_participants SET position = position + 1 WHERE rotation_id = _rid AND position < _old_pos AND position >= _new_pos;
        ELSE
            UPDATE rotation_participants SET position = position - 1 WHERE rotation_id = _rid AND position > _old_pos AND position <= _new_pos;
        END IF;
        UPDATE rotation_participants SET position = _new_pos WHERE id = _id;
    END;
    $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd


-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION remove_rotation_participant(_id UUID) RETURNS UUID AS
    $$
    DECLARE
        _old_pos INT;
        _rid UUID;
    BEGIN
        SELECT position,rotation_id into _old_pos, _rid FROM rotation_participants WHERE id = _id;
        DELETE FROM rotation_participants WHERE id = _id;
        UPDATE rotation_participants SET position = position - 1 WHERE rotation_id = _rid AND position > _old_pos;
        RETURN _rid;
    END;
    $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd


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
		FROM alert_users au, on_call oc
		WHERE oc.schedule_id = au.schedule_id OR au.schedule_id IS NULL;


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

CREATE OR REPLACE VIEW needs_notification_sent AS
    SELECT trig.alert_id, acm.contact_method_id, cm.type, cm.value, a.description, s.name as service_name
    FROM active_contact_methods acm, on_call_alert_users trig, user_contact_methods cm, alerts a, service s
    WHERE acm.user_id = trig.user_id
        AND acm.user_id = trig.user_id
        AND cm.id = acm.contact_method_id
        AND cm.disabled = FALSE
        AND a.id = trig.alert_id
        AND trig.status = 'triggered'::enum_alert_status
        AND s.id = a.service_id
        AND NOT EXISTS (
            SELECT id
            FROM sent_notifications
            WHERE alert_id = trig.alert_id
                AND contact_method_id = acm.contact_method_id
                AND sent_at IS NOT NULL
        );

DROP VIEW triggered_alert_users;



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


-- +migrate Down

CREATE TABLE schedule (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT now(),
    name TEXT,
    description TEXT,
    time_zone INT
);

INSERT INTO schedule (id, name, description, time_zone)
    SELECT s.id::TEXT, s.name, s.description, date_part('hour', tz.utc_offset) 
        FROM schedules s, pg_timezone_names tz
        WHERE tz.name = s.time_zone;

CREATE TABLE schedule_layer (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP DEFAULT now(),
    effective_date TIMESTAMP,
    description TEXT,
    handoff_day INT,
    handoff_time TEXT,
    name TEXT,
    rotation_type TEXT,
    shift_length INT,
    shift_length_unit,
    schedule_id TEXT REFERENCES schedule (id)
);

INSERT INTO schedule_layer (id, effective_date, description, handoff_day, handoff_time, name, rotation_type, shift_length, shift_length_unit, schedule_id)
    SELECT id::TEXT, start, description, EXTRACT(DOW FROM TIMESTAMP start), date_part('hour', start)::TEXT|':'|date_part('minute', start), name, type::TEXT, shift_length, 'hour', schedule_id::TEXT
        FROM rotations;

CREATE TABLE schedule_layer_user (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP DEFAULT now(),
    step_number INT,
    user_id UUID REFERENCES users (id),
    schedule_layer_id TEXT REFERENCES schedule_layer (id)
);

INSERT INTO schedule_layer_user (step_number, user_id, schedule_layer_id)
    SELECT position+1, user_id, rotation_id::TEXT;


ALTER TABLE escalation_policy_actions RENAME COLUMN schedule_id TO old_schedule_id;
ALTER TABLE escalation_policy_actions ADD COLUMN schedule_id UUID REFERENCES schedules (id) ON DELETE CASCADE;
UPDATE escalation_policy_actions SET schedule_id = old_schedule_id::TEXT WHERE old_schedule_id IS NOT NULL;
ALTER TABLE escalation_policy_actions DROP COLUMN old_schedule_id;
ALTER TABLE escalation_policy_actions ADD UNIQUE(escalation_policy_step_id, schedule_id, user_id);
ALTER TABLE escalation_policy_actions ADD CHECK((schedule_id IS NOT NULL AND user_id IS NULL) OR (user_id IS NOT NULL AND schedule_id IS NULL));

DROP TABLE shift_creation_locks;
DROP TABLE on_call;
DROP TABLE rotation_participants;
DROP TABLE rotations;
DROP TYPE enum_rotation_type;
DROP TABLE schedules;
