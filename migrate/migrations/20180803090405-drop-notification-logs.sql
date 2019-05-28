
-- +migrate Up
DROP TABLE notification_logs;
-- +migrate Down
CREATE TABLE notification_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    alert_id bigint NOT NULL,
    contact_method_id uuid NOT NULL,
    process_timestamp timestamp with time zone DEFAULT now() NOT NULL,
    completed boolean DEFAULT false NOT NULL
);

ALTER TABLE ONLY notification_logs
    ADD CONSTRAINT notification_logs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY notification_logs
    ADD CONSTRAINT notification_logs_alert_id_fkey FOREIGN KEY (alert_id) REFERENCES alerts(id) ON DELETE CASCADE;

ALTER TABLE ONLY notification_logs
    ADD CONSTRAINT notification_logs_contact_method_id_fkey FOREIGN KEY (contact_method_id) REFERENCES user_contact_methods(id) ON DELETE CASCADE;

