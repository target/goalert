CREATE
OR REPLACE FUNCTION fn_process_change_log() RETURNS TRIGGER AS $$
DECLARE cur_state enum_switchover_state := 'idle';

BEGIN
SELECT INTO cur_state current_state
FROM switchover_state;

IF cur_state != 'in_progress' THEN RETURN NEW;

END IF;

IF (TG_OP = 'DELETE') THEN
INSERT INTO change_log (op, table_name, row_id, old_hash)
VALUES (
        TG_OP,
        TG_TABLE_NAME,
        cast(OLD .id AS TEXT),
        sha256(OLD::TEXT::BYTEA)
    );

RETURN OLD;

ELSIF (TG_OP = 'UPDATE') THEN
INSERT INTO change_log (
        op,
        table_name,
        row_id,
        row_data,
        new_hash,
        old_hash
    )
VALUES (
        TG_OP,
        TG_TABLE_NAME,
        cast(NEW .id AS TEXT),
        to_jsonb(NEW),
        sha256(NEW::TEXT::BYTEA),
        sha256(OLD::TEXT::BYTEA)
    );

RETURN NEW;

ELSIF (TG_OP = 'INSERT') THEN
INSERT INTO change_log (op, table_name, row_id, row_data, new_hash)
VALUES (
        TG_OP,
        TG_TABLE_NAME,
        cast(NEW .id AS TEXT),
        to_jsonb(NEW),
        sha256(NEW::TEXT::BYTEA)
    );

RETURN NEW;

ELSE RAISE
EXCEPTION 'Unexpected operation in switchover mode: %',
    TG_OP;

END IF;

RETURN NULL;

END;

$$ LANGUAGE 'plpgsql'
