-- +migrate Up
-- Finalize the 2024 notification channel dedup (20240628135216-dedup-notif-channels)
-- for schedule_data. The schedule manager engine module used to rewrite
-- V1.OnCallNotificationRules[].ChannelID at runtime (from a deleted duplicate
-- channel ID to the surviving one), tracking progress with the
-- HasMigratedScheduleData flag in engine_processing_versions.state. Doing it in
-- the migration chain means every installation gets its data rewritten,
-- regardless of the version it is upgrading from.
--
-- Only rows containing at least one rule that still references an old_id are
-- touched, and only the ChannelID key of those elements is replaced. Element
-- order, any other keys, and the rest of the document are preserved.
WITH rewritten AS (
    SELECT
        sd.schedule_id,
        jsonb_agg(CASE WHEN d.new_id IS NULL THEN
                t.e
            ELSE
                t.e || jsonb_build_object('ChannelID', d.new_id)
            END ORDER BY t.o) AS rules
    FROM
        schedule_data sd
        -- Rows where the path is missing or not an array produce no elements and are skipped.
        CROSS JOIN LATERAL jsonb_array_elements(CASE WHEN jsonb_typeof(sd.data #> '{V1,OnCallNotificationRules}') = 'array' THEN
                sd.data #> '{V1,OnCallNotificationRules}'
            ELSE
                '[]'::jsonb
            END) WITH ORDINALITY AS t(e, o)
        -- Compare as text (uuid::text is canonical lowercase) so a malformed ChannelID cannot abort the migration.
        LEFT JOIN notification_channel_duplicates d ON d.old_id::text = lower(t.e ->> 'ChannelID')
    GROUP BY
        sd.schedule_id
    HAVING
        count(d.old_id) > 0
)
UPDATE
    schedule_data sd
SET
    data = jsonb_set(sd.data, '{V1,OnCallNotificationRules}', r.rules)
FROM
    rewritten r
WHERE
    sd.schedule_id = r.schedule_id;

-- Any still-running older binary should take its "already migrated" fast path
-- rather than re-scanning schedule_data.
UPDATE
    engine_processing_versions
SET
    state = coalesce(state, '{}'::jsonb) || '{"HasMigratedScheduleData": true}'::jsonb
WHERE
    type_id = 'schedule';

-- +migrate Down
-- The schedule_data rewrite is intentionally not reverted: the duplicate
-- notification_channels rows were already deleted by
-- 20240628135216-dedup-notif-channels, so the rewritten IDs are the correct
-- ones under every version. With the flag cleared, an older binary simply
-- re-runs its idempotent scan and finds nothing left to rewrite.
UPDATE
    engine_processing_versions
SET
    state = state - 'HasMigratedScheduleData'
WHERE
    type_id = 'schedule';
