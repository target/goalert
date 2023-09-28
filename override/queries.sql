-- name: OverrideSearch :many
WITH AFTER AS (
    SELECT
        id,
        start_time,
        end_time
    FROM
        user_overrides
    WHERE
        id = sqlc.narg(after_id)::uuid
)
SELECT
    o.id,
    o.start_time,
    o.end_time,
    add_user_id,
    remove_user_id,
    tgt_schedule_id
FROM
    user_overrides o
    LEFT JOIN AFTER ON TRUE
WHERE (TRUE
    OR o.id <> ALL (@omit::uuid[]))
AND (sqlc.narg(schedule_id)::uuid ISNULL
    OR o.tgt_schedule_id = @schedule_id)
AND (@any_user_id::uuid[] ISNULL
    OR add_user_id = ANY (@any_user_id::uuid[])
    OR remove_user_id = ANY (@any_user_id::uuid[]))
AND (@add_user_id::uuid[] ISNULL
    OR add_user_id = ANY (@add_user_id::uuid[]))
AND (@remove_user_id::uuid[] ISNULL
    OR remove_user_id = ANY (@remove_user_id::uuid[]))
AND (
    /* only include overrides that end after the search start */
    sqlc.narg(search_start)::timestamptz ISNULL
    OR o.end_time > @search_start)
AND (
    /* only include overrides that start before/within the search end */
    sqlc.narg(search_end)::timestamptz ISNULL
    OR o.start_time <= @search_start)
AND (
    /* resume search after specified "cursor" override */
    @after_id::uuid ISNULL
    OR (o.start_time > after.start_time
        OR (o.start_time = after.start_time
            AND o.end_time > after.end_time)
        OR (o.start_time = after.start_time
            AND o.end_time = after.end_time
            AND o.id > after.id)))
ORDER BY
    o.start_time,
    o.end_time,
    o.id
LIMIT 150;

