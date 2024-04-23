-- +migrate Up
ALTER TABLE integration_keys
    ADD COLUMN external_system_name TEXT,
    DROP CONSTRAINT integration_keys_name_service_id_key;

DROP INDEX IF EXISTS integration_keys_name_service_id_key;

DROP INDEX IF EXISTS integration_keys_name_service_id;

CREATE UNIQUE INDEX idx_int_key_name_svc_ext ON public.integration_keys USING btree(lower(name), service_id, coalesce(external_system_name, ''));

-- +migrate Down
LOCK TABLE integration_keys;

DELETE FROM integration_keys
WHERE external_system_name IS NOT NULL;

ALTER TABLE integration_keys
    DROP COLUMN external_system_name,
    ADD CONSTRAINT integration_keys_name_service_id_key UNIQUE (name, service_id);

CREATE UNIQUE INDEX integration_keys_name_service_id ON public.integration_keys USING btree(lower(name), service_id);

