#!/usr/bin/env bash

set -e

START_DIR=$(pwd)
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

cd "$SCRIPT_DIR/app/handlers" || exit;

for d in *; do
    if [ -d "$d" ]; then
        cd "$d" || exit
        GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ../../../tmp/"$d" main.go
        cd ..
    fi
done

cd "$START_DIR" || exit
