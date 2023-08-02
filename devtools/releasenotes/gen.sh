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
exit 0

rm -f merge-details.txt
while read line; do
    # example: 980f7f848 Merge pull request #2944 from target/cleanup-shasum-noise/bar

    # get the PR number without the leading #
    PR=$(echo $line | awk '{print $5}' | cut -d'#' -f2)

    # branch without the first path segment
    BRANCH=$(echo $line | awk '{print $7}' | cut -d'/' -f2-)
    TITLE=$(curl -H "authorization: Bearer $GH_TOKEN" -s https://api.github.com/repos/target/goalert/pulls/$PR | jq -r .title)

    echo "BRANCH=$BRANCH TITLE=$TITLE" | tee -a merge-details.txt
done <merges.txt

script_dir=$(dirname "$0")

jq -n --arg last_version "$TAG" --arg model "$MODEL" --rawfile prompt "$script_dir/prompt.txt" --rawfile context merge-details.txt \
    '{
    "model": $model,
    "messages": [
        { "role": "system", "content": $prompt | sub("<LAST_VERSION>"; $last_version) },
        { "role": "user", "content": $context }
    ]
}' >request.json

gum spin --title "Generating release notes (this may take awhile)..." -- \
    curl https://api.openai.com/v1/chat/completions \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $OPENAI_API_KEY" \
    -o release-notes.json -d @request.json

cat release-notes.json | jq -r '.choices[0].message.content' >release-notes.md
