-- +migrate Up

DROP FUNCTION public.process_change ();

DROP TABLE  change_log;

-- +migrate Down

CREATE TABLE change_log (
    id BIGSERIAL PRIMARY KEY,
    op TEXT NOT NULL,
    table_name TEXT NOT NULL,
    row_id TEXT NOT NULL,
    tx_id BIGINT,
    cmd_id cid,
    row_data JSONB
);

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION process_change() RETURNS TRIGGER AS $$
DECLARE
    cur_state enum_switchover_state := 'idle';
BEGIN
    SELECT INTO cur_state current_state
    FROM switchover_state;
    
    IF cur_state != 'in_progress' THEN
        RETURN NEW;
    END IF;

    IF (TG_OP = 'DELETE') THEN
        INSERT INTO change_log (op, table_name, row_id, tx_id, cmd_id)
        VALUES (TG_OP, TG_TABLE_NAME, cast(OLD.id as TEXT), txid_current(), OLD.cmax);
        RETURN OLD;
    ELSE
        INSERT INTO change_log (op, table_name, row_id, tx_id, cmd_id, row_data)
        VALUES (TG_OP, TG_TABLE_NAME, cast(NEW.id as TEXT), txid_current(), NEW.cmin, to_jsonb(NEW));
        RETURN NEW;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd
