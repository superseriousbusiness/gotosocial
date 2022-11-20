#!/bin/sh

set -eu

# DEBUG returns whether DEBUG build is enabled.
DEBUG() { [ ! -z "${DEBUG-}" ]; }

export VERSION='v0.6.0-dev-kims-secret-sauce'
CGO_ENABLED=0 go build -trimpath \
                       -tags "netgo osusergo static_build kvformat $(DEBUG && echo 'debugenv')" \
                       -ldflags="-s -w -extldflags '-static' -X 'main.Version=${VERSION:-$(git describe --tags --abbrev=0)}'" \
                       ./cmd/gotosocial
