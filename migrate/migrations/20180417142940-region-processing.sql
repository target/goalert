
-- +migrate Up
CREATE TABLE region_ids (
    name TEXT PRIMARY KEY,
    id SERIAL UNIQUE

);
-- +migrate Down
DROP TABLE region_ids;