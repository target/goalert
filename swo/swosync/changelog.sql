CREATE UNLOGGED TABLE change_log(
    id bigserial PRIMARY KEY,
    table_name text NOT NULL,
    row_id text NOT NULL
);

ALTER TABLE change_log SET (autovacuum_enabled = FALSE, toast.autovacuum_enabled = FALSE);

CREATE OR REPLACE FUNCTION fn_process_change_log()
    RETURNS TRIGGER
    AS $$
DECLARE
    cur_state enum_switchover_state := 'idle';
BEGIN
    SELECT
        INTO cur_state current_state
    FROM
        switchover_state;
    IF cur_state != 'in_progress' THEN
        RETURN NEW;
    END IF;
    IF (TG_OP = 'DELETE') THEN
        INSERT INTO change_log(table_name, row_id)
            VALUES (TG_TABLE_NAME, cast(OLD.id AS text));
        RETURN OLD;
    ELSE
        INSERT INTO change_log(table_name, row_id)
            VALUES (TG_TABLE_NAME, cast(NEW.id AS text));
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$
LANGUAGE 'plpgsql';

