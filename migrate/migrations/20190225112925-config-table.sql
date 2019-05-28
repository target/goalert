
-- +migrate Up
CREATE TABLE config (
    id SERIAL PRIMARY KEY,
    schema INT NOT NULL,
    data BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE TRIGGER zz_99_config_change_log
AFTER INSERT OR UPDATE OR DELETE ON config
FOR EACH ROW EXECUTE PROCEDURE process_change();

-- +migrate Down
DROP TRIGGER IF EXISTS zz_99_config_change_log ON config;
DROP TABLE config;
