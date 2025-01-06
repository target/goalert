-- name: RotationMgrLock :exec
-- Locks tables for exclusive access to the rotation manager.
LOCK rotation_participants,
rotation_state IN exclusive mode;

-- name: RotationMgrUpdateState :exec
-- Updates the state of the rotation.
UPDATE
    rotation_state
SET
    shift_start = now(),
    rotation_participant_id =(
        SELECT
            id
        FROM
            rotation_participants p
        WHERE
            p.rotation_id = $1
            AND p.position = $2),
    version = 2
WHERE
    rotation_id = $1;

-- name: RotationMgrGetConfig :many
-- Gets the configuration of all rotations.
SELECT
    rot.id,
    rot."type",
    rot.start_time,
    rot.shift_length,
    rot.time_zone,
    state.shift_start,
    state."position",
    rot.participant_count,
    state.version
FROM
    rotations rot
    JOIN rotation_state state ON state.rotation_id = rot.id
FOR UPDATE
    SKIP LOCKED;

