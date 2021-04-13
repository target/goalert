-- +migrate Up
drop table user_slack_data;

create table user_linked_accounts (
    id uuid not null primary key,
    account_id text not null,
    user_id uuid references users (id) on delete cascade,
    notification_channel_type enum_notif_channel_type not null,
    metadata jsonb
);

create unique index idx_slack_notification_channels ON user_linked_accounts (account_id, notification_channel_type)

create table notification_channel_last_alert_log (
    notification_channel_id uuid not null references notification_channels(id) on delete cascade,
    alert_id bigint not null references alerts(id) on delete cascade,
    log_id bigint not null references alert_logs(id) on delete cascade,
    next_log_id bigint not null references alert_logs(id) on delete cascade,
    primary key (notification_channel_id, alert_id)
);

-- +migrate StatementBegin
create function fn_insert_notification_channel_last_alert_log() returns trigger as $$
begin
    insert into notification_channel_last_alert_log
        (notification_channel_id, alert_id, log_id, next_log_id)
    values
        (new.sub_channel_id, new.alert_id, new.id, new.id)
    on conflict do nothing;
return new;
end;
$$ language plpgsql;
-- +migrate StatementEnd

-- +migrate StatementBegin
create function fn_update_notification_channel_last_alert_log() returns trigger as $$
begin
    update notification_channel_last_alert_log last
    set next_log_id
    = new.id
    where
        last.alert_id = new.alert_id and
        new.id > last.next_log_id;
return new;
end;
$$ language plpgsql;
-- +migrate StatementEnd

create trigger trg_insert_alert_logs_notification_channel_last_alert
after insert on alert_logs for each row when (new.event = 'notification_sent' and new.sub_type = 'channel')
execute procedure fn_insert_notification_channel_last_alert_log();

create trigger trg_insert_alert_logs_notification_channel_last_alert_update
after insert on alert_logs for each row when (new.event in ('acknowledged', 'closed'))
execute procedure fn_update_notification_channel_last_alert_log();

-- +migrate Down
drop trigger trg_insert_alert_logs_notification_channel_last_alert_update on alert_logs;
drop trigger trg_insert_alert_logs_notification_channel_last_alert on alert_logs;
drop function fn_update_notification_channel_last_alert_log();
drop function fn_insert_notification_channel_last_alert_log();
drop table notification_channel_last_alert_log;
drop table user_linked_accounts;

create table user_slack_data (
    id uuid not null primary key references users (id) on delete cascade,
    access_token text not null
);
