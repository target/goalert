#!/bin/sh
set -e

su postgres -c "pg_ctl start -l /var/log/postgresql/server.log"
