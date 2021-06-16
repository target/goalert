-- +migrate Up notransaction

ALTER TYPE enum_user_contact_method_type ADD VALUE IF NOT EXISTS 'WEBHOOK';

-- +migrate Down

