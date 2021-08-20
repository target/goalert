-- +migrate Up notransaction
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sched_oncall_times ON schedule_on_call_users
USING spgist (tstzrange(start_time, end_time));

-- +migrate Down

DROP INDEX IF EXISTS idx_sched_oncall_times;
