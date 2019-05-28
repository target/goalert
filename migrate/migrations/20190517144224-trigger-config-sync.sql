-- +migrate Up

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_notify_config_refresh() RETURNS TRIGGER AS
    $$
    BEGIN
        NOTIFY "/goalert/config-refresh";
        RETURN NEW;
    END;
    $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

CREATE TRIGGER trg_config_update
    AFTER INSERT ON config
    FOR EACH ROW
    EXECUTE PROCEDURE fn_notify_config_refresh();

-- +migrate Down

DROP TRIGGER trg_config_update ON config;
DROP FUNCTION fn_notify_config_refresh();
