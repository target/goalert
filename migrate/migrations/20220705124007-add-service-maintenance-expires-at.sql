-- +migrate Up
ALTER TABLE services ADD COLUMN maintenance_expires_at TIMESTAMPTZ;

-- +migrate Down
ALTER TABLE services DROP COLUMN maintenance_expires_at;
