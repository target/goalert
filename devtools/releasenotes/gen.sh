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

git log --oneline --merges --first-parent master --since="$TAG" | grep -v dependabot >merges.txt

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