
-- +migrate Up
DROP TABLE user_contact_method_locks;
-- +migrate Down
CREATE TABLE user_contact_method_locks (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    client_id uuid NOT NULL,
    alert_id bigint NOT NULL,
    contact_method_id uuid NOT NULL,
    "timestamp" timestamp with time zone DEFAULT now() NOT NULL
);

ALTER TABLE ONLY user_contact_method_locks
    ADD CONSTRAINT user_contact_method_locks_alert_id_contact_method_id_key UNIQUE (alert_id, contact_method_id);

ALTER TABLE ONLY user_contact_method_locks
    ADD CONSTRAINT user_contact_method_locks_pkey PRIMARY KEY (id);

ALTER TABLE ONLY user_contact_method_locks
    ADD CONSTRAINT user_contact_method_locks_alert_id_fkey FOREIGN KEY (alert_id) REFERENCES alerts(id) ON DELETE CASCADE;

ALTER TABLE ONLY user_contact_method_locks
    ADD CONSTRAINT user_contact_method_locks_contact_method_id_fkey FOREIGN KEY (contact_method_id) REFERENCES user_contact_methods(id) ON DELETE CASCADE;