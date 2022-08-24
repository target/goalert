#!/bin/sh
set -x
start_postgres

trap "stop_postgres; tar czf ../debug/debug-$(date +%Y%m%d%H%M%S)-$(git rev-parse HEAD)-smoketest.tgz -C .. goalert/test/smoke/smoketest_db_dump" EXIT

make -e test-smoke DB_URL=$DB_URL
