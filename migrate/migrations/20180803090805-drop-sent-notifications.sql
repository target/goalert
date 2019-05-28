
-- +migrate Up
DROP TABLE sent_notifications;

-- +migrate Down
CREATE TABLE sent_notifications (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    alert_id bigint NOT NULL,
    contact_method_id uuid NOT NULL,
    sent_at timestamp with time zone,
    cycle_id uuid NOT NULL,
    notification_rule_id uuid NOT NULL
);

ALTER TABLE ONLY sent_notifications
    ADD CONSTRAINT sent_notifications_notification_rule_id_cycle_id_key UNIQUE (notification_rule_id, cycle_id);

CREATE INDEX sent_notifications_id_idx ON public.sent_notifications USING btree (id);

ALTER TABLE ONLY sent_notifications
    ADD CONSTRAINT sent_notifications_alert_id_fkey FOREIGN KEY (alert_id) REFERENCES alerts(id) ON DELETE CASCADE;

ALTER TABLE ONLY sent_notifications
    ADD CONSTRAINT sent_notifications_contact_method_id_fkey FOREIGN KEY (contact_method_id) REFERENCES user_contact_methods(id) ON DELETE CASCADE;

ALTER TABLE ONLY sent_notifications
    ADD CONSTRAINT sent_notifications_notification_rule_id_fkey FOREIGN KEY (notification_rule_id) REFERENCES user_notification_rules(id) ON DELETE CASCADE;
