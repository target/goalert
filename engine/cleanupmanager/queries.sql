-- name: CleanMgrClosedAlerts :execrows
-- CleanMgrClosedAlerts deletes closed alerts older than the specified cutoff time.
DELETE FROM alerts
WHERE id = ANY (
        SELECT
            id
        FROM
            alerts a
        WHERE
            status = 'closed'
            AND a.created_at < $1
        ORDER BY
            id
        LIMIT 100
        FOR UPDATE
            SKIP LOCKED);

-- name: CleanMgrCalSubs :execrows
-- CleanMgrCalSubs deletes calendar subscriptions not used or updated since the specified cutoff time.
UPDATE
    user_calendar_subscriptions
SET
    disabled = TRUE
WHERE
    id = ANY (
        SELECT
            id
        FROM
            user_calendar_subscriptions sub
        WHERE
            greatest(sub.last_access, sub.last_update) < $1::timestamptz
        ORDER BY
            id
        LIMIT 100
        FOR UPDATE
            SKIP LOCKED);

-- name: CleanMgrUserSessions :execrows
-- CleanMgrUserSessions deletes user sessions not used since the specified cutoff time.
DELETE FROM auth_user_sessions
WHERE ctid = ANY (
        SELECT
            ctid
        FROM
            auth_user_sessions s
        WHERE
            s.last_access_at < $1
        LIMIT 100
        FOR UPDATE
            SKIP LOCKED);

-- name: CleanMgrUserOverrides :execrows
-- CleanMgrUserOverrides deletes user overrides that have ended before the specified cutoff time.
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

-- name: CleanMgrSchedHistory :execrows
-- CleanMgrSchedHistory deletes schedule history entries older than the specified cutoff time.
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

-- name: CleanMgrEPHistory :execrows
-- CleanMgrEPHistory deletes EP step on call history entries older than the specified cutoff time.
DELETE FROM ep_step_on_call_users
WHERE id = ANY (
        SELECT
            id
        FROM
            ep_step_on_call_users e
        WHERE
            e.end_time < $1::timestamptz
        LIMIT 100
        FOR UPDATE
            SKIP LOCKED);

-- name: CleanMgrSetTimeout :exec
-- CleanMgrSetTimeout sets the statement timeout to 3000ms.
SET LOCAL statement_timeout = 3000;

-- name: CleanMgrUserIDs :many
-- CleanMgrUserIDs returns the IDs of all users.
SELECT
    id
FROM
    users;

-- name: CleanMgrStaleAlertIDs :many
-- CleanMgrStaleAlertIDs returns the IDs of all alerts that are either triggered or active and have not been updated since the specified cutoff time.
SELECT
    id
FROM
    alerts a
WHERE (a.status = 'triggered'
    OR (@include_active
        AND a.status = 'active'))
AND created_at <= @cutoff
AND NOT EXISTS (
    SELECT
        1
    FROM
        alert_logs log
    WHERE
        timestamp > @cutoff
        AND log.alert_id = a.id)
LIMIT 100;

-- name: CleanMgrScheduleData :many
-- CleanMgrScheduleData returns the IDs and data of all schedule data entries that have not been cleaned up since the specified cutoff time.
SELECT
    schedule_id,
    data
FROM
    schedule_data
WHERE
    data NOTNULL
    AND (last_cleanup_at ISNULL
        OR last_cleanup_at < @cutoff::timestamptz)
ORDER BY
    last_cleanup_at ASC nulls FIRST
FOR UPDATE
    SKIP LOCKED
LIMIT 100;

-- name: CleanMgrUpdateScheduleData :exec
-- CleanMgrUpdateScheduleData updates the data of a schedule data entry.
UPDATE
    schedule_data
SET
    last_cleanup_at = now(),
    data = $2
WHERE
    schedule_id = $1;

-- name: CleanMgrAlertLogs :one
-- CleanMgrAlertLogs deletes a range of alert logs that do not have a corresponding alert entry, returning the ID of the next entry in the range.
WITH scope AS (
    SELECT
        id
    FROM
        alert_logs l
    WHERE
        l.id > $1
    ORDER BY
        id
    LIMIT 100
),
id_range AS (
    SELECT
        min(id),
        max(id)
    FROM
        scope
),
_delete AS (
    DELETE FROM alert_logs
    WHERE id = ANY (
            SELECT
                id
            FROM
                alert_logs
            WHERE
                id BETWEEN (
                    SELECT
                        min
                    FROM
                        id_range)
                    AND (
                        SELECT
                            max
                        FROM
                            id_range)
                        AND NOT EXISTS (
                            SELECT
                                1
                            FROM
                                alerts
                            WHERE
                                alert_id = id)
                            FOR UPDATE
                                SKIP LOCKED))
                SELECT
                    id
                FROM
                    scope offset 99;

