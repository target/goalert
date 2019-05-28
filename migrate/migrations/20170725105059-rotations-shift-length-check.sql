
-- +migrate Up

ALTER TABLE rotations ADD CHECK (shift_length > 0);

-- +migrate Down

ALTER TABLE rotations DROP CONSTRAINT rotations_shift_length_check;

