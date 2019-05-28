-- +migrate Up
CREATE TABLE labels (
    id BIGSERIAL PRIMARY KEY,
    tgt_service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    UNIQUE (tgt_service_id, key)
);

CREATE INDEX idx_labels_service_id ON labels (tgt_service_id);

-- +migrate Down
DROP INDEX idx_labels_service_id;
DROP TABLE labels;