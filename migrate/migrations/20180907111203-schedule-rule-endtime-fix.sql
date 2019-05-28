
-- +migrate Up
UPDATE engine_processing_versions
SET "version" = 2
WHERE type_id = 'schedule';

LOCK schedule_rules IN EXCLUSIVE MODE;

UPDATE schedule_rules
SET
    end_time = cast(date_trunc('minute', end_time)+'1 minute'::interval as time without time zone),
    start_time = date_trunc('minute', start_time)
;

UPDATE schedule_rules
SET end_time = cast(end_time-'1 minute'::interval as time without time zone)
WHERE date_part('minute', end_time)::integer % 5 = 1;

-- +migrate Down
SELECT 1
FROM engine_processing_versions
WHERE type_id = 'schedule'
FOR UPDATE;

LOCK schedule_rules IN EXCLUSIVE MODE;

UPDATE schedule_rules
SET
    end_time = cast(date_trunc('minute', end_time)-'1 minute'::interval as time without time zone)
;

UPDATE engine_processing_versions
SET "version" = 1
WHERE type_id = 'schedule';
