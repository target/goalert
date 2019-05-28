
-- +migrate Up notransaction
drop index if exists idx_alert_service_id;
create index concurrently idx_alert_service_id on alerts (service_id);

-- +migrate Down notransaction
drop index if exists idx_alert_service_id;
