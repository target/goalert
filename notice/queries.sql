-- name: CountUnackedAlertsByService :one
SELECT
    count(*)
FROM
    alerts
WHERE
    service_id = $1::uuid
    AND status = 'triggered';

