
-- +migrate Up


WITH rotation_details AS (
    SELECT
        rotations.id,
        rotations.start_time,
        (((rotations.shift_length)::text ||
        CASE
            WHEN (rotations.type = 'hourly') THEN ' hours'
            WHEN (rotations.type = 'daily') THEN ' days'
            ELSE ' weeks'
        END))::interval AS shift,
        ((
        CASE
            WHEN (type = 'hourly') THEN (date_part('epoch', ((now() at time zone time_zone) - (start_time at time zone time_zone)))::bigint / 3600) -- number of hours
            WHEN (type = 'daily') THEN date_part('days', ((now() at time zone time_zone) - (start_time at time zone time_zone)))::bigint -- number of days
            ELSE (date_part('days'::text, ((now() at time zone time_zone) - (start_time at time zone time_zone)))::bigint / 7) -- number of weeks
        END / shift_length)) AS shift_number
    FROM
        rotations
    WHERE start_time <= now()
),
p_count AS (
    SELECT
        part.rotation_id,
        count(part.id) AS count
    FROM rotation_details rot
    JOIN rotation_participants part ON part.rotation_id = rot.id
    GROUP BY part.rotation_id
),
current_participant AS (
    SELECT
        part.id,
        part."position",
        pc.rotation_id
    FROM rotation_details rot
    JOIN p_count pc ON pc.rotation_id = rot.id
    JOIN rotation_participants part ON part.rotation_id = rot.id AND part."position" = (rot.shift_number % pc.count)
)
INSERT INTO rotation_state (
  rotation_id,
  rotation_participant_id,
  position,
  shift_start
)
SELECT
    cp.rotation_id,
    cp.id,
    cp.position,
    rd.start_time + (rd.shift * rd.shift_number)::interval
FROM rotation_details rd
JOIN current_participant cp ON cp.rotation_id = rd.id;

-- +migrate Down
TRUNCATE rotation_state;
