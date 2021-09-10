#!/bin/sh

set -e

# "./..." = all tests
# "-p 1" = run with parallel value of 1 -- in other words, one test at a time

# run tests with sqlite in-memory database
GTS_DB_TYPE="sqlite" GTS_DB_ADDRESS=":memory:" go test -p 1 ./...

# run tests with postgres database at either GTS_DB_ADDRESS or default localhost
GTS_DB_TYPE="postgres" GTS_DB_ADDRESS="${GTS_DB_ADDRESS:-localhost}" go test -p 1 ./...
