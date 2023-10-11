-- name: NoticeUnackedAlertsByService :one
SELECT
    count(*),
    (
        SELECT
            max
        FROM
            config_limits
        WHERE
            id = 'unacked_alerts_per_service'
    )
FROM
    alerts
WHERE
    service_id = $1::uuid
    AND status = 'triggered';
