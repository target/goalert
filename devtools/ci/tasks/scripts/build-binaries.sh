#!/bin/sh
set -ex

make check test bin/goalert bin/goalert-linux-amd64.tgz bin/goalert-darwin-amd64.tgz BUNDLE=1 BUILD_FLAGS=-trimpath
VERSION=$(./bin/goalert version | head -n 1 |awk '{print $2}')
BVERSION=$(date +%s)-$(git rev-parse --short HEAD)

for PLATFORM in darwin linux
do
    cp bin/goalert-${PLATFORM}-amd64.tgz ../bin/goalert-${VERSION}-${PLATFORM}-amd64.tgz
done

if [ "$BUILD_INTEGRATION" = "1" ]
then
    make bin/integration.tgz BUNDLE=1 BUILD_FLAGS=-trimpath
    cp bin/integration.tgz ../bin/integration-${BVERSION}-linux-amd64.tgz
fi
