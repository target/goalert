-- +migrate Up

CREATE TABLE user_calendar_subscriptions (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    last_access TIMESTAMPTZ,
    last_update TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    disabled BOOLEAN NOT NULL DEFAULT FALSE,
    schedule_id UUID NOT NULL REFERENCES schedules (id) ON DELETE CASCADE,
    config JSONB NOT NULL,

    UNIQUE(name, schedule_id, user_id)
);

-- +migrate Down

DROP TABLE user_calendar_subscriptions;
