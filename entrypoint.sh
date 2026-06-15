#!/bin/sh
set -e

if [ "$GO_ENVIRONMENT" = "development" ] && [ "$1" = "-action" ] && [ "$2" = "serve" ]; then
    exec air
else
    sqlc generate
    go build -buildvcs=false -o main .
    exec ./main "$@"
fi
