-- just for type info
CREATE TABLE pg_stat_activity (
    state TEXT,
    XACT_START timestamptz NOT NULL,
    application_name TEXT
);
