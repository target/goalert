-- +migrate Up

CREATE INDEX om_cm_time_test_verify_idx ON outgoing_messages (contact_method_id, created_at)
WHERE message_type in ('test_notification', 'verification_message');

-- +migrate Down

DROP INDEX om_cm_time_test_verify_idx;
