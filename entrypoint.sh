#!/bin/sh
set -e

# Development serve: air rebuilds and restarts on source changes.
if [ "$GO_ENVIRONMENT" = "development" ] && [ "$1" = "-action" ] && [ "$2" = "serve" ]; then
    exec air
fi

# Everything else (production serve, one-shot cli actions): build once and run.
sqlc generate
go build -buildvcs=false -o /tmp/main .
exec /tmp/main "$@"
