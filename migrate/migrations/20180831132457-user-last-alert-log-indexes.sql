
-- +migrate Up notransaction
drop index if exists idx_ulal_log_id;
create index concurrently idx_ulal_log_id on user_last_alert_log (log_id);

drop index if exists idx_ulal_next_log_id;
create index concurrently idx_ulal_next_log_id on user_last_alert_log (next_log_id);

drop index if exists idx_ulal_alert_id;
create index concurrently idx_ulal_alert_id on user_last_alert_log (alert_id);

-- +migrate Down notransaction
drop index if exists idx_ulal_log_id;
drop index if exists idx_ulal_next_log_id;
drop index if exists idx_ulal_alert_id;
