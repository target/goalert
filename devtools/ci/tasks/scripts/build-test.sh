#!/bin/sh
set -ex

mkdir -p /run/postgresql
chown postgres:postgres /run/postgresql
export PGDATA=/var/lib/postgresql/data
su postgres -c /usr/lib/postgresql/11/bin/initdb
su postgres -c '/usr/lib/postgresql/11/bin/pg_ctl start'
trap 'su postgres -c "/usr/lib/postgresql/11/bin/pg_ctl -m immediate stop"' EXIT
export DB_URL=postgres://postgres@localhost:5432?sslmode=disable

touch web/src/yarn.lock # ensure we refresh node_modules

make check test install smoketest cypress DB_URL=$DB_URL BUNDLE=1

CYPRESS_viewportWidth=1440 CYPRESS_viewportHeight=900 bin/runjson -logs=./wide-logs <devtools/runjson/ci-cypress.json

CYPRESS_viewportWidth=375 CYPRESS_viewportHeight=667 bin/runjson -logs=./mobile-logs <devtools/runjson/ci-cypress.json

cat bin/goalert | gzip -9 >bin/goalert-$(date +%Y%m%d%H%M%S).gz
