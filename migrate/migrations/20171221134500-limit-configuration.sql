
-- +migrate Up
CREATE TYPE enum_limit_type AS ENUM (
    'notification_rules_per_user',
    'contact_methods_per_user',
    'ep_steps_per_policy',
    'ep_actions_per_step',
    'participants_per_rotation',
    'rules_per_schedule',
    'integration_keys_per_service',
    'unacked_alerts_per_service',
    'targets_per_schedule'
);
CREATE TABLE config_limits (
    id enum_limit_type PRIMARY KEY,
    max INT NOT NULL DEFAULT -1
);

-- +migrate Down

DROP TABLE config_limits;
DROP TYPE enum_limit_type;
