-- +migrate Up
CREATE TABLE data_encryption_key_metadata (
    id UUID PRIMARY KEY,
    version INT NOT NULL,
    active BOOLEAN,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    mk_digest BYTEA NOT NULL,
    mk_digest_salt BYTEA NOT NULL,
    mk_digest_iter INT NOT NULL,
    check(active),
    unique(active)
);

CREATE TABLE data_encryption_key_slots (
    key_id UUID NOT NULL REFERENCES data_encryption_key_metadata(id) ON DELETE CASCADE,
    slot_id INT NOT NULL,
    version INT NOT NULL,
    salt BYTEA NOT NULL,
    iterations INT NOT NULL,
    stripes INT NOT NULL,
    key_material BYTEA NOT NULL,

    PRIMARY KEY (key_id, slot_id)
);

-- +migrate Down
