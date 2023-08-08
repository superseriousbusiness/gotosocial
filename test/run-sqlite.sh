#!/bin/sh

set -e

# Ensure test args are set.
ARGS=${@}; [ -z "$ARGS" ] && \
ARGS='./...'

# Run the SQLite tests.
GTS_DB_TYPE=sqlite \
GTS_DB_ADDRESS=':memory:' \
go test ${ARGS}