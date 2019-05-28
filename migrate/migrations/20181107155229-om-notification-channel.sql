
-- +migrate Up
ALTER TABLE outgoing_messages
	ADD COLUMN channel_id UUID REFERENCES notification_channels (id) ON DELETE CASCADE,
	ALTER COLUMN user_id DROP NOT NULL,
	ALTER COLUMN contact_method_id DROP NOT NULL,
	ADD CONSTRAINT om_user_cm_or_channel CHECK(
		(user_id notnull and contact_method_id notnull and channel_id isnull)
		or
		(channel_id notnull and contact_method_id isnull and user_id isnull)
	);
-- +migrate Down
ALTER TABLE outgoing_messages
	DROP CONSTRAINT om_user_cm_or_channel,
	DROP COLUMN channel_id,
	ALTER COLUMN user_id SET NOT NULL,
	ALTER COLUMN contact_method_id SET NOT NULL;

