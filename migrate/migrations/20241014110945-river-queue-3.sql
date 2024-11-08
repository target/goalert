-- +migrate Up
-- River migration 003 [up]
ALTER TABLE river_job ALTER COLUMN tags SET DEFAULT '{}';
UPDATE river_job SET tags = '{}' WHERE tags IS NULL;
ALTER TABLE river_job ALTER COLUMN tags SET NOT NULL;

-- +migrate Down
-- River migration 003 [down]
ALTER TABLE river_job ALTER COLUMN tags DROP NOT NULL,
                      ALTER COLUMN tags DROP DEFAULT;
