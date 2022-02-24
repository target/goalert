-- +migrate Up
ALTER TABLE schedule_data
    ADD COLUMN id BIGSERIAL UNIQUE NOT NULL;

-- +migrate Down
ALTER TABLE schedule_data
    DROP COLUMN id;
