-- +migrate Up
-- increment module version
UPDATE
    engine_processing_versions
SET
    version = 5
WHERE
    type_id = 'status_update';

ALTER TABLE alert_status_subscriptions
    ADD COLUMN updated_at timestamptz NOT NULL DEFAULT now();

CREATE INDEX alert_status_subscriptions_updated_at_idx ON alert_status_subscriptions(updated_at);

-- +migrate Down
UPDATE
    engine_processing_versions
SET
    version = 4
WHERE
    type_id = 'status_update';

DROP INDEX alert_status_subscriptions_updated_at_idx;

ALTER TABLE alert_status_subscriptions
    DROP COLUMN updated_at;

