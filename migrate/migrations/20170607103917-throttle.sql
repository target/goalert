
-- +migrate Up

CREATE TYPE enum_throttle_type as ENUM (
    'notifications'
);

CREATE TABLE throttle (
    action enum_throttle_type PRIMARY KEY,
    client_id UUID,
    last_action_time TIMESTAMP NOT NULL DEFAULT now()
);

INSERT INTO throttle (action) VALUES ('notifications');

-- +migrate Down

DROP TABLE throttle;
DROP TYPE enum_throttle_type;
