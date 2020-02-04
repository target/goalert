#!/bin/sh
set -e

NOFMT=$(gofmt -s -l $(find . -name '*.go' |grep -v /vendor))

if test "$NOFMT" != ""
then
	echo "Found non-formatted files:"
	echo "$NOFMT"
	exit 1
fi

CHANGES=$(git status -s --porcelain)

if test "$CHANGES" != ""
then
	echo "Found changes in git:"
	echo "$CHANGES"
	exit 1
fi

# assert Cypress versions are identical
PKG_JSON_VER=$(grep '"cypress":' web/src/package.json | awk -F '"' '{print $4}')
DOCKERFILE_VER=$(grep 'FROM cypress/included:' devtools/ci/dockerfiles/cypress-env/Dockerfile | awk -F ':' '{print $2}')
if [ "$PKG_JSON_VER" != "$DOCKERFILE_VER" ]; then
  echo "Cypress versions do not match:"
  echo "package.json: ${PKG_JSON_VER} - Dockerfile: ${DOCKERFILE_VER}"
  exit 1
fi
