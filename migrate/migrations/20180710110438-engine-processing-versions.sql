
-- +migrate Up

CREATE TYPE engine_processing_type AS ENUM (
    'escalation',
    'heartbeat',
    'np_cycle',
    'rotation',
    'schedule',
    'status_update',
    'verify',
    'message'
);

CREATE TABLE engine_processing_versions (
    type_id engine_processing_type PRIMARY KEY,
    version INT NOT NULL DEFAULT 1
);

INSERT INTO engine_processing_versions (type_id)
VALUES
    ('escalation'),
    ('heartbeat'),
    ('np_cycle'),
    ('rotation'),
    ('schedule'),
    ('status_update'),
    ('verify'),
    ('message');

-- +migrate Down

DROP TABLE engine_processing_versions;
DROP TYPE engine_processing_type;
