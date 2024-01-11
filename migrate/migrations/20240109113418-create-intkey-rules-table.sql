-- +migrate Up

CREATE TABLE integration_key_rules(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    integration_key_id uuid NOT NULL,
    filter text NOT NULL,
    filterObj jsonb NOT NULL, -- json structure that evaluates to the expr text
    msgTemplate jsonb NOT NULL, -- json structure holding summary, details, and dedup for goalerts
    action int NOT NULL, --TODO: change to bit type and use bitmap to represent multiple actions
    CONSTRAINT integration_key_rules_integration_key_id_fkey FOREIGN KEY (integration_key_id) REFERENCES integration_keys(id) ON DELETE CASCADE
);

-- +migrate Down

DROP TABLE integration_key_rules;