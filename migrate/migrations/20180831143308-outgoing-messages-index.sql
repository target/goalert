
-- +migrate Up notransaction

drop index if exists idx_om_alert_log_id;
create index concurrently idx_om_alert_log_id on outgoing_messages (alert_log_id);

drop index if exists idx_om_vcode_id;
create index concurrently idx_om_vcode_id on outgoing_messages (user_verification_code_id);

-- +migrate Down notransaction

drop index if exists idx_om_alert_log_id;
drop index if exists idx_om_vcode_id;
