#!/bin/sh

GO_ARCH=$(go env GOARCH)
SYS_ARCH=$(uname -m)

# map x86_64 to amd64
if [ "$SYS_ARCH" = "x86_64" ]; then
    SYS_ARCH="amd64"
fi

if ! [ "$GO_ARCH" = "$SYS_ARCH" ]; then

    echo "\033[1;33mWARNING: GOARCH ($GO_ARCH) does not match your system architecture ($SYS_ARCH)\033[0m"
    echo "\033[1;33mThis may cause issues when building or starting the project.\033[0m"
fi
