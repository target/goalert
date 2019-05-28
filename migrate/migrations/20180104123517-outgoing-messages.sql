
-- +migrate Up

CREATE TYPE enum_outgoing_messages_status AS ENUM (
    'pending',
    'sending',
    'queued_remotely', -- for use when sent, but we have status that a remote system has it queued
    'sent',
    'delivered', -- delivery confirmation
    'failed'
);
CREATE TYPE enum_outgoing_messages_type AS ENUM (
    'alert_notification',
    'verification_message',
    'test_notification'
);

CREATE TABLE outgoing_messages (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    message_type enum_outgoing_messages_type NOT NULL,
    contact_method_id UUID NOT NULL REFERENCES user_contact_methods (id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    last_status enum_outgoing_messages_status NOT NULL DEFAULT 'pending',
    last_status_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    status_details TEXT NOT NULL DEFAULT '',
    fired_at TIMESTAMP WITH TIME ZONE,
    sent_at TIMESTAMP WITH TIME ZONE,
    retry_count INT NOT NULL DEFAULT 0,
    next_retry_at TIMESTAMP WITH TIME ZONE,
    sending_deadline TIMESTAMP WITH TIME ZONE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,

    alert_id BIGINT REFERENCES alerts (id) ON DELETE CASCADE,
    cycle_id UUID REFERENCES notification_policy_cycles (id) ON DELETE CASCADE,
    service_id UUID REFERENCES services (id) ON DELETE CASCADE,
    escalation_policy_id UUID REFERENCES escalation_policies (id) ON DELETE CASCADE,

    CONSTRAINT om_pending_no_fired_no_sent CHECK(
        last_status != 'pending' or (fired_at isnull and sent_at isnull)
    ),
    CONSTRAINT om_sending_fired_no_sent CHECK(
        last_status != 'sending' or (fired_at notnull and sent_at isnull)
    ),
    CONSTRAINT om_processed_no_fired_sent CHECK(
        last_status in ('pending','sending','failed') or
        (fired_at isnull and sent_at notnull)
    ),
    CONSTRAINT om_alert_svc_ep_ids CHECK(
        message_type != 'alert_notification' or (
            alert_id notnull and
            service_id notnull and
            escalation_policy_id notnull
        )
    ),
    CONSTRAINT om_sending_deadline_reqd CHECK(
        last_status != 'sending' or sending_deadline notnull
    )
);

CREATE INDEX idx_om_alert_sent ON outgoing_messages (alert_id, sent_at);
CREATE INDEX idx_om_ep_sent ON outgoing_messages (escalation_policy_id, sent_at);
CREATE INDEX idx_om_service_sent ON outgoing_messages (service_id, sent_at);
CREATE INDEX idx_om_cm_sent ON outgoing_messages (contact_method_id, sent_at);
CREATE INDEX idx_om_user_sent ON outgoing_messages (user_id, sent_at);

-- +migrate Down

DROP TABLE outgoing_messages;
DROP TYPE enum_outgoing_messages_type;
DROP TYPE enum_outgoing_messages_status;
