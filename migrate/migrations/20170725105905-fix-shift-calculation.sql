
-- +migrate Up

-- need to coerce to bigint (instead of default of double precision) to get the correct shift_number.
-- need to compare time intervals in the correct time zone

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


-- +migrate Down


CREATE OR REPLACE VIEW on_call AS
    WITH rotation_details AS (
        SELECT
            rotations.id,
            rotations.schedule_id,
            rotations.start_time,
            (((rotations.shift_length)::text ||
            CASE
                WHEN (rotations.type = 'hourly'::enum_rotation_type) THEN ' hours'::text
                WHEN (rotations.type = 'daily'::enum_rotation_type) THEN ' days'::text
                ELSE ' weeks'::text
            END))::interval AS shift,
            ((
            CASE
                WHEN (rotations.type = 'hourly'::enum_rotation_type) THEN (date_part('epoch'::text, (now() - rotations.start_time)) / (3600)::double precision)
                WHEN (rotations.type = 'daily'::enum_rotation_type) THEN date_part('days'::text, (now() - rotations.start_time))
                ELSE (date_part('days'::text, (now() - rotations.start_time)) / (7)::double precision)
            END / (rotations.shift_length)::double precision))::bigint AS shift_number
        FROM rotations
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
        ((d.shift * (d.shift_number)::double precision) + d.start_time) AS start_time,
        ((d.shift * ((d.shift_number + 1))::double precision) + d.start_time) AS end_time,
        d.shift_number
    FROM
        rotation_details d,
        current_participant c,
        next_participant n
    WHERE ((d.id = c.rotation_id)
        AND (c.rotation_id = n.rotation_id));
