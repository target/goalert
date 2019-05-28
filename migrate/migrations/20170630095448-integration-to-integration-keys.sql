
-- +migrate Up

CREATE TYPE enum_integration_keys_type as ENUM (
  'grafana'
);


CREATE TABLE integration_keys (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  type enum_integration_keys_type NOT NULL,
  service_id TEXT NOT NULL REFERENCES service(id) ON DELETE CASCADE,
  UNIQUE (name, service_id)
);


INSERT INTO integration_keys(name, type, service_id)
SELECT name, 'grafana', service_id
FROM integration;


-- +migrate Down


DROP TABLE IF EXISTS  integration_keys;
DROP TYPE IF EXISTS enum_integration_keys_source;
