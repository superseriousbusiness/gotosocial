#!/bin/sh

set -eu

# DEBUG returns whether DEBUG build is enabled.
DEBUG() { [ ! -z "${DEBUG-}" ]; }

# Available Go build tags, with explanation, followed by benefits of enabling it:
# - kvformat:    enables prettier output of log fields                       (slightly better performance)
# - notracing:   disables compiling-in otel tracing support                  (reduced binary size)
# - noerrcaller: disables caller function prefix in errors                   (slightly better performance)
# - debug:       enables /debug/pprof endpoint                               (adds debug, at performance cost)
# - debugenv:    enables /debug/pprof endpoint if DEBUG=1 env during runtime (adds debug, at performance cost)
CGO_ENABLED=0 go build -trimpath -v \
                       -tags "netgo osusergo static_build kvformat notracing $(DEBUG && echo 'debugenv')" \
                       -ldflags="-s -w -extldflags '-static' -X 'main.Version=${VERSION:-$(git describe --tags --abbrev=0)}'" \
                       ./cmd/gotosocial
