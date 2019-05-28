
-- +migrate Up
ALTER TABLE integration
DROP CONSTRAINT integration_service_id_fkey,
ADD CONSTRAINT integration_service_id_fkey
	FOREIGN KEY (service_id)
	REFERENCES service(id)
	ON DELETE CASCADE;

-- +migrate Down
ALTER TABLE integration
DROP CONSTRAINT integration_service_id_fkey,
ADD CONSTRAINT integration_service_id_fkey
	FOREIGN KEY (service_id)
	REFERENCES service(id);
