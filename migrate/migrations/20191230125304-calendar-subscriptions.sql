-- +migrate Up
-- id, name, user_id, last_access, disabled, url(?)
CREATE TABLE calendar_subscriptions (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    user_id uuid NOT NULL UNIQUE REFERENCES users(id),
    last_access TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    disabled BOOLEAN NOT NULL DEFAULT false
);

-- +migrate Down
DROP TABLE calendar_subscriptions;
