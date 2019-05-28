
-- +migrate Up

CREATE TYPE enum_heartbeat_state AS ENUM (
    'inactive',
    'healthy',
    'unhealthy'
);

CREATE TABLE heartbeat_monitors (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    service_id UUID NOT NULL REFERENCES services(id),
    heartbeat_interval INTERVAL NOT NULL,
    last_state enum_heartbeat_state NOT NULL DEFAULT 'inactive',
    last_heartbeat TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX heartbeat_monitor_name_service_id ON heartbeat_monitors (lower("name"), service_id);

-- +migrate Down

DROP TABLE heartbeat_monitors;
DROP TYPE enum_heartbeat_state;
