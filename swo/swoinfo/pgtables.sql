-- pg_catalog tables used by SWO
CREATE TABLE pg_catalog.pg_namespace (
    oid oid NOT NULL,
    nspname NAME NOT NULL,
    nspowner oid NOT NULL,
    nspacl aclitem [ ]
);

CREATE TABLE pg_catalog.pg_class (
    oid oid NOT NULL,
    relname NAME NOT NULL,
    relnamespace oid NOT NULL,
    reltype oid NOT NULL,
    reloftype oid NOT NULL,
    relowner oid NOT NULL,
    relam oid NOT NULL,
    relfilenode oid NOT NULL,
    reltablespace oid NOT NULL,
    relpages INTEGER NOT NULL,
    reltuples REAL NOT NULL,
    relallvisible INTEGER NOT NULL,
    reltoastrelid oid NOT NULL,
    relhasindex BOOLEAN NOT NULL,
    relisshared BOOLEAN NOT NULL,
    relpersistence "char" NOT NULL,
    relkind "char" NOT NULL,
    relnatts SMALLINT NOT NULL,
    relchecks SMALLINT NOT NULL,
    relhasrules BOOLEAN NOT NULL,
    relhastriggers BOOLEAN NOT NULL,
    relhassubclass BOOLEAN NOT NULL,
    relrowsecurity BOOLEAN NOT NULL,
    relforcerowsecurity BOOLEAN NOT NULL,
    relispopulated BOOLEAN NOT NULL,
    relreplident "char" NOT NULL,
    relispartition BOOLEAN NOT NULL,
    relrewrite oid NOT NULL,
    relfrozenxid xid NOT NULL,
    relminmxid xid NOT NULL,
    relacl aclitem [ ],
    reloptions text [ ] COLLATE pg_catalog. "C",
    relpartbound pg_node_tree COLLATE pg_catalog. "C"
);

CREATE TABLE pg_catalog.pg_constraint (
    oid oid NOT NULL,
    conname NAME NOT NULL,
    connamespace oid NOT NULL,
    contype "char" NOT NULL,
    condeferrable BOOLEAN NOT NULL,
    condeferred BOOLEAN NOT NULL,
    convalidated BOOLEAN NOT NULL,
    conrelid oid NOT NULL,
    contypid oid NOT NULL,
    conindid oid NOT NULL,
    conparentid oid NOT NULL,
    confrelid oid NOT NULL,
    confupdtype "char" NOT NULL,
    confdeltype "char" NOT NULL,
    confmatchtype "char" NOT NULL,
    conislocal BOOLEAN NOT NULL,
    coninhcount INTEGER NOT NULL,
    connoinherit BOOLEAN NOT NULL,
    conkey SMALLINT [ ],
    confkey SMALLINT [ ],
    conpfeqop oid [ ],
    conppeqop oid [ ],
    conffeqop oid [ ],
    conexclop oid [ ],
    conbin pg_node_tree COLLATE pg_catalog. "C"
);

-- just for type info
CREATE TABLE pg_stat_activity (
    state TEXT,
    XACT_START timestamptz NOT NULL,
    application_name TEXT
);

CREATE SCHEMA information_schema;

CREATE TABLE information_schema.columns (
    table_name TEXT NOT NULL,
    column_name TEXT NOT NULL,
    data_type TEXT NOT NULL,
    ordinal_position INTEGER NOT NULL
);

CREATE TABLE information_schema.tables ();

CREATE TABLE information_schema.sequences (sequence_name TEXT NOT NULL);
