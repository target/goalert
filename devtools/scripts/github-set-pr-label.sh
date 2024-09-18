#!/usr/bin/env bash
set -e

if [ -z "PR_NUMBER" ]; then
    echo "PR_NUMBER is not set"
    exit 1
fi

LABEL=$(./devtools/scripts/git-diff-label-calc.sh --debug)

# Remove any existing test/* labels
for label in $(gh pr view "$PR_NUMBER" --json labels --jq '.labels[] | select(.name | startswith("size/")) | .name'); do
    if [ "$label" == "$LABEL" ]; then
        continue # Skip the label we want to add
    fi
    gh pr edit "$PR_NUMBER" --remove-label "$label"
done

gh pr edit "$PR_NUMBER" --add-label "$LABEL"
