
-- +migrate Up
DROP TABLE throttle;
-- +migrate Down
CREATE TABLE throttle (
    action enum_throttle_type NOT NULL,
    client_id uuid,
    last_action_time timestamp with time zone DEFAULT now() NOT NULL
);

ALTER TABLE ONLY throttle
    ADD CONSTRAINT throttle_pkey PRIMARY KEY (action);

INSERT INTO throttle (action)
VALUES ('notifications');
