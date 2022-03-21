SELECT col.table_name,
    col.column_name,
    col.data_type,
    col.ordinal_position
FROM information_schema.columns col
    JOIN information_schema.tables t ON t.table_catalog = col.table_catalog
    AND t.table_schema = col.table_schema
    AND t.table_name = col.table_name
    AND t.table_type = 'BASE TABLE'
WHERE col.table_catalog = current_database()
    AND col.table_schema = 'public'
