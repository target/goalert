#!/bin/sh
set -e

su postgres -c "/usr/lib/postgresql/11/bin/pg_ctl stop -m immediate"
