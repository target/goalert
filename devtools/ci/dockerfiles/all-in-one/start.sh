#!/bin/sh
set -e

./usr/local/bin/docker-entrypoint.sh postgres &> /var/log/postgres.log &

export DB_URL=postgres://postgres@127.0.0.1/postgres?sslmode=disable

./bin/waitfor "$DB_URL"

if test -f /bin/init.sql
then
  echo Seeding DB with demo data...
  psql -d "$DB_URL" < /bin/init.sql > /dev/null
  mv /bin/init.sql /bin/init.sql.applied
fi

exec /bin/goalert --db-url "$DB_URL" --log-requests -l ":8081"
