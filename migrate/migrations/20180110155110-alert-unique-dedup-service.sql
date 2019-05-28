
-- +migrate Up

ALTER TABLE alerts
    ALTER COLUMN dedup_key DROP NOT NULL;

UPDATE alerts
SET dedup_key = NULL
WHERE status = 'closed';

ALTER TABLE alerts
    ADD CONSTRAINT dedup_key_only_for_open_alerts CHECK((status = 'closed') = (dedup_key isnull));

CREATE UNIQUE INDEX idx_no_alert_duplicates ON alerts (service_id, dedup_key);

-- +migrate StatementBegin
CREATE FUNCTION fn_clear_dedup_on_close() RETURNS trigger AS $$
BEGIN
    NEW.dedup_key = NULL;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

DROP TRIGGER trg_ensure_alert_dedup_key ON ALERTS;
CREATE TRIGGER trg_ensure_alert_dedup_key BEFORE INSERT ON alerts
FOR EACH ROW
WHEN (NEW.status != 'closed')
EXECUTE PROCEDURE fn_ensure_alert_dedup_key();

CREATE TRIGGER trg_clear_dedup_on_close BEFORE UPDATE ON alerts
FOR EACH ROW
WHEN (NEW.status != OLD.status AND NEW.status = 'closed')
EXECUTE PROCEDURE fn_clear_dedup_on_close();

-- +migrate Down

DROP INDEX idx_no_alert_duplicates;
ALTER TABLE alerts
    DROP CONSTRAINT dedup_key_only_for_open_alerts;

UPDATE alerts
SET dedup_key = concat(
                'auto:1:',
                encode(digest(concat("description"), 'sha512'), 'hex')
            )
WHERE dedup_key ISNULL;

ALTER TABLE alerts
    ALTER COLUMN dedup_key SET NOT NULL;

DROP TRIGGER trg_clear_dedup_on_close ON alerts;
DROP FUNCTION fn_clear_dedup_on_close();

DROP TRIGGER trg_ensure_alert_dedup_key ON ALERTS;
CREATE TRIGGER trg_ensure_alert_dedup_key BEFORE INSERT ON alerts
FOR EACH ROW EXECUTE PROCEDURE fn_ensure_alert_dedup_key();
