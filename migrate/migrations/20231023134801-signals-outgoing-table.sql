-- +migrate Up
ALTER TABLE outgoing_signals
DROP column IF EXISTS outgoing_payload,
DROP column IF EXISTS channel_id,
ADD column IF NOT EXISTS signal_id int NOT NULL,
ADD column IF NOT EXISTS destination_type text NOT NULL,
ADD column IF NOT EXISTS destination_type text NOT NULL,
ADD column IF NOT EXISTS destination_id text NOT NULL,
ADD column IF NOT EXISTS destination_val text NOT NULL,
ADD column IF NOT EXISTS content jsonb NOT NULL,
ADD column IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT now(),
ADD column IF NOT EXISTS sent_at timestamptz;

-- +migrate Down

