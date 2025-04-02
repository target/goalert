#!/usr/bin/env bash
set -e

# Validates that:
# - existing files in the migrate/migrations directory not modified
# - new files in the migrate/migrations directory are alphabetically last

# env setting defaults
if [ -z "$MAIN_BRANCH" ]; then
    MAIN_BRANCH=master
fi

# const settings
MIGRATIONS_DIR=migrate/migrations

if [ -z "$DEBUG" ]; then
    DEBUG=0
fi

if [ "$1" = "--debug" ]; then
    DEBUG=1
fi

# Any change other than A (added) should result in PR rejection.
MODIFIED_FILES=$(git diff --ignore-blank-lines -w --diff-filter=a --merge-base "origin/$MAIN_BRANCH" --name-only -- "$MIGRATIONS_DIR")
if [ -n "$MODIFIED_FILES" ]; then
    echo "The following files in the $MIGRATIONS_DIR directory have been modified:"
    echo "$MODIFIED_FILES"
    echo ""
    echo "Please make alterations/fixes in a new migration."
    echo "Since existing migrations have already been applied in the wild, they cannot be modified in a PR."
    exit 1
fi

# Any new file(s) must be the last file(s) in the directory.
LAST_MIGRATION_ON_MAIN=$(git ls-tree -r --name-only "origin/$MAIN_BRANCH" -- "$MIGRATIONS_DIR" | sort | tail -n 1)

BAD_MIGRATIONS=()
# Check if any new file is less than the last migration on the main branch.
while IFS= read -r file; do
    if [ "$file" \< "$LAST_MIGRATION_ON_MAIN" ]; then
        BAD_MIGRATIONS+=("$file")
    fi
done < <(git diff --ignore-blank-lines -w --diff-filter=A --merge-base "origin/$MAIN_BRANCH" --name-only -- "$MIGRATIONS_DIR")

if [ ${#BAD_MIGRATIONS[@]} -gt 0 ]; then
    echo "New migrations must be named alphabetically AFTER the last migration to be committed to the $MAIN_BRANCH branch."
    echo "This is currently: $LAST_MIGRATION_ON_MAIN"
    echo ""
    echo "The following file(s) are out-of order:"
    echo "${BAD_MIGRATIONS[@]}"
    echo ""
    echo ""
    echo "To fix this automatically, run 'go tool ordermigrations' locally."
    echo "This tool will automatically update the timestamps of all new migrations to comply with this requirement."
    exit 1
fi
