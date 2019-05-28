
-- +migrate Up

CREATE EXTENSION IF NOT EXISTS pgcrypto; 

CREATE TYPE enum_alert_status as ENUM (
    'triggered',
    'active',
    'closed'
);

CREATE TYPE enum_alert_source as ENUM (
    'grafana',
    'manual'
);

CREATE TYPE enum_alert_log_event as ENUM (
    'created',
    'reopened',
    'status_changed',
    'assignment_changed',
    'escalated',
    'closed'
);

CREATE TABLE alerts (
	id BIGSERIAL PRIMARY KEY,
    description TEXT NOT NULL,
    service_id TEXT REFERENCES service (id) ON DELETE CASCADE,
    source enum_alert_source NOT NULL DEFAULT 'manual'::enum_alert_source,
    status enum_alert_status NOT NULL DEFAULT 'triggered'::enum_alert_status,

    escalation_level INT NOT NULL DEFAULT 0,
    last_escalation TIMESTAMP DEFAULT now()
);

CREATE TABLE alert_logs (
    id BIGSERIAL PRIMARY KEY,
    alert_id BIGINT REFERENCES alerts (id) ON DELETE CASCADE,
    timestamp TIMESTAMP DEFAULT now(),
    event enum_alert_log_event NOT NULL,
    message TEXT NOT NULL
);

CREATE VIEW alert_escalation_levels AS
    SELECT alerts.id AS alert_id, 
            count(step.id) AS levels,
            ((alerts.escalation_level + 1) % count(step.id)) as relative_level 
        FROM alerts,escalation_policy_step step,service
        WHERE step.escalation_policy_id = service.escalation_policy_id
            AND service.id = alerts.service_id
        GROUP BY alerts.id;

INSERT INTO alerts (description, service_id, last_escalation, status)
    SELECT description, service_id, created_at, 'active'::enum_alert_status FROM incident;


-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION log_alert_status_changed_insert() RETURNS TRIGGER AS
    $$
        BEGIN
            IF NEW.status = 'closed'::enum_alert_status THEN
                INSERT INTO alert_logs (alert_id, event, message) VALUES (
                    NEW.id, 'closed'::enum_alert_log_event, 'Closed'
                );
            ELSIF OLD.status = 'closed'::enum_alert_status THEN
                INSERT INTO alert_logs (alert_id, event, message) VALUES (
                    NEW.id, 'reopened'::enum_alert_log_event, 'Reopened as '||NEW.status::TEXT
                );
            ELSE
                INSERT INTO alert_logs (alert_id, event, message) VALUES (
                    NEW.id, 'status_changed'::enum_alert_log_event, 'Status updated from '||OLD.status::TEXT||' to '||NEW.status::TEXT
                );
            END IF;
            RETURN NEW;
        END;
    $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION log_alert_creation_insert() RETURNS TRIGGER AS
    $$
        BEGIN
            INSERT INTO alert_logs (alert_id, event, message) VALUES (
                NEW.id, 'created'::enum_alert_log_event, 'Created via: '||NEW.source::TEXT
            );
            RETURN NEW;
        END;
    $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION escalate_alerts() RETURNS VOID AS
    $$
    BEGIN
        UPDATE alerts a
            SET escalation_level = escalation_level + 1, last_escalation = now()
            FROM service s, escalation_policy_step step, alert_escalation_levels lvl, escalation_policy e
            WHERE (last_escalation + (step.delay::TEXT||' minutes')::interval) < now()
                AND a.status = 'triggered'::enum_alert_status
                AND s.id = a.service_id
                AND step.escalation_policy_id = s.escalation_policy_id
                AND lvl.alert_id = a.id
                AND step.step_number = ((a.escalation_level + 1) % lvl.levels)
                AND e.id = s.escalation_policy_id
                AND (e.repeat = -1 OR escalation_level / lvl.levels < e.repeat);
    END;
    $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

CREATE TRIGGER log_alert_status_changed
    AFTER UPDATE ON alerts
    FOR EACH ROW
    WHEN (OLD.status IS DISTINCT FROM NEW.status)
    EXECUTE PROCEDURE log_alert_status_changed_insert();

CREATE TRIGGER log_alert_creation
    AFTER INSERT ON alerts
    FOR EACH ROW
    EXECUTE PROCEDURE log_alert_creation_insert();

-- +migrate Down

DROP VIEW alert_steps;

DROP TABLE alert_logs;
DROP TYPE enum_alert_log_event;

DROP TABLE alerts;
DROP TYPE enum_alert_source;
DROP TYPE enum_alert_status;

DROP FUNCTION log_alert_status_changed_insert();
DROP FUNCTION log_alert_creation_insert();
