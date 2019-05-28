
-- +migrate Up

DROP TYPE IF EXISTS enum_user_contact_method_type;
CREATE TYPE enum_user_contact_method_type as ENUM (
    'PUSH',
    'EMAIL',
    'VOICE',
    'SMS'
);

DROP TYPE IF EXISTS enum_user_contact_method_carrier;
CREATE TYPE enum_user_contact_method_carrier as ENUM (
    'ATT',
    'VERIZON',
    'SPRINT',
    'TMOBILE',
    'FI'
);

CREATE TABLE user_contact_methods (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    type enum_user_contact_method_type NOT NULL,
    value TEXT NOT NULL,
    carrier enum_user_contact_method_carrier,
    disabled BOOLEAN NOT NULL DEFAULT false,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE user_notification_rules (
    id UUID PRIMARY KEY,
    delay INT NOT NULL DEFAULT 0,
    contact_method_id UUID NOT NULL REFERENCES user_contact_methods (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE
);


CREATE EXTENSION IF NOT EXISTS "pgcrypto";

INSERT INTO user_contact_methods (id, name, type, value, carrier, disabled, user_id) 
    SELECT id::UUID, name, type::enum_user_contact_method_type, value, carrier::enum_user_contact_method_carrier, opt_out, user_id FROM contact;

INSERT INTO user_notification_rules (id, delay, contact_method_id, user_id)
    SELECT id::UUID, delay, contact_id::UUID, notification_rule.user_id FROM notification_rule;

DROP TABLE notification_rule;
DROP TABLE contact;

-- +migrate Down

CREATE TABLE contact (
	id TEXT PRIMARY KEY,
  name TEXT,
	type TEXT,
	value TEXT,
  carrier TEXT,
  opt_out BOOLEAN DEFAULT false,
	user_id UUID REFERENCES users (id)
);

INSERT INTO contact (id, name, type, value, carrier, opt_out, user_id)
    SELECT id::TEXT, name, type::TEXT, value, carrier::TEXT, disabled, user_id FROM user_contact_methods;

CREATE TABLE notification_rule (
	id TEXT PRIMARY KEY,
	delay INTEGER,
	user_id UUID REFERENCES users (id),
  contact_id TEXT REFERENCES contact(id)
);

INSERT INTO notification_rule (id, delay, user_id, contact_id)
    SELECT id::TEXT, delay, user_id, contact_method_id::TEXT FROM user_notification_rules;

DROP TABLE user_notification_rules;
DROP TABLE user_contact_methods;

DROP TYPE enum_user_contact_method_carrier;
DROP TYPE enum_user_contact_method_type;
