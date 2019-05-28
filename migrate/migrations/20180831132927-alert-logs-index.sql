
-- +migrate Up notransaction

drop index if exists idx_alert_logs_alert_id;
create index concurrently idx_alert_logs_alert_id on alert_logs (alert_id);


drop index if exists idx_alert_logs_hb_id;
create index concurrently idx_alert_logs_hb_id on alert_logs (sub_hb_monitor_id);

drop index if exists idx_alert_logs_int_id;
create index concurrently idx_alert_logs_int_id on alert_logs (sub_integration_key_id);

drop index if exists idx_alert_logs_user_id;
create index concurrently idx_alert_logs_user_id on alert_logs (sub_user_id);

-- +migrate Down notransaction
drop index if exists idx_alert_logs_alert_id;
drop index if exists idx_alert_logs_hb_id;
drop index if exists idx_alert_logs_int_id;
drop index if exists idx_alert_logs_user_id;
