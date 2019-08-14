#!/bin/sh
set -ex

BINARIES="bin/goalert"

if [ "$BUILD_TEST" = "1" ]
then
    BINARIES="bin/goalert bin/waitfor bin/runjson bin/mockslack bin/simpleproxy"
fi

make check $BINARIES BUNDLE=1
VERSION=$(./bin/goalert version | head -n 1 |awk '{print $2}')

tar czf ../bin/goalert-${VERSION}-linux-amd64.tgz -C .. goalert/bin/goalert

if [ "$BUILD_TEST" = "1" ]
then
    echo Building test files...
    mkdir cypress
    (cd web/src && yarn webpack --config webpack.cypress.js && cp -r cypress/fixtures ../../cypress/ && cp cypress.json ../../)
    sed -i 's/\.ts/\.js/' cypress.json
    tar czf ../bin/goalert-test-${VERSION}-linux-amd64.tgz -C .. goalert/bin goalert/cypress goalert/cypress.json
fi

rm -rf bin && make bin/goalert BUNDLE=1 GOOS=darwin
tar czf ../bin/goalert-${VERSION}-darwin-amd64.tgz -C .. goalert/bin/goalert
