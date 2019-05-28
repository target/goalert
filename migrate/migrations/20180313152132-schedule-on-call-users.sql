
-- +migrate Up

CREATE TABLE schedule_on_call_users (
    schedule_id UUID NOT NULL REFERENCES schedules (id) ON DELETE CASCADE,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    end_time TIMESTAMP WITH TIME ZONE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,

    CHECK(end_time ISNULL OR end_time > start_time),

    UNIQUE(schedule_id, user_id, end_time)
);

-- +migrate Down

DROP TABLE schedule_on_call_users;
