#!/bin/sh

set -eu

COMMIT="${COMMIT:-12345678}"
VERSION="${VERSION:-0.0.0}"

CGO_ENABLED=0 go build -trimpath \
                       -tags 'netgo osusergo static_build' \
                       -ldflags="-s -w -extldflags '-static' -X 'main.Commit=${COMMIT}' -X 'main.Version=${VERSION}'" \
                       ./cmd/gotosocial
