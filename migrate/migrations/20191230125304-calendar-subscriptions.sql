-- +migrate Up
CREATE TABLE calendar_subscriptions (
    id uuid NOT NULL UNIQUE,
    name TEXT NOT NULL,
    user_id uuid NOT NULL REFERENCES users(id),
    last_access TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    disabled BOOLEAN NOT NULL DEFAULT false
);

-- +migrate Down
DROP TABLE calendar_subscriptions;
