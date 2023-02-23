-- +migrate Up
CREATE TABLE switchover_log (
    id BIGINT PRIMARY KEY,
    TIMESTAMP TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    DATA jsonb NOT NULL
);

ALTER TABLE switchover_state
ADD column db_id UUID NOT NULL DEFAULT gen_random_uuid();

-- +migrate Down
DROP TABLE switchover_log;

ALTER TABLE switchover_state DROP column db_id;
