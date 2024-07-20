#!/bin/sh
set -e

for d in *; do
    [ ! -d "$d" ] && continue
    pushd "$d" >/dev/null 2>&1
    go mod tidy
    popd >/dev/null 2>&1
done
