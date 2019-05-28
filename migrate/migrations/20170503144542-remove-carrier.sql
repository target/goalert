
-- +migrate Up

ALTER TABLE user_contact_methods DROP COLUMN carrier;
DROP TYPE enum_user_contact_method_carrier;

-- +migrate Down

CREATE TYPE enum_user_contact_method_carrier as ENUM (
    'ATT',
    'VERIZON',
    'SPRINT',
    'TMOBILE',
    'FI'
);

ALTER TABLE user_contact_methods ADD COLUMN carrier enum_user_contact_method_carrier;
