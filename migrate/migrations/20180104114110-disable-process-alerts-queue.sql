
-- +migrate Up

LOCK process_alerts;

-- +migrate StatementBegin
CREATE FUNCTION fn_disable_inserts() RETURNS TRIGGER AS
$$
BEGIN
    RAISE EXCEPTION 'inserts are disabled on this table';
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

-- prevent new stuff from entering the queue
CREATE TRIGGER trg_disable_old_alert_processing
BEFORE INSERT ON process_alerts
EXECUTE PROCEDURE fn_disable_inserts();


-- remove unclaimed stuff
DELETE FROM process_alerts
WHERE
    client_id ISNULL OR
    deadline ISNULL OR
    deadline <= now();

-- +migrate Down

DROP TRIGGER trg_disable_old_alert_processing ON process_alerts;
DROP FUNCTION fn_disable_inserts();
