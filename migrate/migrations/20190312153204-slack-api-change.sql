
-- +migrate Up

UPDATE engine_processing_versions
SET "version" = 6
WHERE type_id = 'message';

UPDATE notification_channels
SET
    meta = jsonb_set(meta, '{webhookURL}', to_jsonb(value), true),
    value = meta->>'chanID'
WHERE type = 'SLACK';

-- +migrate Down

UPDATE notification_channels
SET
    value = meta->>'webhookURL',
    meta = meta - 'webhookURL'
WHERE type = 'SLACK';

UPDATE engine_processing_versions
SET "version" = 5
WHERE type_id = 'message';
