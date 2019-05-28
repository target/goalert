
-- +migrate Up
DROP TABLE alert, incident, incident_assignment, incident_log, maintenance, service_maintenance;

-- +migrate Down

SELECT 1;

