#!/bin/bash
set -e

if [ "$CI" != "1" ]; then
    echo "This script is only intended to be run in CI."
    exit 0
fi

GLOBAL_CACHE_PATH=/var/cache/arc

ACTION="$1"
if [ "$ACTION" != "restore" ] && [ "$ACTION" != "save" ]; then
    echo "Usage: $0 <restore|save>"
    exit 1
fi

KEY="goalert-cache-$(uname -s)-$(uname -m)-$(cat devtools/scripts/ci-cache.sh go.mod go.sum package.json bun.lock *.version | sha256sum | awk '{print $1}')"

if [ -z "$GLOBAL_CACHE_PATH" ]; then
    echo "GLOBAL_CACHE_PATH is not set. Skipping cache management."
    exit 0
fi

echo "Cache key: $KEY"

path_key() {
    echo "$1" | sha256sum | awk '{print $1}'
}
resolve_path() {
    echo "$GLOBAL_CACHE_PATH/$KEY/$(path_key "$1")"
}

resolve_tmp_path() {
    local hash=$(echo "$1" | sha256sum | awk '{print $1}')
    echo "$1/$hash"
}

PATHS=(
    "$(go env GOMODCACHE)"
    "$(go env GOCACHE)"
    "$HOME/.cache/goalert-gettool"
    "$HOME/.cache/Cypress"
    "$HOME/.cache/ms-playwright"
    "$HOME/.bun/install/cache"
)

if [ "$ACTION" == "restore" ]; then
    echo "Restoring cache..."
    for path in "${PATHS[@]}"; do
        SRC=$(resolve_path "$path")
        if [ -d "$SRC" ]; then
            echo "Restoring $path"
            mkdir -p "$path"
            cp -r "$SRC/"* "$path/"
        else
            echo "No cache found for $path"
        fi
    done
    if [ -d "$GLOBAL_CACHE_PATH/$KEY" ]; then
        touch "$GLOBAL_CACHE_PATH/$KEY/.last_used"
    else
        echo "No cache found for $KEY"
    fi
else
    # Skip saving if the cache already exists
    if [ -d "$GLOBAL_CACHE_PATH/$KEY" ]; then
        echo "Cache already exists. Skipping save."
        exit 0
    fi
    echo "Saving cache..."
    TMP_DIR=$(mktemp -d -p "$GLOBAL_CACHE_PATH")
    if [ ! -d "$TMP_DIR" ]; then
        echo "Failed to create temporary directory for cache."
        exit 1
    fi
    for path in "${PATHS[@]}"; do
        if [ -d "$path" ]; then
            echo "Saving $path"
            DST="$TMP_DIR/$(path_key "$path")"
            mkdir "$DST"
            cp -r "$path/"* "$DST/"
        else
            echo "No cache to save for $path"
        fi
    done
    touch "$TMP_DIR/.last_used"
    mv "$TMP_DIR" "$GLOBAL_CACHE_PATH/$KEY"
fi

MONTH_AGO=$(date -d '1 month ago' +%s)
if [ -z "$(ls -A "$GLOBAL_CACHE_PATH")" ]; then
    echo "GLOBAL_CACHE_PATH is empty. Skipping old cache cleanup."
else
    for last_used in "$GLOBAL_CACHE_PATH"/*/.last_used; do
        TIME=$(stat -c %Y "$last_used")
        if [ "$TIME" -lt "$MONTH_AGO" ]; then
            echo "Deleting old cache $(dirname "$last_used")"
            rm -rf "$(dirname "$last_used")"
        fi
    done
fi

echo "Cache $ACTION complete."
