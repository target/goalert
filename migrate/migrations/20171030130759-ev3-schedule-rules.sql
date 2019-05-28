
-- +migrate Up

CREATE TABLE schedule_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id UUID NOT NULL REFERENCES schedules (id) ON DELETE CASCADE,
    sunday BOOLEAN NOT NULL DEFAULT true,
    monday BOOLEAN NOT NULL DEFAULT true,
    tuesday BOOLEAN NOT NULL DEFAULT true,
    wednesday BOOLEAN NOT NULL DEFAULT true,
    thursday BOOLEAN NOT NULL DEFAULT true,
    friday BOOLEAN NOT NULL DEFAULT true,
    saturday BOOLEAN NOT NULL DEFAULT true,    
    start_time TIME NOT NULL DEFAULT '00:00:00',
    end_time TIME NOT NULL DEFAULT '23:59:59',

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),

    tgt_user_id UUID REFERENCES users (id) ON DELETE CASCADE,
    tgt_rotation_id UUID REFERENCES rotations (id) ON DELETE CASCADE,

    CHECK(
        (tgt_user_id IS NULL AND tgt_rotation_id IS NOT NULL)
        OR
        (tgt_user_id IS NOT NULL AND tgt_rotation_id IS NULL)
    )
);

-- +migrate Down

DROP TABLE schedule_rules;
