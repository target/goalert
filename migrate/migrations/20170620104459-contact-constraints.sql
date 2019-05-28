
-- +migrate Up

ALTER TABLE user_contact_methods ADD UNIQUE(name, type, user_id);
ALTER TABLE user_contact_methods ADD UNIQUE(type, value);
ALTER TABLE user_contact_methods ALTER id SET DEFAULT gen_random_uuid();

-- +migrate Down

ALTER TABLE user_contact_methods DROP CONSTRAINT UNIQUE(name, type, user_id);
ALTER TABLE user_contact_methods DROP CONSTRAINT UNIQUE(type, value);
ALTER TABLE user_contact_methods ALTER id DROP DEFAULT;
