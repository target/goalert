
-- +migrate Up
DROP TABLE process_alerts;

-- +migrate Down
CREATE TABLE process_alerts (
    alert_id bigint NOT NULL,
    client_id uuid,
    deadline timestamp with time zone,
    last_processed timestamp with time zone
);

ALTER TABLE ONLY process_alerts
    ADD CONSTRAINT process_alerts_pkey PRIMARY KEY (alert_id);

CREATE INDEX process_alerts_oldest_first ON public.process_alerts USING btree (last_processed NULLS FIRST);

CREATE TRIGGER trg_disable_old_alert_processing BEFORE INSERT ON public.process_alerts FOR EACH STATEMENT EXECUTE PROCEDURE fn_disable_inserts();

ALTER TABLE ONLY process_alerts
    ADD CONSTRAINT process_alerts_alert_id_fkey FOREIGN KEY (alert_id) REFERENCES alerts(id) ON DELETE CASCADE;
