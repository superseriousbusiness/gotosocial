#!/bin/sh

set -e

# Log and execute provided args.
log_exec() { echo "$ ${*}"; "$@"; }

# Grab environment variables and set defaults + requirements.
GO_BUILDTAGS="${GO_BUILDTAGS-} netgo osusergo static_build kvformat timetzdata"
GO_LDFLAGS="${GO_LDFLAGS-} -s -w -extldflags '-static' -X 'main.Version=${VERSION:-$(git describe --tags --abbrev=0)}'"
GO_GCFLAGS=${GO_GCFLAGS-}

# Maintain old $DEBUG compat.
[ ! -z "$DEBUG" ] && \
    GO_BUILDTAGS="${GO_BUILDTAGS} debugenv"

# Available Go build tags, with explanation, followed by benefits of enabling it:
# - kvformat:       enables prettier output of log fields                          (slightly better performance)
# - timetzdata:     embed timezone database inside binary                          (allow setting local time inside Docker containers, at cost of 450KB)
# - notracing:      disables compiling-in otel tracing support                     (reduced binary size, better performance)
# - nometrics:      disables compiling-in otel metrics support                     (reduced binary size, better performance)
# - noerrcaller:    disables caller function prefix in errors                      (slightly better performance, at cost of err readability)
# - debug:          enables /debug/pprof endpoint                                  (adds debug, at performance cost)
# - debugenv:       enables /debug/pprof endpoint if DEBUG=1 env during runtime    (adds debug, at performance cost)
# - moderncsqlite3: reverts to using the C-to-Go transpiled SQLite driver          (disables the WASM-based SQLite driver)
# - nowasm:         [UNSUPPORTED] removes all WebAssembly from builds including 
#                   ffmpeg, ffprobe and SQLite (instead falling back to modernc).
log_exec env CGO_ENABLED=0 go build -trimpath -v \
                       -tags "${GO_BUILDTAGS}" \
                       -ldflags="${GO_LDFLAGS}" \
                       -gcflags="${GO_GCFLAGS}" \
                       ./cmd/gotosocial
