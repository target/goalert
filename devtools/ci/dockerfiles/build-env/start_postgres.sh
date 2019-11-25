#!/bin/sh
set -e

su postgres -c "pg_ctl start -w -l /var/log/postgresql/server.log"
