#!/bin/sh

set -eu

CGO_ENABLED=0 go build -trimpath \
                       -tags 'netgo osusergo static_build' \
                       -ldflags="-s -w -extldflags '-static' -X 'main.Version=${VERSION:-$(git describe --tags --abbrev=0)}'" \
                       ./cmd/gotosocial
