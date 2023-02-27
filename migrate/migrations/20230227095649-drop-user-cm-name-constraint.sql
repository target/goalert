-- +migrate Up
ALTER TABLE user_contact_methods DROP CONSTRAINT user_contact_methods_name_type_user_id_key;

-- +migrate Down

ALTER TABLE user_contact_methods ADD CONSTRAINT user_contact_methods_name_type_user_id_key UNIQUE(name, type, user_id);
