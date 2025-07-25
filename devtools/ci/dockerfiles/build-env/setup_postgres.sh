#!/bin/sh
set -e

for VERSION in $*; do
    echo "Initializing PostgreSQL $VERSION"
    PGDATA=/var/lib/postgresql/$VERSION/data
    mkdir -p $PGDATA /run/postgresql /var/log/postgresql/$VERSION

    chown postgres $PGDATA /run/postgresql /var/log/postgresql/$VERSION
    su postgres -c "/usr/lib/postgresql/$VERSION/bin/initdb $PGDATA"
    echo "host all  all    0.0.0.0/0  md5" >>$PGDATA/pg_hba.conf
    echo "listen_addresses='*'" >>$PGDATA/postgresql.conf
    echo "fsync = off" >>$PGDATA/postgresql.conf
    echo "full_page_writes = off" >>$PGDATA/postgresql.conf
done
chown -R postgres:postgres /var/lib/postgresql
