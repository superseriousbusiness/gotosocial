#!/bin/sh

set -eu

goreleaser release --rm-dist --snapshot --skip-publish
