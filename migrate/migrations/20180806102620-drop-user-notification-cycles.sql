
-- +migrate Up
DROP TABLE user_notification_cycles;

-- +migrate Down
CREATE TABLE user_notification_cycles (
    id uuid NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    user_id uuid REFERENCES users(id) ON DELETE CASCADE,
    alert_id bigint REFERENCES alerts(id) ON DELETE CASCADE,
    escalation_level integer NOT NULL,
    started_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT user_notification_cycles_pkey PRIMARY KEY (user_id, alert_id)
);
