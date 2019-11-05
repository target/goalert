#!/bin/sh
set -e

su postgres -c "/usr/lib/postgresql/9.6/bin/pg_ctl start -l /var/log/postgresql/server.log"
