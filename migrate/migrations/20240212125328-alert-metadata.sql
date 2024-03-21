-- +migrate Up
CREATE TABLE alerts_metadata (
	alert_id integer NOT NULL,
	id uuid DEFAULT gen_random_uuid() NOT NULL,
	metadata jsonb,
	CONSTRAINT alerts_metadata_alert_id_fkey FOREIGN KEY (alert_id) REFERENCES alerts(id) ON DELETE CASCADE,
	CONSTRAINT alerts_metadata_pkey PRIMARY KEY (id, alert_id),
	CONSTRAINT alerts_metadata_id UNIQUE (alert_id)
);

-- +migrate Down
DROP TABLE alerts_metadata;
