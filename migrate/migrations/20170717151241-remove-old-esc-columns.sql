
-- +migrate Up
ALTER TABLE escalation_policy DROP COLUMN team_id;

-- +migrate Down
ALTER TABLE escalation_policy ADD COLUMN team_id TEXT REFERENCES team (id);
