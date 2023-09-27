-- +migrate Up
ALTER TYPE enum_integration_keys_type
    ADD VALUE IF NOT EXISTS 'signal';

ALTER TYPE enum_alert_source
    ADD VALUE IF NOT EXISTS 'signal';

CREATE TABLE service_rules(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    service_id uuid NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    filter TEXT NOT NULL,
    send_alert boolean NOT NULL,
    actions jsonb
);

CREATE TABLE service_rule_integration_keys(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    service_rule_id uuid NOT NULL REFERENCES service_rules(id) ON DELETE CASCADE,
    integration_key_id uuid NOT NULL REFERENCES integration_keys(id) ON DELETE CASCADE,
    CONSTRAINT unique_int_key_service_rule UNIQUE (service_rule_id, integration_key_id)
);

CREATE TABLE signals(
    id bigserial PRIMARY KEY,
    service_id uuid NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    service_rule_id uuid NOT NULL REFERENCES service_rules(id) ON DELETE CASCADE,
    timestamp timestamptz NOT NULL DEFAULT now(),
    outgoing_payload jsonb NOT NULL,
    scheduled boolean NOT NULL
);

CREATE TABLE outgoing_signals(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id uuid NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    outgoing_payload jsonb NOT NULL,
    channel_id uuid NOT NULL REFERENCES notification_channels(id) ON DELETE CASCADE
);

-- +migrate Down
DROP TABLE outgoing_signals;

DROP TABLE signals;

DROP TABLE service_rule_integration_keys;

DROP TABLE service_rules;

