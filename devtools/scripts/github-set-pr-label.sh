#!/usr/bin/env bash
set -e

if [ -z "PR_NUMBER" ]; then
    echo "PR_NUMBER is not set"
    exit 1
fi

LABEL=$(./devtools/scripts/git-diff-label-calc.sh --debug)

gh pr edit "$PR_NUMBER" --add-label "test/$LABEL"
