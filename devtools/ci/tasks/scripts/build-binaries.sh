#!/bin/sh
set -ex

BINARIES="bin/goalert"

if [ "$BUILD_INTEGRATION" = "1" ]
then
    BINARIES="bin/goalert bin/waitfor bin/runjson bin/mockslack bin/simpleproxy bin/psql-lite"
fi

make check test $BINARIES BUNDLE=1
VERSION=$(./bin/goalert version | head -n 1 |awk '{print $2}')

tar czvf ../bin/goalert-${VERSION}-linux-amd64.tgz -C .. goalert/bin/goalert

make ../darwin/goalert/bin/goalert BUNDLE=1 GOOS=darwin BIN_DIR=../darwin/goalert/bin
tar czvf ../bin/goalert-${VERSION}-darwin-amd64.tgz -C ../darwin goalert/bin/goalert

if [ "$BUILD_INTEGRATION" = "1" ]
then
    echo Building integration test files...
    mkdir cypress
    (cd web/src && yarn webpack --config webpack.cypress.js && cp -r cypress/fixtures ../../cypress/ && cp cypress.json ../../)
    sed -i 's/\.ts/\.js/' cypress.json
    git rev-parse HEAD >COMMIT
    tar czvf ../bin/integration-${VERSION}-linux-amd64.tgz -C .. goalert/bin goalert/cypress goalert/cypress.json goalert/COMMIT goalert/devtools/ci goalert/.git/resource
fi
