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
            AND a.created_at < now() -(sqlc.arg(cleanup_days)::bigint * '1 day'::interval)
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
AND created_at <= now() - '1 day'::interval * sqlc.arg(auto_close_days)
AND NOT EXISTS (
    SELECT
        1
    FROM
        alert_logs log
    WHERE
        timestamp > now() - '1 day'::interval * sqlc.arg(auto_close_days)
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
            end_time <(now() - '1 day'::interval * sqlc.arg(cleanup_days))
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
            end_time <(now() - '1 day'::interval * sqlc.arg(cleanup_days))
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
            end_time <(now() - '1 day'::interval * sqlc.arg(cleanup_days))
        LIMIT 100
        FOR UPDATE
            SKIP LOCKED);

