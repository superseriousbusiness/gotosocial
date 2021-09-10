#!/bin/sh

set -e

# "-count 1" = run all tests once, ignoring cache; useful for when we're running tests with different database back to back like this
# "-p 1" = run with parallel value of 1 -- in other words, one test at a time
# "./..." = all tests

# run tests with sqlite in-memory database
GTS_DB_TYPE="sqlite" GTS_DB_ADDRESS=":memory:" go test -count 1 -p 1 ./...

# run tests with postgres database at either GTS_DB_ADDRESS or default localhost
GTS_DB_TYPE="postgres" GTS_DB_ADDRESS="${GTS_DB_ADDRESS:-localhost}" go test -count 1 -p 1 ./...
