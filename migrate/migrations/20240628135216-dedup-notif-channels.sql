-- +migrate Up
CREATE TABLE notification_channel_duplicates(
    old_id uuid PRIMARY KEY,
    new_id uuid NOT NULL,
    old_created_at timestamp with time zone NOT NULL
);

LOCK notification_channels;

INSERT INTO notification_channel_duplicates(old_id, old_created_at, new_id)
SELECT
    nc.id,
    nc.created_at,
(
        SELECT
            id
        FROM
            notification_channels sub
        WHERE
            sub.type = nc.type
            AND sub.value = nc.value
        ORDER BY
            sub.created_at
        LIMIT 1)
FROM
    notification_channels nc
WHERE (
    SELECT
        count(*)
    FROM
        notification_channels sub
    WHERE
        sub.type = nc.type
        AND sub.value = nc.value) > 1;

-- remove any non-duplicates
DELETE FROM notification_channel_duplicates
WHERE old_id = new_id;

-- update alert_logs
UPDATE
    alert_logs
SET
    sub_channel_id = dup.new_id
FROM
    notification_channel_duplicates dup
WHERE
    sub_channel_id = dup.old_id;

-- update ep actions, status updates, and outgoing_messages
-- ensure that no updates/inserts can be made until we are done
LOCK TABLE escalation_policy_actions IN share ROW exclusive mode;

UPDATE
    escalation_policy_actions
SET
    channel_id = dup.new_id
FROM
    notification_channel_duplicates dup
WHERE
    channel_id = dup.old_id;

LOCK TABLE alert_status_subscriptions IN share ROW exclusive mode;

UPDATE
    alert_status_subscriptions
SET
    channel_id = dup.new_id
FROM
    notification_channel_duplicates dup
WHERE
    channel_id = dup.old_id;

LOCK TABLE outgoing_messages IN share ROW exclusive mode;

UPDATE
    outgoing_messages
SET
    channel_id = dup.new_id
FROM
    notification_channel_duplicates dup
WHERE
    channel_id = dup.old_id;

-- delete duplicates
DELETE FROM notification_channels USING notification_channel_duplicates
WHERE id = old_id;

ALTER TABLE notification_channels
    ADD CONSTRAINT nc_unique_type_value UNIQUE (type, value);

-- +migrate Down
ALTER TABLE notification_channels
    DROP CONSTRAINT nc_unique_type_value;

INSERT INTO notification_channels(id, type, value, name, created_at)
SELECT
    old_id,
    type,
    value,
    name,
    old_created_at
FROM
    notification_channel_duplicates
    JOIN notification_channels ON notification_channels.id = new_id;

DROP TABLE notification_channel_duplicates;

