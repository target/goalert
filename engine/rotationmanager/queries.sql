-- name: RotMgrRotationData :one
-- Get rotation data for a given rotation ID
SELECT
    now()::timestamptz AS now,
    rot.id,
    rot.type,
    rot.start_time,
    rot.shift_length,
    rot.time_zone,
    state.position,
    state.shift_start,
    ARRAY (
        SELECT
            p.id
        FROM
            rotation_participants p
        WHERE
            p.rotation_id = rot.id
        ORDER BY
            position)::uuid[] AS participants
    FROM
        rotations rot
    LEFT JOIN rotation_state state ON rot.id = state.rotation_id
WHERE
    rot.id = @rotation_id;

-- name: RotMgrStart :exec
-- Start a rotation.
INSERT INTO rotation_state(rotation_id, position, shift_start, rotation_participant_id)
SELECT
    p.rotation_id,
    0,
    now(),
    id
FROM
    rotation_participants p
WHERE
    p.rotation_id = @rotation_id
    AND position = 0;

-- name: RotMgrEnd :exec
-- End a rotation.
DELETE FROM rotation_state
WHERE rotation_id = @rotation_id;

-- name: RotMgrUpdate :exec
-- Update the rotation state.
UPDATE
    rotation_state
SET
    position = @position,
    shift_start = now(),
    rotation_participant_id = @rotation_participant_id
WHERE
    rotation_id = @rotation_id;

