-- +migrate Up notransaction

DROP TRIGGER IF EXISTS zz_99_alert_logs_change_log ON public.alert_logs;
DROP TRIGGER IF EXISTS zz_99_alerts_change_log ON public.alerts;
DROP TRIGGER IF EXISTS zz_99_auth_basic_users_change_log ON public.auth_basic_users;
DROP TRIGGER IF EXISTS zz_99_auth_nonce_change_log ON public.auth_nonce;
DROP TRIGGER IF EXISTS zz_99_auth_subjects_change_log ON public.auth_subjects;
DROP TRIGGER IF EXISTS zz_99_auth_user_sessions_change_log ON public.auth_user_sessions;
DROP TRIGGER IF EXISTS zz_99_config_change_log ON public.config;
DROP TRIGGER IF EXISTS zz_99_config_limits_change_log ON public.config_limits;
DROP TRIGGER IF EXISTS zz_99_ep_step_on_call_users_change_log ON public.ep_step_on_call_users;
DROP TRIGGER IF EXISTS zz_99_escalation_policies_change_log ON public.escalation_policies;
DROP TRIGGER IF EXISTS zz_99_escalation_policy_actions_change_log ON public.escalation_policy_actions;
DROP TRIGGER IF EXISTS zz_99_escalation_policy_state_change_log ON public.escalation_policy_state;
DROP TRIGGER IF EXISTS zz_99_escalation_policy_steps_change_log ON public.escalation_policy_steps;
DROP TRIGGER IF EXISTS zz_99_heartbeat_monitors_change_log ON public.heartbeat_monitors;
DROP TRIGGER IF EXISTS zz_99_integration_keys_change_log ON public.integration_keys;
DROP TRIGGER IF EXISTS zz_99_keyring_change_log ON public.keyring;
DROP TRIGGER IF EXISTS zz_99_labels_change_log ON public.labels;
DROP TRIGGER IF EXISTS zz_99_notification_policy_cycles_change_log ON public.notification_policy_cycles;
DROP TRIGGER IF EXISTS zz_99_outgoing_messages_change_log ON public.outgoing_messages;
DROP TRIGGER IF EXISTS zz_99_region_ids_change_log ON public.region_ids;
DROP TRIGGER IF EXISTS zz_99_rotation_participants_change_log ON public.rotation_participants;
DROP TRIGGER IF EXISTS zz_99_rotation_state_change_log ON public.rotation_state;
DROP TRIGGER IF EXISTS zz_99_rotations_change_log ON public.rotations;
DROP TRIGGER IF EXISTS zz_99_schedule_on_call_users_change_log ON public.schedule_on_call_users;
DROP TRIGGER IF EXISTS zz_99_schedule_rules_change_log ON public.schedule_rules;
DROP TRIGGER IF EXISTS zz_99_schedules_change_log ON public.schedules;
DROP TRIGGER IF EXISTS zz_99_services_change_log ON public.services;
DROP TRIGGER IF EXISTS zz_99_twilio_sms_callbacks_change_log ON public.twilio_sms_callbacks;
DROP TRIGGER IF EXISTS zz_99_twilio_sms_errors_change_log ON public.twilio_sms_errors;
DROP TRIGGER IF EXISTS zz_99_twilio_voice_errors_change_log ON public.twilio_voice_errors;
DROP TRIGGER IF EXISTS zz_99_user_contact_methods_change_log ON public.user_contact_methods;
DROP TRIGGER IF EXISTS zz_99_user_favorites_change_log ON public.user_favorites;
DROP TRIGGER IF EXISTS zz_99_user_last_alert_log_change_log ON public.user_last_alert_log;
DROP TRIGGER IF EXISTS zz_99_user_notification_rules_change_log ON public.user_notification_rules;
DROP TRIGGER IF EXISTS zz_99_user_overrides_change_log ON public.user_overrides;
DROP TRIGGER IF EXISTS zz_99_user_verification_codes_change_log ON public.user_verification_codes;
DROP TRIGGER IF EXISTS zz_99_users_change_log ON public.users;


-- +migrate Down notransaction

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_alert_logs_change_log AFTER INSERT OR DELETE OR UPDATE ON public.alert_logs FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_alerts_change_log AFTER INSERT OR DELETE OR UPDATE ON public.alerts FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_auth_basic_users_change_log AFTER INSERT OR DELETE OR UPDATE ON public.auth_basic_users FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_auth_nonce_change_log AFTER INSERT OR DELETE OR UPDATE ON public.auth_nonce FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_auth_subjects_change_log AFTER INSERT OR DELETE OR UPDATE ON public.auth_subjects FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_auth_user_sessions_change_log AFTER INSERT OR DELETE OR UPDATE ON public.auth_user_sessions FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_config_change_log AFTER INSERT OR DELETE OR UPDATE ON public.config FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_config_limits_change_log AFTER INSERT OR DELETE OR UPDATE ON public.config_limits FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_ep_step_on_call_users_change_log AFTER INSERT OR DELETE OR UPDATE ON public.ep_step_on_call_users FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_escalation_policies_change_log AFTER INSERT OR DELETE OR UPDATE ON public.escalation_policies FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_escalation_policy_actions_change_log AFTER INSERT OR DELETE OR UPDATE ON public.escalation_policy_actions FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_escalation_policy_state_change_log AFTER INSERT OR DELETE OR UPDATE ON public.escalation_policy_state FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_escalation_policy_steps_change_log AFTER INSERT OR DELETE OR UPDATE ON public.escalation_policy_steps FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_heartbeat_monitors_change_log AFTER INSERT OR DELETE OR UPDATE ON public.heartbeat_monitors FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_integration_keys_change_log AFTER INSERT OR DELETE OR UPDATE ON public.integration_keys FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_keyring_change_log AFTER INSERT OR DELETE OR UPDATE ON public.keyring FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_labels_change_log AFTER INSERT OR DELETE OR UPDATE ON public.labels FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_notification_policy_cycles_change_log AFTER INSERT OR DELETE OR UPDATE ON public.notification_policy_cycles FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_outgoing_messages_change_log AFTER INSERT OR DELETE OR UPDATE ON public.outgoing_messages FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_region_ids_change_log AFTER INSERT OR DELETE OR UPDATE ON public.region_ids FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_rotation_participants_change_log AFTER INSERT OR DELETE OR UPDATE ON public.rotation_participants FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_rotation_state_change_log AFTER INSERT OR DELETE OR UPDATE ON public.rotation_state FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_rotations_change_log AFTER INSERT OR DELETE OR UPDATE ON public.rotations FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_schedule_on_call_users_change_log AFTER INSERT OR DELETE OR UPDATE ON public.schedule_on_call_users FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_schedule_rules_change_log AFTER INSERT OR DELETE OR UPDATE ON public.schedule_rules FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_schedules_change_log AFTER INSERT OR DELETE OR UPDATE ON public.schedules FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_services_change_log AFTER INSERT OR DELETE OR UPDATE ON public.services FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_twilio_sms_callbacks_change_log AFTER INSERT OR DELETE OR UPDATE ON public.twilio_sms_callbacks FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_twilio_sms_errors_change_log AFTER INSERT OR DELETE OR UPDATE ON public.twilio_sms_errors FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_twilio_voice_errors_change_log AFTER INSERT OR DELETE OR UPDATE ON public.twilio_voice_errors FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_user_contact_methods_change_log AFTER INSERT OR DELETE OR UPDATE ON public.user_contact_methods FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_user_favorites_change_log AFTER INSERT OR DELETE OR UPDATE ON public.user_favorites FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_user_last_alert_log_change_log AFTER INSERT OR DELETE OR UPDATE ON public.user_last_alert_log FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_user_notification_rules_change_log AFTER INSERT OR DELETE OR UPDATE ON public.user_notification_rules FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_user_overrides_change_log AFTER INSERT OR DELETE OR UPDATE ON public.user_overrides FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_user_verification_codes_change_log AFTER INSERT OR DELETE OR UPDATE ON public.user_verification_codes FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd

-- +migrate StatementBegin
DO $$ BEGIN
CREATE TRIGGER zz_99_users_change_log AFTER INSERT OR DELETE OR UPDATE ON public.users FOR EACH ROW EXECUTE PROCEDURE public.process_change();
EXCEPTION
  WHEN duplicate_object
  THEN null;
END $$;
-- +migrate StatementEnd
