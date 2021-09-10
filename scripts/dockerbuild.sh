#!/bin/bash

set -eu

BRANCH_NAME="$(git rev-parse --abbrev-ref HEAD)"

docker build -t "superseriousbusiness/gotosocial:${BRANCH_NAME}" .
