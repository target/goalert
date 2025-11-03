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
    VERSION=17 # Default to PostgreSQL 17 for compatibility
fi

echo "Starting PostgreSQL $VERSION"

export PGDATA=/var/lib/postgresql/$VERSION/data
/usr/lib/postgresql/$VERSION/bin/pg_ctl start -w -l /var/log/postgresql/$VERSION/server.log || {
    cat /var/log/postgresql/$VERSION/server.log
    echo "Failed to start PostgreSQL $VERSION"
    exit 1
}

echo "$VERSION" >/var/lib/postgresql/.version

# symlink current log file to a consistent name (replace if necessary)
chmod a+r /var/log/postgresql/$VERSION/server.log
ln -sf /var/log/postgresql/$VERSION/server.log /var/log/postgresql/server.log
