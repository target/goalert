#!/bin/sh
export GOALERT_DB_URL="$DB_URL"
set -x

trap "tar czf ../debug/debug-$(git rev-parse HEAD)-smoketest.tgz -C .. goalert/smoketest/smoketest_db_dump" EXIT

make smoketest
