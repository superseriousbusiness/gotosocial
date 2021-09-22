#!/bin/sh

set -eu

goreleaser release --snapshot --skip-publish
