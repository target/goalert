-- +migrate Up
CREATE TABLE uik_samples(
    id uuid PRIMARY KEY,
    key_id uuid NOT NULL REFERENCES integration_keys(id) ON DELETE CASCADE,
    request_data jsonb NOT NULL,
    created_at timestamp NOT NULL DEFAULT now(),
    failed boolean NOT NULL DEFAULT FALSE,
    user_note text
);

CREATE INDEX uik_samples_uik_created_idx ON uik_samples(key_id, created_at DESC)
WHERE
    user_note IS NULL AND failed = FALSE;

CREATE INDEX uik_samples_uik_failed_idx ON uik_samples(key_id, created_at DESC)
WHERE
    user_note IS NULL AND failed = TRUE;

-- +migrate Down
DROP TABLE uik_samples;

