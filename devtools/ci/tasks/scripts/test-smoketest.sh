#!/bin/sh
set -x

trap "tar czf ../debug/debug-$(date +%Y%m%d%H%M%S)-$(git rev-parse HEAD)-smoketest.tgz -C .. goalert/smoketest/smoketest_db_dump" EXIT

make -e smoketest
