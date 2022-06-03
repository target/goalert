-- name: ForeignKeys :many
SELECT src.relname::text,
    dst.relname::text
FROM pg_catalog.pg_constraint con
    JOIN pg_catalog.pg_namespace ns ON ns.nspname = 'public'
    AND ns.oid = con.connamespace
    JOIN pg_catalog.pg_class src ON src.oid = con.conrelid
    JOIN pg_catalog.pg_class dst ON dst.oid = con.confrelid
WHERE con.contype = 'f'
    AND NOT con.condeferrable;

-- name: Changes :many
SELECT id,
    table_name,
    row_id
FROM change_log;

-- name: DeleteChanges :exec
DELETE FROM change_log
WHERE id = ANY($1);

-- name: CurrentTime :one
SELECT now()::timestamptz;

-- name: ActiveTxCount :one
SELECT COUNT(*)
FROM pg_stat_activity
WHERE "state" <> 'idle'
    AND "xact_start" <= $1;

-- name: GlobalSwitchoverSharedConnLock :exec
SELECT pg_advisory_lock_shared(4369);

-- name: GlobalSwitchoverTxExclusiveConnLock :exec
SELECT pg_advisory_xact_lock(4369);

-- name: GlobalSwitchoverExecLock :one
SELECT pg_try_advisory_lock(4370)
FROM switchover_state
WHERE current_state != 'use_next_db';

-- name: TableColumns :many
SELECT col.table_name,
    col.column_name,
    col.data_type,
    col.ordinal_position
FROM information_schema.columns col
    JOIN information_schema.tables t ON t.table_catalog = col.table_catalog
    AND t.table_schema = col.table_schema
    AND t.table_name = col.table_name
    AND t.table_type = 'BASE TABLE'
WHERE col.table_catalog = current_database()
    AND col.table_schema = 'public';

-- name: SetIdleTimeout :exec
SET idle_in_transaction_session_timeout = 3000;

-- name: CurrentSwitchoverState :one
SELECT current_state
FROM switchover_state;

-- name: CurrentSwitchoverStateNoWait :one
SELECT current_state
FROM switchover_state NOWAIT;

-- name: UnlockAll :exec
SELECT pg_advisory_unlock_all();

-- name: LogEvents :many
SELECT id,
    TIMESTAMP,
    DATA
FROM switchover_log
WHERE id > $1
ORDER BY id ASC
LIMIT 100;

-- name: LastLogID :one
SELECT COALESCE(MAX(id), 0)::bigint
FROM switchover_log;

-- name: SequenceNames :many
SELECT sequence_name
FROM information_schema.sequences
WHERE sequence_catalog = current_database()
    AND sequence_schema = 'public'
    AND sequence_name != 'change_log_id_seq';
