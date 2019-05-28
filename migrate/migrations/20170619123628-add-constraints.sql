
-- +migrate Up
ALTER TABLE escalation_policy ADD UNIQUE(name);
ALTER TABLE service ADD UNIQUE(name);


-- +migrate Down
ALTER TABLE escalation_policy DROP CONSTRAINT UNIQUE(name);
ALTER TABLE service DROP CONSTRAINT UNIQUE(name);
