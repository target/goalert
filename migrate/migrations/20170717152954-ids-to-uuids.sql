-- +migrate Up

-- create new tables -- makes migration/testing easier than trying to rename, alter AND recreate
CREATE TABLE escalation_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    repeat INT NOT NULL DEFAULT 0
);

CREATE TABLE escalation_policy_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    delay INT NOT NULL DEFAULT 1,
    step_number INT NOT NULL DEFAULT -1,
    escalation_policy_id UUID NOT NULL REFERENCES escalation_policies (id) ON DELETE CASCADE,
    UNIQUE (escalation_policy_id, step_number)
);

CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    escalation_policy_id UUID NOT NULL REFERENCES escalation_policies (id)
);

-- copy data over
INSERT INTO escalation_policies (id, name, description, repeat)
SELECT id::UUID, name, description, repeat FROM escalation_policy;

INSERT INTO escalation_policy_steps (id, delay, step_number, escalation_policy_id)
SELECT id::UUID, delay, step_number, escalation_policy_id::UUID
FROM escalation_policy_step;

INSERT INTO services (id, name, description, escalation_policy_id)
SELECT id::UUID, name, description, escalation_policy_id::UUID
FROM service;

-- drop views
DROP VIEW needs_notification_sent, on_call_alert_users, alert_escalation_levels;

ALTER TABLE alerts
    DROP CONSTRAINT alerts_service_id_fkey,
    ALTER service_id TYPE UUID USING service_id::UUID,
    ADD CONSTRAINT alerts_services_id_fkey FOREIGN KEY (service_id) REFERENCES services (id) ON DELETE CASCADE;

ALTER TABLE escalation_policy_actions
    DROP CONSTRAINT escalation_policy_actions_escalation_policy_step_id_fkey,
    ALTER escalation_policy_step_id TYPE UUID USING escalation_policy_step_id::UUID,
    ADD CONSTRAINT escalation_policy_actions_escalation_policy_step_id_fkey FOREIGN KEY (escalation_policy_step_id) REFERENCES escalation_policy_steps (id) ON DELETE CASCADE;

ALTER TABLE integration_keys
    DROP CONSTRAINT integration_keys_service_id_fkey,
    ALTER service_id TYPE UUID USING service_id::UUID,
    ADD CONSTRAINT integration_keys_services_id_fkey FOREIGN KEY (service_id) REFERENCES services (id) ON DELETE CASCADE;

CREATE VIEW alert_escalation_levels AS
    SELECT
        alerts.id AS alert_id,
        count(step.id) AS levels,
        alerts.escalation_level::bigint % count(step.id) AS relative_level
    FROM
        alerts,
        escalation_policy_steps step,
        services s
    WHERE step.escalation_policy_id = s.escalation_policy_id
        AND s.id = alerts.service_id
    GROUP BY alerts.id;

CREATE VIEW on_call_alert_users AS
    WITH alert_users AS (
        SELECT act.user_id,
            act.schedule_id,
            a.id AS alert_id,
            a.status,
            a.escalation_level
        FROM alerts a,
            services s,
            alert_escalation_levels lvl,
            escalation_policy_steps step,
            escalation_policy_actions act
        WHERE s.id = a.service_id
            AND lvl.alert_id = a.id
            AND step.escalation_policy_id = s.escalation_policy_id
            AND step.step_number = lvl.relative_level
            AND a.status <> 'closed'::enum_alert_status
            AND act.escalation_policy_step_id = step.id
        GROUP BY act.user_id, act.schedule_id, a.id
    )
    SELECT DISTINCT au.alert_id,
        au.status,
        CASE
            WHEN au.user_id IS NULL THEN oc.user_id
            ELSE au.user_id
        END AS user_id,
        au.escalation_level
    FROM alert_users au
    LEFT JOIN on_call oc ON au.schedule_id = oc.schedule_id;

CREATE VIEW needs_notification_sent AS
    SELECT DISTINCT
        cs.alert_id,
        nr.contact_method_id,
        cm.type,
        cm.value,
        a.description,
        s.name AS service_name,
        nr.id AS notification_rule_id,
        cs.escalation_level,
        cs.cycle_id
    FROM
        user_notification_cycle_state cs,
        alerts a,
        user_contact_methods cm,
        user_notification_rules nr,
        services s
    WHERE a.id = cs.alert_id
        AND a.status = 'triggered'::enum_alert_status
        AND cs.escalation_level = a.escalation_level
        AND cm.id = nr.contact_method_id
        AND nr.id = cs.notification_rule_id
        AND s.id = a.service_id
        AND cs.pending
        AND NOT cs.future;

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION set_escalation_policy_step() RETURNS TRIGGER AS
    $$
        BEGIN
            SELECT count(step_number) INTO NEW.step_number FROM escalation_policy_steps WHERE escalation_policy_id = NEW.escalation_policy_id;
            RETURN NEW;
        END;
    $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION incr_escalation_policy_steps_on_delete() RETURNS TRIGGER AS
    $$
        BEGIN
            UPDATE escalation_policy_steps
            SET step_number = step_number-1
            WHERE escalation_policy_id = OLD.escalation_policy_id
                AND step_number > OLD.step_number;

            RETURN OLD;
        END;
    $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION escalate_alerts() RETURNS VOID AS
    $$
        BEGIN
            UPDATE alerts a
                SET escalation_level = escalation_level + 1, last_escalation = now()
                FROM services s, escalation_policy_steps step, alert_escalation_levels lvl, escalation_policies e
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


    

CREATE TRIGGER incr_escalation_policy_steps_on_delete
    AFTER DELETE ON escalation_policy_steps
    FOR EACH ROW
    EXECUTE PROCEDURE incr_escalation_policy_steps_on_delete();


CREATE TRIGGER set_escalation_policy_step_on_insert
    BEFORE INSERT ON escalation_policy_steps
    FOR EACH ROW
    EXECUTE PROCEDURE set_escalation_policy_step();


DROP TABLE service, escalation_policy_step, escalation_policy;

-- +migrate Down

-- start by creating old tables
CREATE TABLE escalation_policy (
    id text DEFAULT (gen_random_uuid())::text NOT NULL PRIMARY KEY,
    description text DEFAULT ''::text NOT NULL,
    name text UNIQUE,
    repeat integer DEFAULT 0 NOT NULL
);
CREATE TABLE service (
    id text PRIMARY KEY NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    name text NOT NULL UNIQUE,
    escalation_policy_id text REFERENCES escalation_policy (id)
);

CREATE TABLE escalation_policy_step (
    id text DEFAULT (gen_random_uuid())::text NOT NULL PRIMARY KEY,
    delay integer DEFAULT 1 NOT NULL,
    step_number integer,
    escalation_policy_id text REFERENCES escalation_policy (id) ON DELETE CASCADE,
    UNIQUE(step_number, escalation_policy_id)
);

-- drop views
DROP VIEW needs_notification_sent, on_call_alert_users, alert_escalation_levels;

-- copy data over
INSERT INTO escalation_policy (id, name, description, repeat)
SELECT id::text, name, description, repeat
FROM escalation_policies;

INSERT INTO escalation_policy_step (id, delay, step_number, escalation_policy_id)
SELECT id::text, delay, step_number, escalation_policy_id::text
FROM escalation_policy_steps;

INSERT INTO service (id, name, description, escalation_policy_id)
SELECT id::text, name, description, escalation_policy_id::text
FROM services;

ALTER TABLE alerts
    DROP CONSTRAINT alerts_services_id_fkey,
    ALTER service_id TYPE TEXT USING service_id::TEXT,
    ADD CONSTRAINT alerts_service_id_fkey FOREIGN KEY (service_id) REFERENCES service (id) ON DELETE CASCADE;

ALTER TABLE escalation_policy_actions
    DROP CONSTRAINT escalation_policy_actions_escalation_policy_step_id_fkey,
    ALTER escalation_policy_step_id TYPE TEXT USING escalation_policy_step_id::TEXT,
    ADD CONSTRAINT escalation_policy_actions_escalation_policy_step_id_fkey FOREIGN KEY (escalation_policy_step_id) REFERENCES escalation_policy_step (id) ON DELETE CASCADE;

ALTER TABLE integration_keys
    DROP CONSTRAINT integration_keys_services_id_fkey,
    ALTER service_id TYPE TEXT USING service_id::TEXT,
    ADD CONSTRAINT integration_keys_service_id_fkey FOREIGN KEY (service_id) REFERENCES service (id) ON DELETE CASCADE;


-- restore old views
CREATE VIEW alert_escalation_levels AS SELECT alerts.id AS alert_id,
    count(step.id) AS levels,
    ((alerts.escalation_level)::bigint % count(step.id)) AS relative_level
   FROM alerts,
    escalation_policy_step step,
    service
  WHERE ((step.escalation_policy_id = service.escalation_policy_id) AND (service.id = alerts.service_id))
  GROUP BY alerts.id;

CREATE VIEW on_call_alert_users AS WITH alert_users AS (
         SELECT act.user_id,
            act.schedule_id,
            a.id AS alert_id,
            a.status,
            a.escalation_level
           FROM alerts a,
            service s,
            alert_escalation_levels lvl,
            escalation_policy_step step,
            escalation_policy_actions act
          WHERE ((s.id = a.service_id) AND (lvl.alert_id = a.id) AND (step.escalation_policy_id = s.escalation_policy_id) AND (step.step_number = lvl.relative_level) AND (a.status <> 'closed'::enum_alert_status) AND (act.escalation_policy_step_id = step.id))
          GROUP BY act.user_id, act.schedule_id, a.id
        )
 SELECT DISTINCT au.alert_id,
    au.status,
        CASE
            WHEN (au.user_id IS NULL) THEN oc.user_id
            ELSE au.user_id
        END AS user_id,
    au.escalation_level
   FROM (alert_users au
     LEFT JOIN on_call oc ON ((au.schedule_id = oc.schedule_id)));

CREATE VIEW needs_notification_sent AS
 SELECT DISTINCT cs.alert_id,
    nr.contact_method_id,
    cm.type,
    cm.value,
    a.description,
    s.name AS service_name,
    nr.id AS notification_rule_id,
    cs.escalation_level,
    cs.cycle_id
   FROM user_notification_cycle_state cs,
    alerts a,
    user_contact_methods cm,
    user_notification_rules nr,
    service s
  WHERE ((a.id = cs.alert_id) AND (a.status = 'triggered'::enum_alert_status) AND (cs.escalation_level = a.escalation_level) AND (cm.id = nr.contact_method_id) AND (nr.id = cs.notification_rule_id) AND (s.id = a.service_id) AND cs.pending AND (NOT cs.future));

-- restore old function code

-- +migrate StatementBegin

CREATE OR REPLACE FUNCTION set_escalation_policy_step() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
        BEGIN
            SELECT count(step_number) INTO NEW.step_number FROM escalation_policy_step WHERE escalation_policy_id = NEW.escalation_policy_id;
            RETURN NEW;
        END;
    $$;
-- +migrate StatementEnd
-- +migrate StatementBegin

CREATE OR REPLACE FUNCTION incr_escalation_policy_steps_on_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
        BEGIN
            UPDATE escalation_policy_step
            SET step_number = step_number-1
            WHERE escalation_policy_id = OLD.escalation_policy_id
                AND step_number > OLD.step_number;

            RETURN OLD;
        END;
    $$;
-- +migrate StatementEnd
-- +migrate StatementBegin

CREATE OR REPLACE FUNCTION escalate_alerts() RETURNS void
    LANGUAGE plpgsql
    AS $$
    BEGIN
        UPDATE alerts a
            SET escalation_level = escalation_level + 1, last_escalation = now()
            FROM service s, escalation_policy_step step, alert_escalation_levels lvl, escalation_policy e
            WHERE (last_escalation + (step.delay::TEXT||' minutes')::interval) < now()
                AND a.status = 'triggered'::enum_alert_status
                AND s.id = a.service_id
                AND step.escalation_policy_id = s.escalation_policy_id
                AND lvl.alert_id = a.id
                AND step.step_number = (a.escalation_level % lvl.levels)
                AND e.id = s.escalation_policy_id
                AND (e.repeat = -1 OR (escalation_level+1) / lvl.levels <= e.repeat);
    END;
    $$;
-- +migrate StatementEnd

CREATE TRIGGER incr_escalation_policy_steps_on_delete AFTER DELETE ON escalation_policy_step FOR EACH ROW EXECUTE PROCEDURE incr_escalation_policy_steps_on_delete();
CREATE TRIGGER set_escalation_policy_step_on_insert BEFORE INSERT ON escalation_policy_step FOR EACH ROW EXECUTE PROCEDURE set_escalation_policy_step();

DROP TABLE escalation_policy_steps, services, escalation_policies;


