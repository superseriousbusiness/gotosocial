#!/bin/sh

set -eu

# DEBUG returns whether DEBUG build is enabled.
DEBUG() { [ ! -z "${DEBUG-}" ]; }

CGO_ENABLED=0 go build -trimpath \
                       -tags "netgo osusergo static_build $(DEBUG && echo 'debugenv')" \
                       -ldflags="-s -w -extldflags '-static' -X 'main.Version=${VERSION:-$(git describe --tags --abbrev=0)}'" \
                       ./cmd/gotosocial
