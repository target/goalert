#!/bin/sh
set -e

NOFMT=$(gofmt -s -l $(find . -name '*.go' |grep -v /vendor))

if test "$NOFMT" != ""
then
	echo "Found non-formatted files:"
	echo "$NOFMT"
	exit 1
fi

# assert Cypress versions are identical
PKG_JSON_VER=$(grep '"cypress":' web/src/package.json | awk -F '"' '{print $4}')
DOCKERFILE_VER=$(grep 'FROM cypress/included:' devtools/ci/dockerfiles/cypress-env/Dockerfile | awk -F ':' '{print $2}')
TASKFILE_VER=$(grep 'goalert/cypress-env' devtools/ci/tasks/test-integration.yml | awk '{print $6}')
if [ "$PKG_JSON_VER" != "$DOCKERFILE_VER" ]; then
  echo "Cypress versions do not match:"
  echo "package.json: ${PKG_JSON_VER} - Dockerfile: ${DOCKERFILE_VER}"
  exit 1
fi

# assert build-env versions are identical
BUILD_ENV_VER=go1.16.3-postgres13
for file in $(find devtools -name 'Dockerfile*')
do
  if ! grep -q "goalert/build-env" "$file"; then
    continue
  fi
  if ! grep -q "goalert/build-env:$BUILD_ENV_VER" "$file"; then
    echo "build-env version mismatch, expected $BUILD_ENV_VER"
    echo "  $file:"
    echo "  $(grep goalert/build-env "$file")"
    exit 1
  fi
done
for file in $(find devtools -name '*.yml')
do
  if ! grep -q "goalert/build-env" "$file"; then
    continue
  fi
  if ! grep -q "goalert/build-env, tag: $BUILD_ENV_VER" "$file"; then
    echo "build-env version mismatch, expected $BUILD_ENV_VER"
    echo "  $file:"
    echo "  $(grep goalert/build-env "$file")"
    exit 1
  fi
done

# taskfile contains quotes
if [ "'$PKG_JSON_VER'" != "$TASKFILE_VER" ]; then
  echo "Cypress versions do not match:"
  echo "package.json: ${PKG_JSON_VER} - test-integration.yml: ${TASKFILE_VER}"
  exit 1
fi

CHANGES=$(git status -s --porcelain)

if test "$CHANGES" != ""
then
	echo "Found changes in git:"
	echo "$CHANGES"
	exit 1
fi
