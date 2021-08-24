-- +migrate Up
CREATE INDEX idx_valid_contact_methods ON user_contact_methods (id) WHERE NOT disabled;
CREATE INDEX idx_user_status_updates ON users (alert_status_log_contact_method_id) WHERE alert_status_log_contact_method_id IS NOT NULL;

-- +migrate Down

DROP INDEX idx_valid_contact_methods;
DROP INDEX idx_user_status_updates;
