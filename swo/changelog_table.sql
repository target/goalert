CREATE TABLE change_log (
    id BIGSERIAL PRIMARY KEY,
    op TEXT NOT NULL,
    table_name TEXT NOT NULL,
    row_id TEXT NOT NULL,
    tx_id BIGINT,
    cmd_id cid,
    row_data JSONB
)
