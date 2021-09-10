#!/bin/bash

set -e

BRANCH_NAME="$(git rev-parse --abbrev-ref HEAD)"

docker push "superseriousbusiness/gotosocial:${BRANCH_NAME}"
