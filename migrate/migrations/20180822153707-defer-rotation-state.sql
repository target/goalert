
-- +migrate Up
ALTER TABLE rotation_state
    DROP CONSTRAINT rotation_state_rotation_participant_id_fkey,
    ADD CONSTRAINT rotation_state_rotation_participant_id_fkey FOREIGN KEY (rotation_participant_id) REFERENCES rotation_participants (id) ON DELETE NO ACTION DEFERRABLE;

-- +migrate Down
ALTER TABLE rotation_state
    DROP CONSTRAINT rotation_state_rotation_participant_id_fkey,
    ADD CONSTRAINT rotation_state_rotation_participant_id_fkey FOREIGN KEY (rotation_participant_id) REFERENCES rotation_participants (id) ON DELETE RESTRICT;
