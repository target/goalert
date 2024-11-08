-- name: CleanMgrTrimOverrides :execrows
DELETE FROM user_overrides
WHERE id = ANY (
        SELECT
            id
        FROM
            user_overrides o
        WHERE
            o.end_time < $1
        LIMIT 100
        FOR UPDATE
            SKIP LOCKED);

-- name: CleanMgrTrimSchedOnCall :execrows
DELETE FROM schedule_on_call_users
WHERE id = ANY (
        SELECT
            id
        FROM
            schedule_on_call_users s
        WHERE
            s.end_time < $1::timestamptz
        LIMIT 100
        FOR UPDATE
            SKIP LOCKED);

-- name: CleanMgrTrimEPOnCall :execrows
DELETE FROM ep_step_on_call_users
WHERE id = ANY (
        SELECT
            id
        FROM
            ep_step_on_call_users ep
        WHERE
            ep.end_time < $1::timestamptz
        LIMIT 100
        FOR UPDATE
            SKIP LOCKED);

-- name: CleanMgrTrimClosedAlerts :execrows
DELETE FROM alerts
WHERE id = ANY (
        SELECT
            id
        FROM
            alerts
        WHERE
            status = 'closed'
            AND created_at < $1::timestamptz
        ORDER BY
            id
        LIMIT 100
        FOR UPDATE
            SKIP LOCKED);

