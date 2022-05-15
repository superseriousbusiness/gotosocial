#!/bin/sh

(
    # Run in subshell with cmd echo
    set -ex

    # Run debug tests
    DEBUG=  go test -tags= -v
    DEBUG=  go test -tags=debug -v
    DEBUG=  go test -tags=debugenv -v
    DEBUG=y go test -tags=debugenv -v
    DEBUG=1 go test -tags=debugenv -v
    DEBUG=y go test -tags=debugenv,debug -v
    DEBUG=y go test -tags= -v
)

echo 'success!'