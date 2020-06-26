#!/bin/sh
set -e

export PGDATA=/var/lib/postgresql/data PGUSER=postgres DB_URL=postgresql://postgres@

if test -f /bin/init.sql
then
  mkdir -p ${PGDATA} /run/postgresql /var/log/postgresql &&\
  chown postgres ${PGDATA} /run/postgresql /var/log/postgresql &&\
  su postgres -s /bin/sh -c "initdb $PGDATA" &&\
  echo "host all  all    0.0.0.0/0  md5" >> $PGDATA/pg_hba.conf &&\
  echo "listen_addresses='*'" >> $PGDATA/postgresql.conf &&\
  echo "fsync = off" >> $PGDATA/postgresql.conf &&\
  echo "full_page_writes = off" >> $PGDATA/postgresql.conf
fi

su postgres -s /bin/sh -c "pg_ctl start -w -l /var/log/postgresql/server.log"

if test -f /bin/init.sql
then
  echo Seeding DB with demo data...
  psql -d "$DB_URL" < /bin/init.sql > /dev/null
  mv /bin/init.sql /bin/init.sql.applied
fi

exec /bin/goalert --db-url "$DB_URL" --log-requests
