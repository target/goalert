#!/bin/sh
set -ex
start_postgres

DB_URL=postgresql://postgres@

./bin/resetdb -with-rand-data -admin-id=00000000-0000-0000-0000-000000000000 -db-url "$DB_URL" -skip-drop
./bin/goalert add-user --user-id=00000000-0000-0000-0000-000000000000 --user admin --pass admin123 --db-url "$DB_URL"

pg_dump -d "$DB_URL" --exclude-table-data=rotation_state  >/init.sql

stop_postgres
