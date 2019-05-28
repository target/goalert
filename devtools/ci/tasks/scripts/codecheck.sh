#!/bin/sh
set -e

NOFMT=$(gofmt -l $(find . -name '*.go' |grep -v /vendor))

if test "$NOFMT" != ""
then
	echo "Found non-formatted files:"
	echo "$NOFMT"
	exit 1
fi

CHANGES=$(git status -s --porcelain)

if test "$CHANGES" != ""
then
	echo "Found changes in git:"
	echo "$CHANGES"
	exit 1
fi

