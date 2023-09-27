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

-- name: SvcRuleInsert :exec
INSERT INTO service_rules(name, service_id, filter, send_alert, actions)
    VALUES ($1, $2, $3, $4, $5);

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

