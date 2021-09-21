#!/bin/sh

set -eu

COMMIT=$(git rev-list -1 HEAD)
VERSION=$(cat ./version)

COMMIT="${COMMIT}" VERSION="${VERSION}" goreleaser release --snapshot --skip-publish
