-- +migrate Up
create index idx_alert_cleanup on alerts (id, created_at) where status = 'closed';

-- +migrate Down
drop index idx_alert_cleanup;
