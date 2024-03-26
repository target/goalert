#!/bin/sh
set -e

export PGDATA=/var/lib/postgresql/data PGUSER=postgres DB_URL=postgresql://postgres@

if ! test -f /init-db
then
  mkdir -p ${PGDATA} /run/postgresql /var/log/postgresql &&\
  chown postgres ${PGDATA} /run/postgresql /var/log/postgresql &&\
  su postgres -s /bin/sh -c "initdb --auth-local trust -N $PGDATA" &&\
  echo "listen_addresses=''" >> $PGDATA/postgresql.conf &&\
  echo "fsync = off" >> $PGDATA/postgresql.conf &&\
  echo "full_page_writes = off" >> $PGDATA/postgresql.conf
  touch /init-db
fi

su postgres -s /bin/sh -c "pg_ctl start -w -l /var/log/postgresql/server.log"

if ! test -f /init-data
then
  echo Seeding DB with demo data...
  if [ "$SKIP_SEED" = "1" ]; then
    /bin/goalert migrate --db-url "$DB_URL" # run migrations, but don't seed
    /bin/goalert add-user --admin --user admin --pass admin123 --db-url "$DB_URL"
  else
    /bin/resetdb -with-rand-data -admin-id=00000000-0000-0000-0000-000000000001 -db-url "$DB_URL" -skip-drop
    /bin/goalert add-user --user-id=00000000-0000-0000-0000-000000000001 --user admin --pass admin123 --db-url "$DB_URL"
  fi

  /bin/goalert add-user --user user --pass user1234 --db-url "$DB_URL"
  touch /init-data
fi

exec /bin/goalert --db-url "$DB_URL" --log-requests
