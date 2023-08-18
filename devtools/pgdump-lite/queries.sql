-- name: ListExtensions :many
SELECT
    extname::text AS ext_name,
    n.nspname::text AS schema_name
FROM
    pg_catalog.pg_extension e
    JOIN pg_catalog.pg_namespace n ON n.oid = e.extnamespace
        AND n.nspname != 'pg_catalog'
    ORDER BY
        n.nspname,
        extname;

-- name: ListEnums :many
SELECT
    n.nspname::text AS schema_name,
    t.typname::text AS enum_name,
    string_agg(e.enumlabel, ',' ORDER BY e.enumlabel) AS enum_values
FROM
    pg_catalog.pg_type t
    JOIN pg_catalog.pg_namespace n ON n.oid = t.typnamespace
    JOIN pg_catalog.pg_enum e ON e.enumtypid = t.oid
WHERE (t.typrelid = 0
    OR (
        SELECT
            c.relkind = 'c'
        FROM
            pg_catalog.pg_class c
        WHERE
            c.oid = t.typrelid))
AND NOT EXISTS (
    SELECT
        1
    FROM
        pg_catalog.pg_type el
    WHERE
        el.oid = t.typelem
        AND el.typarray = t.oid)
AND n.nspname NOT IN ('pg_catalog', 'information_schema')
GROUP BY
    n.nspname,
    t.typname
ORDER BY
    n.nspname,
    t.typname;

-- name: ListFunctions :many
SELECT
    n.nspname::text AS schema_name,
    p.proname::text AS function_name,
    pg_get_functiondef(p.oid) AS func_def
FROM
    pg_catalog.pg_proc p
    JOIN pg_catalog.pg_namespace n ON p.pronamespace = n.oid
    LEFT JOIN pg_catalog.pg_depend d ON p.oid = d.objid
        AND d.deptype = 'e'
    LEFT JOIN pg_catalog.pg_extension e ON d.refobjid = e.oid
WHERE
    n.nspname NOT IN ('pg_catalog', 'information_schema')
    AND p.prokind = 'f'
    AND d.objid IS NULL
ORDER BY
    n.nspname,
    p.proname;

-- name: ListColumns :many
SELECT
    n.nspname::text AS schema_name,
    c.relname::text AS table_name,
    a.attnum AS column_number,
    a.attname::text AS column_name,
    pg_catalog.format_type(a.atttypid, a.atttypmod) AS column_type,
    coalesce(pg_get_expr(d.adbin, d.adrelid), '')::text AS column_default,
    a.attnotnull AS not_null
FROM
    pg_catalog.pg_attribute a
    JOIN pg_catalog.pg_class c ON a.attnum > 0
        AND a.attrelid = c.oid
    JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid
    LEFT JOIN pg_catalog.pg_attrdef d ON a.attrelid = d.adrelid
        AND a.attnum = d.adnum
WHERE
    n.nspname NOT IN ('pg_catalog', 'information_schema')
    AND c.relkind = 'r'
    AND NOT a.attisdropped
ORDER BY
    n.nspname,
    c.relname,
    a.attname;

-- name: ListCheckConstraints :many
SELECT
    n.nspname::text AS schema_name,
    c.relname::text AS table_name,
    cc.conname::text AS constraint_name,
    pg_get_constraintdef(cc.oid) AS check_clause
FROM
    pg_catalog.pg_constraint cc
    JOIN pg_catalog.pg_class c ON cc.conrelid = c.oid
    JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid
WHERE
    cc.contype = 'c'
ORDER BY
    n.nspname,
    c.relname,
    cc.conname;

-- name: ListSequences :many
SELECT
    n.nspname::text AS schema_name,
    s.relname::text AS sequence_name,
    seq.start_value,
    seq.increment_by AS increment,
    seq.min_value AS min_value,
    seq.max_value AS max_value,
    seq.cache_size AS
    CACHE,
    coalesce((
        SELECT
            tn.nspname::text
        FROM pg_catalog.pg_namespace tn
        WHERE
            tn.oid = tc.relnamespace), '')::text AS table_schema,
    coalesce(tc.relname, '')::text AS table_name,
    coalesce(a.attname, '')::text AS column_name
FROM
    pg_catalog.pg_class s
    JOIN pg_catalog.pg_namespace n ON s.relnamespace = n.oid
    JOIN pg_catalog.pg_sequences seq ON n.nspname = seq.schemaname
        AND s.relname = seq.sequencename
    LEFT JOIN pg_catalog.pg_depend d ON s.oid = d.objid
        AND d.deptype = 'a'
    LEFT JOIN pg_catalog.pg_attribute a ON a.attnum = d.refobjsubid
        AND a.attrelid = d.refobjid
    LEFT JOIN pg_catalog.pg_class tc ON tc.oid = d.refobjid
WHERE
    s.relkind = 'S'
ORDER BY
    n.nspname,
    s.relname;

-- name: ListConstraints :many
SELECT
    n.nspname::text AS schema_name,
    t.relname::text AS table_name,
    c.conname::text AS constraint_name,
    pg_catalog.pg_get_constraintdef(c.oid, TRUE) AS constraint_definition
FROM
    pg_catalog.pg_constraint c
    JOIN pg_catalog.pg_class t ON c.conrelid = t.oid
    JOIN pg_catalog.pg_namespace n ON n.oid = t.relnamespace
WHERE
    t.relkind = 'r'
    AND n.nspname NOT IN ('pg_catalog', 'information_schema')
ORDER BY
    n.nspname,
    t.relname,
    c.conname;

-- name: ListIndexes :many
SELECT
    n.nspname::text AS schema_name,
    t.relname::text AS table_name,
    i.indexname::text AS index_name,
    i.indexdef::text AS index_definition
FROM
    pg_catalog.pg_indexes i
    JOIN pg_catalog.pg_class t ON t.relname = i.tablename
    JOIN pg_catalog.pg_namespace n ON n.oid = t.relnamespace
WHERE
    n.nspname NOT IN ('pg_catalog', 'information_schema')
ORDER BY
    n.nspname,
    t.relname,
    i.indexname;

-- name: ListTriggers :many
SELECT
    n.nspname::text AS schema_name,
    t.relname::text AS table_name,
    trg.tgname::text AS trigger_name,
    pg_catalog.pg_get_triggerdef(trg.oid) AS trigger_definition
FROM
    pg_catalog.pg_trigger trg
    JOIN pg_catalog.pg_class t ON t.oid = trg.tgrelid
    JOIN pg_catalog.pg_namespace n ON n.oid = t.relnamespace
WHERE
    NOT trg.tgisinternal
    AND n.nspname NOT IN ('pg_catalog', 'information_schema')
ORDER BY
    n.nspname,
    t.relname,
    trg.tgname;

