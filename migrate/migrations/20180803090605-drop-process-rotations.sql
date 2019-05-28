
-- +migrate Up
DROP TABLE process_rotations;

-- +migrate Down
CREATE TABLE process_rotations (
    rotation_id uuid NOT NULL,
    client_id uuid,
    deadline timestamp with time zone,
    last_processed timestamp with time zone
);

ALTER TABLE ONLY process_rotations
    ADD CONSTRAINT process_rotations_pkey PRIMARY KEY (rotation_id);

CREATE INDEX process_rotations_oldest_first ON public.process_rotations USING btree (last_processed NULLS FIRST);

ALTER TABLE ONLY process_rotations
    ADD CONSTRAINT process_rotations_rotation_id_fkey FOREIGN KEY (rotation_id) REFERENCES rotations(id) ON DELETE CASCADE;
