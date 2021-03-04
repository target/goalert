-- +migrate Up

ALTER TABLE user_slack_data RENAME TO slack_users;
ALTER TABLE slack_users RENAME COLUMN id TO user_id;
ALTER TABLE slack_users ADD COLUMN slack_id TEXT;

-- +migrate Down

ALTER TABLE slack_users DROP COLUMN slack_id;
ALTER TABLE slack_users RENAME COLUMN user_id TO id;
ALTER TABLE slack_users RENAME TO user_slack_data;