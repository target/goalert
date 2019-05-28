
-- +migrate Up notransaction

ALTER TABLE auth_basic_users
DROP COLUMN IF EXISTS id,
ADD COLUMN id BIGSERIAL,
ADD CONSTRAINT auth_basic_users_uniq_id UNIQUE(id);

ALTER TABLE auth_subjects
DROP COLUMN IF EXISTS id,
ADD COLUMN id BIGSERIAL,
ADD CONSTRAINT auth_subjects_uniq_id UNIQUE(id);

ALTER TABLE ep_step_on_call_users
DROP COLUMN IF EXISTS id,
ADD COLUMN id BIGSERIAL,
ADD CONSTRAINT ep_step_on_call_users_uniq_id UNIQUE(id);

ALTER TABLE escalation_policy_state
DROP COLUMN IF EXISTS id,
ADD COLUMN id BIGSERIAL,
ADD CONSTRAINT escalation_policy_state_uniq_id UNIQUE(id);

ALTER TABLE rotation_state
DROP COLUMN IF EXISTS id,
ADD COLUMN id BIGSERIAL,
ADD CONSTRAINT rotation_state_uniq_id UNIQUE(id);

ALTER TABLE schedule_on_call_users
DROP COLUMN IF EXISTS id,
ADD COLUMN id BIGSERIAL,
ADD CONSTRAINT schedule_on_call_users_uniq_id UNIQUE(id);

ALTER TABLE twilio_sms_callbacks
DROP COLUMN IF EXISTS id,
ADD COLUMN id BIGSERIAL,
ADD CONSTRAINT twilio_sms_callbacks_uniq_id UNIQUE(id);

ALTER TABLE twilio_sms_errors
DROP COLUMN IF EXISTS id,
ADD COLUMN id BIGSERIAL,
ADD CONSTRAINT twilio_sms_errors_uniq_id UNIQUE(id);

ALTER TABLE twilio_voice_errors
DROP COLUMN IF EXISTS id,
ADD COLUMN id BIGSERIAL,
ADD CONSTRAINT twilio_voice_errors_uniq_id UNIQUE(id);

ALTER TABLE user_favorites
DROP COLUMN IF EXISTS id,
ADD COLUMN id BIGSERIAL,
ADD CONSTRAINT user_favorites_uniq_id UNIQUE(id);

ALTER TABLE user_last_alert_log
DROP COLUMN IF EXISTS id,
ADD COLUMN id BIGSERIAL,
ADD CONSTRAINT user_last_alert_log_uniq_id UNIQUE(id);

-- +migrate Down notransaction

ALTER TABLE auth_basic_users DROP COLUMN IF EXISTS id;
ALTER TABLE auth_subjects DROP COLUMN IF EXISTS id;
ALTER TABLE ep_step_on_call_users DROP COLUMN IF EXISTS id;
ALTER TABLE escalation_policy_state DROP COLUMN IF EXISTS id;
ALTER TABLE rotation_state DROP COLUMN IF EXISTS id;
ALTER TABLE schedule_on_call_users DROP COLUMN IF EXISTS id;
ALTER TABLE twilio_sms_callbacks DROP COLUMN IF EXISTS id;
ALTER TABLE twilio_sms_errors DROP COLUMN IF EXISTS id;
ALTER TABLE twilio_voice_errors DROP COLUMN IF EXISTS id;
ALTER TABLE user_favorites DROP COLUMN IF EXISTS id;
ALTER TABLE user_last_alert_log DROP COLUMN IF EXISTS id;
