
-- +migrate Up
DROP TABLE alert_escalation_policy_snapshots;
-- +migrate Down
CREATE TABLE alert_escalation_policy_snapshots (
    alert_id bigint NOT NULL,
    step_number integer NOT NULL,
    step_max integer NOT NULL,
    step_delay interval NOT NULL,
    repeat integer NOT NULL,
    user_id uuid,
    schedule_id uuid
);

ALTER TABLE ONLY alert_escalation_policy_snapshots
    ADD CONSTRAINT alert_escalation_policy_snapshots_alert_id_fkey FOREIGN KEY (alert_id) REFERENCES alerts(id) ON DELETE CASCADE;

ALTER TABLE ONLY alert_escalation_policy_snapshots
    ADD CONSTRAINT alert_escalation_policy_snapshots_schedule_id_fkey FOREIGN KEY (schedule_id) REFERENCES schedules(id) ON DELETE CASCADE;

ALTER TABLE ONLY alert_escalation_policy_snapshots
    ADD CONSTRAINT alert_escalation_policy_snapshots_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;