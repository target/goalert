
-- +migrate Up
-- Add new integration key type 'generic'
ALTER TYPE enum_integration_keys_type RENAME TO enum_integration_keys_type_old;
CREATE TYPE enum_integration_keys_type AS ENUM (
	'grafana',
	'generic'
);
ALTER TABLE integration_keys ALTER COLUMN type TYPE enum_integration_keys_type USING type::TEXT::enum_integration_keys_type;
DROP TYPE enum_integration_keys_type_old;

-- Add new alert source type 'generic'
ALTER TYPE enum_alert_source RENAME TO enum_alert_source_old;
CREATE TYPE enum_alert_source AS ENUM (
	'grafana',
	'manual',
	'generic'
);
ALTER TABLE alerts
	ALTER COLUMN source DROP DEFAULT,
	ALTER COLUMN source TYPE enum_alert_source USING source::TEXT::enum_alert_source,
	ALTER COLUMN source SET DEFAULT 'manual';
DROP TYPE enum_alert_source_old;


-- +migrate Down

-- Go back to just grafana keys (generic keys will manually have to be dropped first if they exist)
ALTER TYPE enum_integration_keys_type RENAME TO enum_integration_keys_type_old;
CREATE TYPE enum_integration_keys_type AS ENUM (
	'grafana'
);
ALTER TABLE integration_keys ALTER COLUMN type TYPE enum_integration_keys_type USING type::TEXT::enum_integration_keys_type;
DROP TYPE enum_integration_keys_type_old;

-- Go back to just grafana keys (generic keys will manually have to be dropped first if they exist)
ALTER TYPE enum_alert_source RENAME TO enum_alert_source_old;
CREATE TYPE enum_alert_source AS ENUM (
	'grafana',
	'manual'
);
ALTER TABLE alerts
	ALTER COLUMN source DROP DEFAULT,
	ALTER COLUMN source TYPE enum_alert_source USING source::TEXT::enum_alert_source,
	ALTER COLUMN source SET DEFAULT 'manual';;
DROP TYPE enum_alert_source_old;

