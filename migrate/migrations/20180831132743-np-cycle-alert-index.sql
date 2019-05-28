
-- +migrate Up notransaction
drop index if exists idx_np_cycle_alert_id;
create index concurrently idx_np_cycle_alert_id on notification_policy_cycles (alert_id);

-- +migrate Down notransaction
drop index if exists idx_np_cycle_alert_id;
