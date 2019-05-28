
-- +migrate Up

-- Don't allow alert state change or creation, service EP change, or state changes
-- while changing everything.
LOCK services, escalation_policy_state IN EXCLUSIVE MODE;


-- Make alert_id the primary key, require
ALTER TABLE escalation_policy_state
    ADD PRIMARY KEY (alert_id),
    ADD COLUMN service_id UUID REFERENCES services (id) ON DELETE CASCADE,
    DROP CONSTRAINT escalation_policy_state_alert_id_escalation_policy_id_key;


CREATE INDEX idx_escalation_policy_state_policy_ids ON escalation_policy_state (escalation_policy_id, service_id);

-- Set service_id col, then enforce NOT NULL.
UPDATE escalation_policy_state
SET service_id = a.service_id
FROM alerts a
WHERE a.id = alert_id;

ALTER TABLE escalation_policy_state
    ALTER service_id SET NOT NULL;


-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_set_ep_state_svc_id_on_insert() RETURNS TRIGGER AS
$$
BEGIN
    SELECT service_id INTO NEW.service_id
    FROM alerts
    WHERE id = NEW.alert_id;

    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

CREATE TRIGGER trg_10_set_ep_state_svc_id_on_insert
BEFORE INSERT ON escalation_policy_state
FOR EACH ROW
WHEN (NEW.service_id ISNULL)
EXECUTE PROCEDURE fn_set_ep_state_svc_id_on_insert();


-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_clear_ep_state_on_svc_ep_change() RETURNS TRIGGER AS
$$
BEGIN
    DELETE FROM escalation_policy_state
    WHERE service_id = NEW.id;

    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

CREATE TRIGGER trg_10_clear_ep_state_on_svc_ep_change
BEFORE UPDATE ON services
FOR EACH ROW
WHEN (OLD.escalation_policy_id != NEW.escalation_policy_id)
EXECUTE PROCEDURE fn_clear_ep_state_on_svc_ep_change();


-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_clear_ep_state_on_alert_close() RETURNS TRIGGER AS
$$
BEGIN
    DELETE FROM escalation_policy_state
    WHERE alert_id = NEW.id;

    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

CREATE TRIGGER trg_10_clear_ep_state_on_alert_close
AFTER UPDATE ON alerts
FOR EACH ROW
WHEN (OLD.status != NEW.status AND NEW.status = 'closed')
EXECUTE PROCEDURE fn_clear_ep_state_on_alert_close();

ALTER TABLE escalation_policy_state SET (fillfactor = 85);

-- +migrate Down

DROP INDEX idx_escalation_policy_state_policy_ids;
ALTER TABLE escalation_policy_state RESET (fillfactor);

DROP TRIGGER trg_10_clear_ep_state_on_alert_close ON alerts;
DROP TRIGGER trg_10_clear_ep_state_on_svc_ep_change ON services;
DROP TRIGGER trg_10_set_ep_state_svc_id_on_insert ON escalation_policy_state;

DROP FUNCTION fn_set_ep_state_svc_id_on_insert();
DROP FUNCTION fn_clear_ep_state_on_svc_ep_change();
DROP FUNCTION fn_clear_ep_state_on_alert_close();

ALTER TABLE escalation_policy_state
    DROP CONSTRAINT escalation_policy_state_pkey,
    DROP COLUMN service_id,
    ADD UNIQUE(alert_id, escalation_policy_id);
