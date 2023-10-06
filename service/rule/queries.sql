-- name: SvcRuleFindMany :many
SELECT
    service_rules.id,
    service_rules.name,
    service_rules.service_id,
    service_rules.filter,
    service_rules.send_alert,
    service_rules.actions,
    STRING_AGG(service_rule_integration_keys.integration_key_id::text, ',')::text integration_keys
FROM
    service_rules
    JOIN service_rule_integration_keys ON service_rule_integration_keys.service_rule_id = service_rules.id
WHERE
    service_rules.id = ANY (@service_rule_ids::uuid[])
GROUP BY
    service_rules.id;

-- name: SvcRuleFindManyByService :many
SELECT
    service_rules.id,
    service_rules.name,
    service_rules.service_id,
    service_rules.filter,
    service_rules.send_alert,
    service_rules.actions,
    STRING_AGG(service_rule_integration_keys.integration_key_id::text, ',')::text integration_keys
FROM
    service_rules
    JOIN service_rule_integration_keys ON service_rule_integration_keys.service_rule_id = service_rules.id
WHERE
    service_rules.service_id = $1
GROUP BY
    service_rules.id;

-- name: SvcRuleInsert :one
INSERT INTO service_rules(name, service_id, FILTER, send_alert, actions)
    VALUES ($1, $2, $3, $4, $5)
RETURNING
    id;

-- name: SvcRuleSetIntKeys :exec
WITH deleted_rows AS (
    DELETE FROM service_rule_integration_keys sk
    WHERE sk.service_rule_id = $1
        AND sk.integration_key_id != ALL (@integration_key_ids::uuid[])
    RETURNING
        *)
INSERT INTO service_rule_integration_keys(service_rule_id, integration_key_id)
SELECT
    $1,
    ik
FROM
    unnest(@integration_key_ids::uuid[]) ik
ON CONFLICT
    DO NOTHING;

-- name: SvcRuleFindManyByIntKey :many
SELECT
    r.id,
    r.name,
    r.service_id,
    r.filter,
    r.send_alert,
    r.actions
FROM
    service_rule_integration_keys AS sk
    JOIN service_rules AS r ON sk.service_rule_id = r.id
        AND r.service_id = $1
        AND sk.integration_key_id = $2;

-- name: SvcRuleFindOne :one
SELECT
    r.id,
    r.name,
    r.service_id,
    r.filter,
    r.send_alert,
    r.actions,
    STRING_AGG(si.integration_key_id::text, ',')::text integration_keys
FROM
    service_rules r
    JOIN service_rule_integration_keys si ON si.service_rule_id = r.id
WHERE
    r.id = $1
GROUP BY
    r.id;

-- name: SvcRuleUpdate :exec
UPDATE
    service_rules
SET
    name = $2,
    FILTER = $3,
    send_alert = $4,
    actions = $5
WHERE
    id = $1;

-- name: SvcRuleDelete :exec
DELETE FROM service_rules
WHERE id = $1;

