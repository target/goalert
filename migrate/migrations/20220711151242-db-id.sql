-- +migrate Up
ALTER TABLE switchover_state
ADD column db_id UUID NOT NULL DEFAULT gen_random_uuid();

-- +migrate Down
ALTER TABLE switchover_state DROP column db_id;
