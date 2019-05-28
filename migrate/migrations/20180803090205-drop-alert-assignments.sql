
-- +migrate Up
DROP TABLE alert_assignments;
-- +migrate Down
CREATE TABLE alert_assignments (
    user_id uuid NOT NULL,
    alert_id bigint NOT NULL
);

ALTER TABLE ONLY alert_assignments
    ADD CONSTRAINT alert_assignments_pkey PRIMARY KEY (user_id, alert_id);

ALTER TABLE ONLY alert_assignments
    ADD CONSTRAINT alert_assignments_alert_id_fkey FOREIGN KEY (alert_id) REFERENCES alerts(id) ON DELETE CASCADE;

ALTER TABLE ONLY alert_assignments
    ADD CONSTRAINT alert_assignments_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
