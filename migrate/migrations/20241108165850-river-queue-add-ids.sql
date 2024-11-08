-- +migrate Up
ALTER TABLE river_client_queue
    ADD COLUMN id BIGSERIAL UNIQUE;

ALTER TABLE river_leader
    ADD COLUMN id BIGSERIAL UNIQUE;

ALTER TABLE river_queue
    ADD COLUMN id BIGSERIAL UNIQUE;

-- +migrate Down
ALTER TABLE river_client_queue
    DROP COLUMN id;

ALTER TABLE river_leader
    DROP COLUMN id;

ALTER TABLE river_queue
    DROP COLUMN id;

