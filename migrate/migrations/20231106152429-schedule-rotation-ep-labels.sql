-- +migrate Up
ALTER TABLE labels
    ADD COLUMN tgt_schedule_id uuid REFERENCES schedules(id) ON DELETE CASCADE,
    ADD COLUMN tgt_rotation_id uuid REFERENCES rotations(id) ON DELETE CASCADE,
    ADD COLUMN tgt_ep_id uuid REFERENCES escalation_policies(id) ON DELETE CASCADE,
    DROP CONSTRAINT labels_tgt_service_id_key_key,
    ALTER COLUMN tgt_service_id DROP NOT NULL;

-- drop redundant index
DROP INDEX idx_labels_service_id;

CREATE UNIQUE INDEX idx_labels_resource_id ON labels(tgt_service_id, tgt_schedule_id, tgt_rotation_id, tgt_ep_id, key);

-- +migrate Down
-- Since we're re-adding the NOT NULL constraint, we need to make sure there are no NULL values
-- in the column. We can do this by deleting all rows where the service ID is NULL.
DELETE FROM labels
WHERE tgt_service_id IS NULL;

ALTER TABLE labels
    DROP COLUMN tgt_schedule_id,
    DROP COLUMN tgt_rotation_id,
    DROP COLUMN tgt_ep_id,
    ADD CONSTRAINT labels_tgt_service_id_key_key UNIQUE (tgt_service_id, key),
    ALTER COLUMN tgt_service_id SET NOT NULL;

CREATE INDEX idx_labels_service_id ON labels(tgt_service_id);

