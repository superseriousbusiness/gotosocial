#!/bin/sh

set -eu

export COMMIT=$(git rev-list -1 HEAD)
export VERSION=$(cat ./version)

go build -ldflags="-X 'main.Commit=$COMMIT' -X 'main.Version=$VERSION'" ./cmd/gotosocial
