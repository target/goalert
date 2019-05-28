
-- +migrate Up
ALTER TABLE users
    ALTER CONSTRAINT users_alert_status_log_contact_method_id_fkey DEFERRABLE;

-- +migrate Down
ALTER TABLE users
    ALTER CONSTRAINT users_alert_status_log_contact_method_id_fkey NOT DEFERRABLE;
