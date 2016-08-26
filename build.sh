#!/bin/bash

export GOOS=linux
export GOARCH=amd64

NAME="shitfall"

echo -e "Building $NAME with:"

echo ""
echo "GOOS=$GOOS"
echo "GOARCH=$GOARCH"
echo ""

BINARY_FILENAME="$NAME-$GOOS-$GOARCH"

go build -v -o $BINARY_FILENAME *.go

echo -e "\nDone: \033[33m$BINARY_FILENAME\033[0m 💪"
