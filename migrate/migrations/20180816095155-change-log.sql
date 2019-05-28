
-- +migrate Up notransaction

-- +migrate StatementBegin
BEGIN;

DROP TABLE IF EXISTS change_log;
CREATE TABLE IF NOT EXISTS change_log (
    id BIGSERIAL PRIMARY KEY,
    op TEXT NOT NULL,
    table_name TEXT NOT NULL,
    row_id TEXT NOT NULL,
    tx_id BIGINT,
    cmd_id cid,
    row_data JSONB
);


CREATE OR REPLACE FUNCTION process_change() RETURNS TRIGGER AS $$
DECLARE
    cur_state enum_switchover_state := 'idle';
BEGIN
    SELECT INTO cur_state current_state
    FROM switchover_state;
    
    IF cur_state != 'in_progress' THEN
        RETURN NEW;
    END IF;

    IF (TG_OP = 'DELETE') THEN
        INSERT INTO change_log (op, table_name, row_id, tx_id, cmd_id)
        VALUES (TG_OP, TG_TABLE_NAME, cast(OLD.id as TEXT), txid_current(), OLD.cmax);
        RETURN OLD;
    ELSE
        INSERT INTO change_log (op, table_name, row_id, tx_id, cmd_id, row_data)
        VALUES (TG_OP, TG_TABLE_NAME, cast(NEW.id as TEXT), txid_current(), NEW.cmin, to_jsonb(NEW));
        RETURN NEW;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE 'plpgsql';
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_alert_logs_change_log ON alert_logs;
CREATE TRIGGER zz_99_alert_logs_change_log
AFTER INSERT OR UPDATE OR DELETE ON alert_logs
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_alerts_change_log ON alerts;
CREATE TRIGGER zz_99_alerts_change_log
AFTER INSERT OR UPDATE OR DELETE ON alerts
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_auth_basic_users_change_log ON auth_basic_users;
CREATE TRIGGER zz_99_auth_basic_users_change_log
AFTER INSERT OR UPDATE OR DELETE ON auth_basic_users
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_auth_nonce_change_log ON auth_nonce;
CREATE TRIGGER zz_99_auth_nonce_change_log
AFTER INSERT OR UPDATE OR DELETE ON auth_nonce
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_auth_subjects_change_log ON auth_subjects;
CREATE TRIGGER zz_99_auth_subjects_change_log
AFTER INSERT OR UPDATE OR DELETE ON auth_subjects
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_auth_user_sessions_change_log ON auth_user_sessions;
CREATE TRIGGER zz_99_auth_user_sessions_change_log
AFTER INSERT OR UPDATE OR DELETE ON auth_user_sessions
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_config_limits_change_log ON config_limits;
CREATE TRIGGER zz_99_config_limits_change_log
AFTER INSERT OR UPDATE OR DELETE ON config_limits
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_ep_step_on_call_users_change_log ON ep_step_on_call_users;
CREATE TRIGGER zz_99_ep_step_on_call_users_change_log
AFTER INSERT OR UPDATE OR DELETE ON ep_step_on_call_users
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_escalation_policies_change_log ON escalation_policies;
CREATE TRIGGER zz_99_escalation_policies_change_log
AFTER INSERT OR UPDATE OR DELETE ON escalation_policies
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_escalation_policy_actions_change_log ON escalation_policy_actions;
CREATE TRIGGER zz_99_escalation_policy_actions_change_log
AFTER INSERT OR UPDATE OR DELETE ON escalation_policy_actions
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_escalation_policy_state_change_log ON escalation_policy_state;
CREATE TRIGGER zz_99_escalation_policy_state_change_log
AFTER INSERT OR UPDATE OR DELETE ON escalation_policy_state
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_escalation_policy_steps_change_log ON escalation_policy_steps;
CREATE TRIGGER zz_99_escalation_policy_steps_change_log
AFTER INSERT OR UPDATE OR DELETE ON escalation_policy_steps
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_heartbeat_monitors_change_log ON heartbeat_monitors;
CREATE TRIGGER zz_99_heartbeat_monitors_change_log
AFTER INSERT OR UPDATE OR DELETE ON heartbeat_monitors
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_integration_keys_change_log ON integration_keys;
CREATE TRIGGER zz_99_integration_keys_change_log
AFTER INSERT OR UPDATE OR DELETE ON integration_keys
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_keyring_change_log ON keyring;
CREATE TRIGGER zz_99_keyring_change_log
AFTER INSERT OR UPDATE OR DELETE ON keyring
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_notification_policy_cycles_change_log ON notification_policy_cycles;
CREATE TRIGGER zz_99_notification_policy_cycles_change_log
AFTER INSERT OR UPDATE OR DELETE ON notification_policy_cycles
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_outgoing_messages_change_log ON outgoing_messages;
CREATE TRIGGER zz_99_outgoing_messages_change_log
AFTER INSERT OR UPDATE OR DELETE ON outgoing_messages
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_region_ids_change_log ON region_ids;
CREATE TRIGGER zz_99_region_ids_change_log
AFTER INSERT OR UPDATE OR DELETE ON region_ids
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_rotation_participants_change_log ON rotation_participants;
CREATE TRIGGER zz_99_rotation_participants_change_log
AFTER INSERT OR UPDATE OR DELETE ON rotation_participants
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_rotation_state_change_log ON rotation_state;
CREATE TRIGGER zz_99_rotation_state_change_log
AFTER INSERT OR UPDATE OR DELETE ON rotation_state
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_rotations_change_log ON rotations;
CREATE TRIGGER zz_99_rotations_change_log
AFTER INSERT OR UPDATE OR DELETE ON rotations
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_schedule_on_call_users_change_log ON schedule_on_call_users;
CREATE TRIGGER zz_99_schedule_on_call_users_change_log
AFTER INSERT OR UPDATE OR DELETE ON schedule_on_call_users
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_schedule_rules_change_log ON schedule_rules;
CREATE TRIGGER zz_99_schedule_rules_change_log
AFTER INSERT OR UPDATE OR DELETE ON schedule_rules
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_schedules_change_log ON schedules;
CREATE TRIGGER zz_99_schedules_change_log
AFTER INSERT OR UPDATE OR DELETE ON schedules
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_services_change_log ON services;
CREATE TRIGGER zz_99_services_change_log
AFTER INSERT OR UPDATE OR DELETE ON services
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_twilio_sms_callbacks_change_log ON twilio_sms_callbacks;
CREATE TRIGGER zz_99_twilio_sms_callbacks_change_log
AFTER INSERT OR UPDATE OR DELETE ON twilio_sms_callbacks
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_twilio_sms_errors_change_log ON twilio_sms_errors;
CREATE TRIGGER zz_99_twilio_sms_errors_change_log
AFTER INSERT OR UPDATE OR DELETE ON twilio_sms_errors
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_twilio_voice_errors_change_log ON twilio_voice_errors;
CREATE TRIGGER zz_99_twilio_voice_errors_change_log
AFTER INSERT OR UPDATE OR DELETE ON twilio_voice_errors
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_user_contact_methods_change_log ON user_contact_methods;
CREATE TRIGGER zz_99_user_contact_methods_change_log
AFTER INSERT OR UPDATE OR DELETE ON user_contact_methods
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_user_favorites_change_log ON user_favorites;
CREATE TRIGGER zz_99_user_favorites_change_log
AFTER INSERT OR UPDATE OR DELETE ON user_favorites
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_user_last_alert_log_change_log ON user_last_alert_log;
CREATE TRIGGER zz_99_user_last_alert_log_change_log
AFTER INSERT OR UPDATE OR DELETE ON user_last_alert_log
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_user_notification_rules_change_log ON user_notification_rules;
CREATE TRIGGER zz_99_user_notification_rules_change_log
AFTER INSERT OR UPDATE OR DELETE ON user_notification_rules
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_user_overrides_change_log ON user_overrides;
CREATE TRIGGER zz_99_user_overrides_change_log
AFTER INSERT OR UPDATE OR DELETE ON user_overrides
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_user_verification_codes_change_log ON user_verification_codes;
CREATE TRIGGER zz_99_user_verification_codes_change_log
AFTER INSERT OR UPDATE OR DELETE ON user_verification_codes
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_users_change_log ON users;
CREATE TRIGGER zz_99_users_change_log
AFTER INSERT OR UPDATE OR DELETE ON users
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate Down notransaction
DROP TRIGGER IF EXISTS zz_99_alert_logs_change_log ON alert_logs;
DROP TRIGGER IF EXISTS zz_99_alerts_change_log ON alerts;
DROP TRIGGER IF EXISTS zz_99_auth_basic_users_change_log ON auth_basic_users;
DROP TRIGGER IF EXISTS zz_99_auth_nonce_change_log ON auth_nonce;
DROP TRIGGER IF EXISTS zz_99_auth_subjects_change_log ON auth_subjects;
DROP TRIGGER IF EXISTS zz_99_auth_user_sessions_change_log ON auth_user_sessions;
DROP TRIGGER IF EXISTS zz_99_config_limits_change_log ON config_limits;
DROP TRIGGER IF EXISTS zz_99_ep_step_on_call_users_change_log ON ep_step_on_call_users;
DROP TRIGGER IF EXISTS zz_99_escalation_policies_change_log ON escalation_policies;
DROP TRIGGER IF EXISTS zz_99_escalation_policy_actions_change_log ON escalation_policy_actions;
DROP TRIGGER IF EXISTS zz_99_escalation_policy_state_change_log ON escalation_policy_state;
DROP TRIGGER IF EXISTS zz_99_escalation_policy_steps_change_log ON escalation_policy_steps;
DROP TRIGGER IF EXISTS zz_99_heartbeat_monitors_change_log ON heartbeat_monitors;
DROP TRIGGER IF EXISTS zz_99_integration_keys_change_log ON integration_keys;
DROP TRIGGER IF EXISTS zz_99_keyring_change_log ON keyring;
DROP TRIGGER IF EXISTS zz_99_notification_policy_cycles_change_log ON notification_policy_cycles;
DROP TRIGGER IF EXISTS zz_99_outgoing_messages_change_log ON outgoing_messages;
DROP TRIGGER IF EXISTS zz_99_region_ids_change_log ON region_ids;
DROP TRIGGER IF EXISTS zz_99_rotation_participants_change_log ON rotation_participants;
DROP TRIGGER IF EXISTS zz_99_rotation_state_change_log ON rotation_state;
DROP TRIGGER IF EXISTS zz_99_rotations_change_log ON rotations;
DROP TRIGGER IF EXISTS zz_99_schedule_on_call_users_change_log ON schedule_on_call_users;
DROP TRIGGER IF EXISTS zz_99_schedule_rules_change_log ON schedule_rules;
DROP TRIGGER IF EXISTS zz_99_schedules_change_log ON schedules;
DROP TRIGGER IF EXISTS zz_99_services_change_log ON services;
DROP TRIGGER IF EXISTS zz_99_twilio_sms_callbacks_change_log ON twilio_sms_callbacks;
DROP TRIGGER IF EXISTS zz_99_twilio_sms_errors_change_log ON twilio_sms_errors;
DROP TRIGGER IF EXISTS zz_99_twilio_voice_errors_change_log ON twilio_voice_errors;
DROP TRIGGER IF EXISTS zz_99_user_contact_methods_change_log ON user_contact_methods;
DROP TRIGGER IF EXISTS zz_99_user_favorites_change_log ON user_favorites;
DROP TRIGGER IF EXISTS zz_99_user_last_alert_log_change_log ON user_last_alert_log;
DROP TRIGGER IF EXISTS zz_99_user_notification_rules_change_log ON user_notification_rules;
DROP TRIGGER IF EXISTS zz_99_user_overrides_change_log ON user_overrides;
DROP TRIGGER IF EXISTS zz_99_user_verification_codes_change_log ON user_verification_codes;
DROP TRIGGER IF EXISTS zz_99_users_change_log ON users;

DROP FUNCTION IF EXISTS process_change();
DROP TABLE IF EXISTS change_log;
