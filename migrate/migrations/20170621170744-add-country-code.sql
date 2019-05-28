
-- +migrate Up
UPDATE user_contact_methods SET value = '+'||value WHERE (type = 'SMS' OR type = 'VOICE') AND value NOT LIKE '+%';

-- +migrate Down

UPDATE user_contact_methods SET value = substring(value from 2) WHERE (type = 'SMS' OR type = 'VOICE') AND value LIKE '+%';
