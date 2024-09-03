#!/usr/bin/env bash
set -e

if [ -z "$GITHUB_API_URL" ]; then
    GITHUB_API_URL=https://api.github.com
fi

if [ -z "$GITHUB_TOKEN" ]; then
    echo "GITHUB_TOKEN is not set"
    exit 1
fi

PR_NUMBER=$(jq --raw-output .pull_request.number "$GITHUB_EVENT_PATH")
LABEL=$(./devtools/scripts/git-diff-label-calc.sh --debug)

curl -sSL \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    -X PATCH \
    -H "Content-Type: application/json" \
    -d "{\"labels\":[\"test/$LABEL\"]}" \
    "$GITHUB_API_URL/repos/$GITHUB_REPOSITORY/issues/$PR_NUMBER"
