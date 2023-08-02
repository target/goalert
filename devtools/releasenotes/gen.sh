#!/bin/bash
set -e

if [ -z "$GH_TOKEN" ]; then
    GH_TOKEN=$(gum input --password --header "Enter your GitHub token.")
fi

if [ -z "$OPENAI_API_KEY" ]; then
    OPENAI_API_KEY=$(gum input --password --header "Enter your OpenAI key.")
fi

MODEL=$(gum choose --header "Pick a model." "gpt-4" "gpt-3.5-turbo" "gpt-3.5")

TAGS=$(git tag -l 'v*.*.*' | sort -rV)
TAG=$(gum choose --header "Pick the most recent release tag." $TAGS)

DIR=$(pwd)/bin/release-notes
mkdir -p "$DIR"

git log --oneline --merges --first-parent master --since="$TAG" | grep -v dependabot >"$DIR/merges.txt"

cat "$DIR/merges.txt" | awk '{print $5}' | sed 's/^#//' | tac >"$DIR/prs.txt"

mkdir -p "$DIR/prs"
# Summaries of PRs
while read PR; do
    if [ -f "$DIR/prs/$PR.json" ]; then
        echo "Skipping PR $PR"
        continue
    fi
    echo "Generating summary for PR $PR"

    # try generating the summary up to 3 times, otherwise exit
    FAILED=1
    for i in 1 2 3; do
        if go run ./devtools/releasenotes -pr "$PR" >"$DIR/prs/_$PR.json"; then
            FAILED=0
            break
        fi
        echo "Failed to generate summary for PR $PR, retrying..."
    done
    if [ "$FAILED" -ne 0 ]; then
        echo "Failed to generate summary for PR $PR, exiting..."
        exit 1
    fi
    mv "$DIR/prs/_$PR.json" "$DIR/prs/$PR.json"
done <"$DIR/prs.txt"
