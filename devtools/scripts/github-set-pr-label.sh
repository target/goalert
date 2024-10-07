#!/usr/bin/env bash
set -e

if [ -z "PR_NUMBER" ]; then
    echo "PR_NUMBER is not set"
    exit 1
fi

# test change

MY_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LABEL=$("$MY_DIR/git-diff-label-calc.sh" --debug)

# Remove any existing test/* labels
for label in $(gh pr view "$PR_NUMBER" --json labels --jq '.labels[] | select(.name | startswith("size/")) | .name'); do
    if [ "$label" == "$LABEL" ]; then
        continue # Skip the label we want to add
    fi
    gh pr edit "$PR_NUMBER" --remove-label "$label"
done

gh pr edit "$PR_NUMBER" --add-label "$LABEL"
