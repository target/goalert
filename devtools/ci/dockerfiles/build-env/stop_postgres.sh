#!/bin/sh
set -e

su postgres -c "pg_ctl stop -m immediate"
