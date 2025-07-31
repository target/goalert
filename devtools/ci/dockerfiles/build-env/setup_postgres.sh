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
    echo "synchronous_commit = off" >>$PGDATA/postgresql.conf
    echo "max_connections = 500" >>$PGDATA/postgresql.conf
    
    # Additional performance optimizations for ephemeral test environments
    echo "wal_level = minimal" >>$PGDATA/postgresql.conf
    echo "max_wal_senders = 0" >>$PGDATA/postgresql.conf
    echo "checkpoint_completion_target = 0.9" >>$PGDATA/postgresql.conf
    echo "wal_buffers = 16MB" >>$PGDATA/postgresql.conf
    echo "shared_buffers = 128MB" >>$PGDATA/postgresql.conf
    echo "work_mem = 2MB" >>$PGDATA/postgresql.conf
    echo "maintenance_work_mem = 32MB" >>$PGDATA/postgresql.conf
    echo "effective_cache_size = 512MB" >>$PGDATA/postgresql.conf
    echo "random_page_cost = 1.1" >>$PGDATA/postgresql.conf
    echo "logging_collector = off" >>$PGDATA/postgresql.conf
    echo "log_statement = 'none'" >>$PGDATA/postgresql.conf
    echo "log_duration = off" >>$PGDATA/postgresql.conf
    echo "log_lock_waits = off" >>$PGDATA/postgresql.conf
    echo "log_checkpoints = off" >>$PGDATA/postgresql.conf
    echo "autovacuum = off" >>$PGDATA/postgresql.conf
    echo "max_locks_per_transaction = 256" >>$PGDATA/postgresql.conf
    echo "max_pred_locks_per_transaction = 128" >>$PGDATA/postgresql.conf
    echo "temp_buffers = 8MB" >>$PGDATA/postgresql.conf
    echo "max_prepared_transactions = 0" >>$PGDATA/postgresql.conf
done
chown -R postgres:postgres /var/lib/postgresql
