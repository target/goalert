-- name: IntKeyRuleFindManyByIntKey :many
SELECT
    *
FROM
    integration_key_rules
WHERE integration_key_id = $1;

-- name: IntKeyRuleInsert :one
INSERT INTO integration_key_rules(name, integration_key_id, filter, filterobj, msgTemplate, action)
    VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    id;