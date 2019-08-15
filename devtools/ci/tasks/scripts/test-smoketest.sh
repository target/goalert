#!/bin/sh
set -x

trap "tar czf ../debug/debug-$(git rev-parse HEAD)-smoketest.tgz -C .. goalert/smoketest/smoketest_db_dump" EXIT

make -e smoketest
