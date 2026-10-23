-- +migrate Up notransaction
-- schedule_rules is only indexed with schedule_id as the leading column
-- (idx_rule_schedule, idx_target_schedule), so looking up which schedules
-- reference a rotation (a WHERE on tgt_rotation_id alone) has to scan the whole
-- table. That lookup is needed by the upcoming rotation_state and
-- rotation_participants -> schedule fan-out triggers (a rotation change must
-- re-evaluate every schedule that targets it) and by the ON DELETE CASCADE path
-- of schedule_rules_tgt_rotation_id_fkey.
--
-- Partial so the index only holds rows that actually target a rotation. Rules
-- that target a user directly have a NULL tgt_rotation_id and are never looked
-- up this way.
--
-- CONCURRENTLY so the build does not block writes to schedule_rules on a live
-- system. It must run outside a transaction, hence notransaction. A failed
-- CONCURRENTLY build leaves an INVALID index behind, so the Up drops any
-- existing index first and a re-run rebuilds it from scratch (same pattern as
-- 20180831132707-alerts-service-index.sql).
--
-- Comments in a notransaction migration must not contain a semicolon or a
-- dollar sign (the runner splits statements on those without stripping
-- comments).
DROP INDEX CONCURRENTLY IF EXISTS idx_schedule_rules_tgt_rotation;

CREATE INDEX CONCURRENTLY idx_schedule_rules_tgt_rotation ON schedule_rules (tgt_rotation_id)
WHERE
    tgt_rotation_id IS NOT NULL;

-- +migrate Down notransaction
DROP INDEX CONCURRENTLY IF EXISTS idx_schedule_rules_tgt_rotation;
