
-- +migrate Up

CREATE UNIQUE INDEX escalation_policies_name ON escalation_policies (lower("name"));
CREATE UNIQUE INDEX services_name ON services (lower("name"));
CREATE UNIQUE INDEX schedules_name ON schedules (lower("name"));
CREATE UNIQUE INDEX rotations_name ON rotations (lower("name"));
CREATE UNIQUE INDEX integration_keys_name_service_id ON integration_keys (lower("name"), service_id);

-- +migrate Down

DROP INDEX 
    escalation_policies_name,
    services_name,
    schedules_name,
    rotations_name,
    integration_keys_name_service_id;
