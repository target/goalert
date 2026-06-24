#!/bin/sh
set -e

VERSION=$1

if [ -z "$VERSION" ]; then
    # Try to read the version from the file
    if [ -f /var/lib/postgresql/.version ]; then
        VERSION=$(cat /var/lib/postgresql/.version)
    else
        echo "No version specified defaulting to 17"
        VERSION=17 # Default to PostgreSQL 17 for compatibility
    fi
fi

echo "Stopping PostgreSQL $VERSION"

export PGDATA=/var/lib/postgresql/$VERSION/data
if [ ! -f $PGDATA/postmaster.pid ]; then
    exit 0
fi
/usr/lib/postgresql/$VERSION/bin/pg_ctl stop -m immediate || rm -f $PGDATA/postmaster.pid
