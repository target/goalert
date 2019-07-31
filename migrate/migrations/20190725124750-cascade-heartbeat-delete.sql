-- +migrate Up

ALTER TABLE heartbeat_monitors
    DROP CONSTRAINT heartbeat_monitors_service_id_fkey,
    ADD CONSTRAINT heartbeat_monitors_service_id_fkey FOREIGN KEY (service_id) REFERENCES services (id) ON DELETE CASCADE;

-- +migrate Down
ALTER TABLE heartbeat_monitors
    DROP CONSTRAINT heartbeat_monitors_service_id_fkey,
    ADD CONSTRAINT heartbeat_monitors_service_id_fkey FOREIGN KEY (service_id) REFERENCES services (id);
