#!/bin/sh
set -e

# check if any PostgreSQL version is already running
if [ -f /var/lib/postgresql/.version ]; then
    VERSION=$(cat /var/lib/postgresql/.version)
    if [ -f /var/lib/postgresql/$VERSION/data/postmaster.pid ]; then
        stop_postgres
    fi
fi

VERSION=$1

if [ -z "$VERSION" ]; then
    VERSION=13 # Default to PostgreSQL 13 for compatibility
fi

echo "Starting PostgreSQL $VERSION"

export PGDATA=/var/lib/postgresql/$VERSION/data
/usr/lib/postgresql/$VERSION/bin/pg_ctl start -w -l /var/log/postgresql/$VERSION/server.log

echo "$VERSION" >/var/lib/postgresql/.version
