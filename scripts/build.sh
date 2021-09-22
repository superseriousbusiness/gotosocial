#!/bin/sh

set -eu

CGO_ENABLED=0 go build -trimpath \
                       -tags 'netgo osusergo static_build' \
                       -ldflags="-s -w -extldflags '-static' -X 'main.Commit=${COMMIT}' -X 'main.Version=${VERSION}'" \
                       ./cmd/gotosocial
