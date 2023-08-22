
-- +migrate Up

LOCK alerts;

ALTER TABLE alerts
    ADD COLUMN dedup_key TEXT;

UPDATE alerts
SET dedup_key =
    concat(
        'auto:1:',
        encode(digest(concat("description"), 'sha512'), 'hex')
    );

ALTER TABLE alerts
    ALTER COLUMN dedup_key SET NOT NULL;

CREATE INDEX idx_dedup_alerts ON alerts (dedup_key);

-- +migrate StatementBegin
CREATE FUNCTION fn_ensure_alert_dedup_key() RETURNS TRIGGER AS
$$
BEGIN
    IF NEW.dedup_key ISNULL THEN
        NEW.dedup_key = 
            concat(
                'auto:1:',
                encode(digest(concat(NEW."description"), 'sha512'), 'hex')
            );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

CREATE TRIGGER trg_ensure_alert_dedup_key BEFORE INSERT ON alerts
FOR EACH ROW EXECUTE PROCEDURE fn_ensure_alert_dedup_key();

-- +migrate Down

DROP TRIGGER trg_ensure_alert_dedup_key ON alerts;
DROP FUNCTION fn_ensure_alert_dedup_key();

ALTER TABLE alerts
    DROP COLUMN dedup_key; -- drops the dependent index
