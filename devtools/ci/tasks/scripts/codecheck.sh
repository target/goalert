#!/bin/sh
set -e

# assert Cypress versions are identical
PKG_JSON_VER=$(grep '"cypress":' package.json | awk -F '"' '{print $4}')
DOCKERFILE_VER=$(grep 'FROM docker.io/cypress/included:' devtools/ci/dockerfiles/cypress-env/Dockerfile | awk -F ':' '{print $2}')
TASKFILE_VER=$(grep 'goalert/cypress-env' devtools/ci/tasks/test-integration.yml | awk '{print $6}')
if [ "$PKG_JSON_VER" != "$DOCKERFILE_VER" ]; then
  echo "Cypress versions do not match:"
  echo "package.json: ${PKG_JSON_VER} - Dockerfile: ${DOCKERFILE_VER}"
  exit 1
fi

# assert build-env versions are identical
BUILD_ENV_VER=go1.20.5-postgres13
for file in $(find devtools -name 'Dockerfile*'); do
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
for file in $(find devtools -name '*.yml'); do
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

# disk and DB MUST agree in schema file
DISK_HASH=$(grep "^-- DISK=" migrate/schema.sql | awk '{print $2}' | awk -F'=' '{print $2}')
PSQL_HASH=$(grep "^-- PSQL=" migrate/schema.sql | awk '{print $2}' | awk -F'=' '{print $2}')
if [ "$DISK_HASH" != "$PSQL_HASH" ]; then
  echo "Schema file describes mismatch in applied migration names:"
  echo "  DISK: $DISK_HASH"
  echo "  PSQL: $PSQL_HASH"
  exit 1
fi

SHA_CMD=$(if [ -x "$(command -v sha256sum)" ]; then echo "sha256sum"; else echo "shasum -a 256"; fi)
MIGRATION_HASH=$($SHA_CMD migrate/migrations/* | sort | $SHA_CMD | awk '{print $1}')
SCHEMA_HASH=$(grep "^-- DATA=" migrate/schema.sql | awk '{print $2}' | awk -F'=' '{print $2}')
if [ "$MIGRATION_HASH" != "$SCHEMA_HASH" ]; then
  echo "migrate/schema.sql is out-of-date (run make db-schema):"
  echo "  MIGRATIONS: $MIGRATION_HASH"
  echo "  SCHEMA: $SCHEMA_HASH"
  exit 1
fi

CHANGES=$(git status -s --porcelain)

if test "$CHANGES" != ""; then
  echo "Found changes in git:"
  echo "$CHANGES"
  exit 1
fi
