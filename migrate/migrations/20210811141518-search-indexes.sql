-- +migrate Up notransaction
create index concurrently if not exists idx_search_alerts_summary_eng on alerts using gin (to_tsvector('english', replace(lower(summary), '.', ' ')));

create index concurrently if not exists idx_search_escalation_policies_name_eng on escalation_policies using gin (to_tsvector('english', replace(lower(name), '.', ' ')));
create index concurrently if not exists idx_search_escalation_policies_desc_eng on escalation_policies using gin (to_tsvector('english', replace(lower(description), '.', ' ')));

create index concurrently if not exists idx_search_rotations_name_eng on rotations using gin (to_tsvector('english', replace(lower(name), '.', ' ')));
create index concurrently if not exists idx_search_rotations_desc_eng on rotations using gin (to_tsvector('english', replace(lower(description), '.', ' ')));

create index concurrently if not exists idx_search_schedules_name_eng on schedules using gin (to_tsvector('english', replace(lower(name), '.', ' ')));
create index concurrently if not exists idx_search_schedules_desc_eng on schedules using gin (to_tsvector('english', replace(lower(description), '.', ' ')));

create index concurrently if not exists idx_search_services_name_eng on services using gin (to_tsvector('english', replace(lower(name), '.', ' ')));
create index concurrently if not exists idx_search_services_desc_eng on services using gin (to_tsvector('english', replace(lower(description), '.', ' ')));

create index concurrently if not exists idx_search_users_name_eng on users using gin (to_tsvector('english', replace(lower(name), '.', ' ')));

-- +migrate Down

drop index idx_search_alerts_summary_eng;
drop index idx_search_escalation_policies_name_eng;
drop index idx_search_escalation_policies_desc_eng;
drop index idx_search_rotations_name_eng;
drop index idx_search_rotations_desc_eng;
drop index idx_search_schedules_name_eng;
drop index idx_search_schedules_desc_eng;
drop index idx_search_services_name_eng;
drop index idx_search_services_desc_eng;
drop index idx_search_users_name_eng;
