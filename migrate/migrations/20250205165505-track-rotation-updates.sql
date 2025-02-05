-- +migrate Up
CREATE TABLE entity_updates(
    id bigserial PRIMARY KEY,
    entity_type text NOT NULL,
    entity_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
)
WITH (
    autovacuum_vacuum_threshold = 50, -- Lower threshold to trigger vacuum sooner
    autovacuum_vacuum_scale_factor = 0.05 -- Lower scale factor for frequent vacuuming
);

CREATE INDEX idx_entity_updates_entity_type ON entity_updates(entity_type);

INSERT INTO entity_updates(entity_type, entity_id)
SELECT
    'rotation',
    id
FROM
    rotations;

CREATE OR REPLACE FUNCTION track_rotation_updates()
    RETURNS TRIGGER
    AS $$
BEGIN
    INSERT INTO entity_updates(entity_type, entity_id)
        VALUES('rotation', NEW.id);
    RETURN new;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER track_rotation_updates
    AFTER INSERT OR UPDATE ON rotations FOR EACH ROW
    EXECUTE FUNCTION track_rotation_updates();

CREATE OR REPLACE FUNCTION track_rotation_dep_updates()
    RETURNS TRIGGER
    AS $$
BEGIN
    INSERT INTO entity_updates(entity_type, entity_id)
        VALUES('rotation', NEW.rotation_id);
    RETURN new;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER track_rotation_state_updates
    AFTER INSERT OR UPDATE ON rotation_state FOR EACH ROW
    EXECUTE FUNCTION track_rotation_dep_updates();

CREATE TRIGGER track_rotation_participants_updates
    AFTER INSERT OR UPDATE OR DELETE ON rotation_participants FOR EACH ROW
    EXECUTE FUNCTION track_rotation_dep_updates();

-- +migrate Down
DROP TABLE entity_updates;

