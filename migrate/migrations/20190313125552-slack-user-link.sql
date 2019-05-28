
-- +migrate Up

CREATE TABLE user_slack_data (
    id UUID NOT NULL PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    access_token TEXT NOT NULL
);

-- +migrate Down

DROP TABLE user_slack_data;
