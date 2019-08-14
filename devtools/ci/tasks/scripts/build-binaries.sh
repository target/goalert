#!/bin/sh
set -ex

make bin/goalert bin/waitfor bin/runjson bin/mockslack bin/simpleproxy BUNDLE=1
VERSION=$(./bin/goalert version | head -n 1 |awk '{print $2}')

tar czf ../bin/goalert-${VERSION}-linux-amd64.tgz -C .. goalert/bin/goalert
mkdir -p ../bin/goalert/bin
cp bin/goalert ../bin/goalert/bin/

tar czf ../bin/goalert-test-binaries-${VERSION}-linux-amd64.tgz -C .. goalert/bin

rm -rf bin && make bin/goalert BUNDLE=1 GOOS=darwin
tar czf ../bin/goalert-${VERSION}-darwin-amd64.tgz -C .. goalert/bin/goalert
