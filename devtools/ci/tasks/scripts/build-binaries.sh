#!/bin/sh
set -ex

BINARIES="bin/goalert"

if [ -n "$BUILD_ALL" ]
then
    BINARIES="bin/goalert bin/waitfor bin/runjson bin/mockslack bin/simpleproxy"
fi

make $BINARIES BUNDLE=1
VERSION=$(./bin/goalert version | head -n 1 |awk '{print $2}')

tar czf ../bin/goalert-${VERSION}-linux-amd64.tgz -C .. goalert/bin/goalert
mkdir -p ../bin/goalert/bin
cp bin/goalert ../bin/goalert/bin/

if [ -n "$BUILD_ALL" ]
then
    tar czf ../bin/goalert-all-${VERSION}-linux-amd64.tgz -C .. goalert/bin
fi

rm -rf bin && make bin/goalert BUNDLE=1 GOOS=darwin
tar czf ../bin/goalert-${VERSION}-darwin-amd64.tgz -C .. goalert/bin/goalert
