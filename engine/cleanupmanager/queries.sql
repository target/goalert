-- name: CleanupMgrDeleteOldAlerts :execrows
-- CleanupMgrDeleteOldAlerts will delete old alerts from the alerts table that are closed and older than the given number of days before now.
DELETE FROM alerts
WHERE id = ANY (
        SELECT
            id
        FROM
            alerts a
        WHERE
            status = 'closed'
            AND a.created_at < now() -(sqlc.arg(stale_threshold_days)::bigint * '1 day'::interval)
        ORDER BY
            id
        LIMIT 100
        FOR UPDATE
            SKIP LOCKED);

-- name: CleanupMgrFindStaleAlerts :many
-- CleanupMgrFindStaleAlerts will find alerts that are triggered or active and have no activity in specified number of days.
SELECT
    id
FROM
    alerts a
WHERE (a.status = 'triggered'
    OR (sqlc.arg(include_acked)
        AND a.status = 'active'))
AND created_at <= now() - '1 day'::interval * sqlc.arg(auto_close_threshold_days)
AND NOT EXISTS (
    SELECT
        1
    FROM
        alert_logs log
    WHERE
        timestamp > now() - '1 day'::interval * sqlc.arg(auto_close_threshold_days)
        AND log.alert_id = a.id)
LIMIT 100;

-- name: CleanupMgrDeleteOldOverrides :execrows
-- CleanupMgrDeleteOldOverrides will delete old overrides from the user_overrides table that are older than the given number of days before now.
DELETE FROM user_overrides
WHERE id = ANY (
        SELECT
            id
        FROM
            user_overrides
        WHERE
            end_time <(now() - '1 day'::interval * sqlc.arg(history_threshold_days))
        LIMIT 100
        FOR UPDATE
            SKIP LOCKED);

-- name: CleanupMgrDeleteOldScheduleShifts :execrows
-- CleanupMgrDeleteOldScheduleShifts will delete old schedule shifts from the schedule_on_call_users table that are older than the given number of days before now.
DELETE FROM schedule_on_call_users
WHERE id = ANY (
        SELECT
            id
        FROM
            schedule_on_call_users
        WHERE
            end_time <(now() - '1 day'::interval * sqlc.arg(history_threshold_days))
        LIMIT 100
        FOR UPDATE
            SKIP LOCKED);

-- name: CleanupMgrDeleteOldStepShifts :execrows
-- CleanupMgrDeleteOldStepShifts will delete old EP step shifts from the ep_step_on_call_users table that are older than the given number of days before now.
DELETE FROM ep_step_on_call_users
WHERE id = ANY (
        SELECT
            id
        FROM
            ep_step_on_call_users
        WHERE
            end_time <(now() - '1 day'::interval * sqlc.arg(history_threshold_days))
        LIMIT 100
        FOR UPDATE
            SKIP LOCKED);

-- name: CleanupMgrScheduleNeedsCleanup :many
-- CleanupMgrScheduleNeedsCleanup will find schedules that need to be cleaned up. The last_cleanup_at field is used to ensure we clean up each schedule data at most once per interval.
SELECT
    schedule_id
FROM
    schedule_data
WHERE
    data NOTNULL
    AND (last_cleanup_at ISNULL
        OR last_cleanup_at <= now() - '1 day'::interval * sqlc.arg(cleanup_interval_days)::int)
ORDER BY
    last_cleanup_at ASC nulls FIRST;

-- name: CleanupMgrScheduleData :one
-- CleanupMgrScheduleData will select the schedule data for the given schedule id.
SELECT
    data
FROM
    schedule_data
WHERE
    schedule_id = $1
FOR UPDATE
LIMIT 1;

-- name: CleanupMgrUpdateScheduleData :exec
-- CleanupMgrUpdateScheduleData will update the last_cleanup_at and data fields in the schedule_data table.
UPDATE
    schedule_data
SET
    last_cleanup_at = now(),
    data = $2
WHERE
    schedule_id = $1;

-- name: CleanupMgrScheduleDataSkip :exec
-- CleanupMgrScheduleDataSkip will update the last_cleanup_at field in the schedule_data table.
UPDATE
    schedule_data
SET
    last_cleanup_at = now()
WHERE
    schedule_id = $1;

-- name: CleanupMgrVerifyUsers :many
-- CleanupMgrVerifyUsers will verify that the given user ids exist in the users table.
SELECT
    id
FROM
    users
WHERE
    id = ANY (sqlc.arg(user_ids)::uuid[]);

-- name: CleanupMgrAlertLogsMinMax :one
-- CleanupMgrAlertLogsMinMax will find the minimum and maximum id of the alert_logs table.
SELECT
    coalesce(min(id), 0)::bigint AS min_id,
    coalesce(max(id), 0)::bigint AS max_id
FROM
    alert_logs;

-- name: CleanupAlertLogs :one
WITH scope AS (
    SELECT
        id
    FROM
        alert_logs l
    WHERE
        l.id BETWEEN @start_id AND @end_id - 1
    ORDER BY
        l.id
    LIMIT @batch_size
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
                    scope OFFSET @batch_size - 1
                LIMIT 1;

