#!/bin/sh
set -ex
PREFIX=$1
make check test bin/goalert
VERSION=$(./bin/goalert version | head -n 1 |awk '{print $2}')
BVERSION=$(date +%s)-$(git rev-parse --short HEAD)

for PLATFORM in darwin-amd64 linux-amd64 linux-arm linux-arm64
do
    SRC=bin/goalert-${PLATFORM}.tgz
    make $SRC
    cp $SRC ${PREFIX}goalert-${VERSION}-${PLATFORM}.tgz
done

if [ "$BUILD_INTEGRATION" = "1" ]
then
    make bin/integration.tgz BUNDLE=1 BUILD_FLAGS=-trimpath
    cp bin/integration.tgz ${PREFIX}integration-${BVERSION}-linux-amd64.tgz
fi
