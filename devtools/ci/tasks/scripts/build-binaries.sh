#!/bin/sh
BINARIES="bin/runjson bin/waitfor bin/simpleproxy bin/mockslack bin/goalert"
set -ex

make $BINARIES BUNDLE=1
VERSION=$(./bin/goalert version | head -n 1 |awk '{print $2}')

tar czf ../bin/goalert-${VERSION}-linux-amd64.tgz -C .. goalert/bin

rm -rf bin && make $BINARIES BUNDLE=1 GOOS=darwin
tar czf ../bin/goalert-${VERSION}-darwin-amd64.tgz -C .. goalert/bin
