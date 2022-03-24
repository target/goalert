CREATE UNLOGGED TABLE change_log (
    id BIGSERIAL PRIMARY KEY,
    table_name TEXT NOT NULL,
    row_id TEXT NOT NULL
);

ALTER TABLE change_log
SET (
        autovacuum_enabled = FALSE,
        toast.autovacuum_enabled = FALSE
    )
