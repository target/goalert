-- +migrate Up

CREATE TABLE schedule_data (
    schedule_id UUID REFERENCES schedules (id) PRIMARY KEY ON DELETE CASCADE,
    last_updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_updated_by UUID REFERENCES users (id) ON DELETE SET NULL,
    last_cleanup_at TIMESTAMPTZ,
    data JSONB NOT NULL
);

-- +migrate Down

DROP TABLE schedule_data;
