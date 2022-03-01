-- +migrate Up
CREATE TABLE switchover_log (
    id BIGINT PRIMARY KEY,
    timestamp timestamp with time zone NOT NULL DEFAULT now(),
    data jsonb NOT NULL
);

-- +migrate Down
DROP TABLE switchover_log;
