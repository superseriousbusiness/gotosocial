#!/bin/sh

set -eu

# DEBUG returns whether DEBUG build is enabled.
DEBUG() { [ ! -z "${DEBUG-}" ]; }

CGO_ENABLED=0 go build -trimpath -v \
                       -tags "netgo osusergo static_build kvformat notracing $(DEBUG && echo 'debugenv')" \
                       -ldflags="-s -w -extldflags '-static' -X 'main.Version=${VERSION:-$(git describe --tags --abbrev=0)}'" \
                       -gcflags=all='-B -C -l=4' \
                       ./cmd/gotosocial
