
-- +migrate Up
CREATE TYPE enum_switchover_state as ENUM (
    'idle',
    'in_progress',
    'use_next_db'
);

CREATE TABLE switchover_state (
    ok BOOL PRIMARY KEY,
    current_state enum_switchover_state NOT NULL,
    CHECK(ok)
);

INSERT INTO switchover_state (ok, current_state)
VALUES (true, 'idle');

-- +migrate Down

DROP TABLE switchover_state;
DROP TYPE enum_switchover_state;