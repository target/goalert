-- just for type info
CREATE TABLE pg_stat_activity(
    state text,
    XACT_START timestamptz NOT NULL,
    application_name text
);

