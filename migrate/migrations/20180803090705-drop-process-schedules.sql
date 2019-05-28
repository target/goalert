
-- +migrate Up
DROP TABLE process_schedules;
-- +migrate Down
CREATE TABLE process_schedules (
    schedule_id uuid NOT NULL,
    client_id uuid,
    deadline timestamp with time zone,
    last_processed timestamp with time zone
);

ALTER TABLE ONLY process_schedules
    ADD CONSTRAINT process_schedules_pkey PRIMARY KEY (schedule_id);

CREATE INDEX process_schedules_oldest_first ON public.process_schedules USING btree (last_processed NULLS FIRST);

ALTER TABLE ONLY process_schedules
    ADD CONSTRAINT process_schedules_schedule_id_fkey FOREIGN KEY (schedule_id) REFERENCES schedules(id) ON DELETE CASCADE;