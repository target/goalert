SELECT src.relname,
    dst.relname
FROM pg_catalog.pg_constraint con
    JOIN pg_namespace ns ON ns.nspname = 'public'
    AND ns.oid = con.connamespace
    JOIN pg_class src ON src.oid = con.conrelid
    JOIN pg_class dst ON dst.oid = con.confrelid
WHERE con.contype = 'f'
    AND NOT con.condeferrable
