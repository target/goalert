
-- +migrate Up

CREATE INDEX idx_integration_key_service ON integration_keys (service_id);

-- +migrate StatementBegin
CREATE FUNCTION fn_enforce_integration_key_limit() RETURNS trigger AS $$
DECLARE
    max_count INT := -1;
    val_count INT := 0;
BEGIN
    SELECT INTO max_count max
    FROM config_limits
    WHERE id = 'integration_keys_per_service';

    IF max_count = -1 THEN
        RETURN NEW;
    END IF;

    SELECT INTO val_count COUNT(*)
    FROM integration_keys
    WHERE service_id = NEW.service_id;

    IF val_count > max_count THEN
        RAISE 'limit exceeded' USING ERRCODE='check_violation', CONSTRAINT='integration_keys_per_service_limit', HINT='max='||max_count;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd


CREATE CONSTRAINT TRIGGER trg_enforce_integration_key_limit 
    AFTER INSERT ON integration_keys
    FOR EACH ROW EXECUTE PROCEDURE fn_enforce_integration_key_limit();

-- +migrate Down

DROP TRIGGER trg_enforce_integration_key_limit ON integration_keys;
DROP FUNCTION fn_enforce_integration_key_limit();
DROP INDEX idx_integration_key_service;
