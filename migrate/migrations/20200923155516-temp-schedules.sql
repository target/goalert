-- +migrate Up

CREATE TABLE schedule_data (
    schedule_id UUID PRIMARY KEY REFERENCES schedules (id) ON DELETE CASCADE,
    last_cleanup_at TIMESTAMPTZ,
    data JSONB NOT NULL
);

-- +migrate Down

DROP TABLE schedule_data;
