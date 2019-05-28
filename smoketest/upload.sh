#!/bin/sh

FAILURES=""

while read -r line
do
    echo "$line"
    if echo "$line" | grep -q "^--- FAIL: "
    then
        NAME=$(echo "$line" | awk '{print $3}')
        FAILURES="$FAILURES $NAME"
    fi
done

if [ "$CI" == "drone" ] && [ -n "$FAILURES" ]
then
    SERVER=https://build-image-service.dev.target.goalert.me
    echo ""
    for name in $FAILURES
    do
        fname="smoketest/smoketest_db_dump/$NAME.sql"
        if [ -e "$fname" ]
        then
            FILE=$(curl -LksS "$SERVER/sql/" --data-binary "@$fname" -H "Authorization: Bearer $BUILD_IMAGE_SERVICE_TOKEN")
            if [ -n "$FILE" ]
            then
                echo "DB DUMP: $name"
                echo "    $SERVER/$FILE"
                echo ""
            else
                echo "Failed to upload: $name"
            fi
        fi
    done

    exit 1
fi
