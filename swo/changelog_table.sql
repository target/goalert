CREATE UNLOGGED TABLE change_log (
    id BIGSERIAL PRIMARY KEY,
    op TEXT NOT NULL,
    table_name TEXT NOT NULL,
    row_id TEXT NOT NULL,
    row_data JSONB DEFAULT '{}',
    old_hash bytea,
    new_hash bytea
);

ALTER TABLE change_log
SET (
        autovacuum_enabled = FALSE,
        toast.autovacuum_enabled = FALSE
    )
