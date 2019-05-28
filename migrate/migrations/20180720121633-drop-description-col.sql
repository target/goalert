
-- +migrate Up
DROP TRIGGER trg_00_ensure_description ON public.alerts;
DROP FUNCTION fn_alerts_ensure_description();

ALTER TABLE alerts
    DROP COLUMN "description";

-- +migrate Down

ALTER TABLE alerts
    ADD COLUMN "description" text;

UPDATE alerts
SET "description" = TRIM(TRAILING chr(10) FROM summary || chr(10) || details);

ALTER TABLE alerts
    ALTER COLUMN "description" SET NOT NULL;

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_alerts_ensure_description() RETURNS TRIGGER AS
$$
BEGIN
    IF NEW.description ISNULL THEN
        NEW.description = TRIM(TRAILING chr(10) FROM NEW.summary || chr(10) || NEW.details);
    END IF;

    IF NEW.summary ISNULL THEN
        NEW.summary = split_part(NEW.description, chr(10), 1);
        NEW.details = (
            case when strpos(NEW.description, chr(10)) > 0 then
                substr(NEW.description, strpos(NEW.description, chr(10))+1)
            else
                ''
            end
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

CREATE TRIGGER trg_00_ensure_description
BEFORE INSERT ON public.alerts
FOR EACH ROW
EXECUTE PROCEDURE fn_alerts_ensure_description();
