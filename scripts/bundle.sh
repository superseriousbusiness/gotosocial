#!/bin/sh

# this script is really just here because GoReleaser doesn't let
# you set env vars in your 'before' commands in the free version

set -eu

# Transpile typescript assets.
yarn --cwd web/source tsc

# Bundle transpiled javascript.
BUDO_BUILD=1 node web/source
