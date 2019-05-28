
-- +migrate Up


CREATE TABLE rotation_state (
    rotation_id UUID PRIMARY KEY REFERENCES rotations (id) ON DELETE CASCADE,
    position INT NOT NULL DEFAULT 0,

    -- it's ok if it's NULL, we just resume based on position
    rotation_participant_id UUID REFERENCES rotation_participants (id) ON DELETE SET NULL,
    shift_start TIMESTAMP WITH TIME ZONE NOT NULL
);

-- +migrate Down

DROP TABLE rotation_state;
