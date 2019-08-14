#!/bin/sh
set -ex

make bin/goalert BUNDLE=1
VERSION=$(./bin/goalert version | head -n 1 |awk '{print $2}')

tar czf ../bin/goalert-${VERSION}-linux-amd64.tgz -C .. goalert/bin/goalert

rm -rf bin && make bin/goalert BUNDLE=1 GOOS=darwin
tar czf ../bin/goalert-${VERSION}-darwin-amd64.tgz -C .. goalert/bin/goalert
