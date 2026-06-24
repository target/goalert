-- +migrate Up

CREATE TABLE uik_logs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_key_id uuid NOT NULL REFERENCES integration_keys(id) ON DELETE CASCADE,
    received_at timestamptz NOT NULL DEFAULT now(),

    status text NOT NULL CHECK (status IN (
        'success', 
        'parse_error', 
        'exec_error', 
        'send_error'
    )),

    raw_body bytea,
    content_type text,
    user_agent text,
    error_message text
);

CREATE UNIQUE INDEX uik_logs_unique_key_id ON uik_logs (integration_key_id);

-- +migrate Down
DROP TABLE IF EXISTS uik_logs;
