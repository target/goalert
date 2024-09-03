#!/usr/bin/env bash
set -e

ensure_env() {
    if [ -z "$1" ]; then
        echo "$2 is not set"
        exit 1
    fi
}

if [ -z "$GITHUB_API_URL" ]; then
    GITHUB_API_URL=https://api.github.com
fi

ensure_env "$GITHUB_API_URL" "GITHUB_API_URL"
ensure_env "$PR_NUMBER" "PR_NUMBER"
ensure_env "$OWNER" "OWNER"
ensure_env "$REPO" "REPO"
LABEL=$(./devtools/scripts/git-diff-label-calc.sh --debug)

curl -sSL \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    -X PATCH \
    -H "Content-Type: application/json" \
    -d "{\"labels\":[\"test/$LABEL\"]}" \
    "$GITHUB_API_URL/repos/$OWNER/$REPO/issues/$PR_NUMBER"
