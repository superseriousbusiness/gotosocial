#!/bin/sh

# Test that the Swagger spec is up to date and valid.

set -eu

swagger_spec='docs/api/swagger.yaml'

# Temporary file for the regenerated Swagger spec.
regenerated_swagger_spec=$(mktemp -t 'swagger.yaml')
cleanup() {
  rm -f "${regenerated_swagger_spec}"
}
trap cleanup INT TERM EXIT

# Regenerate the Swagger spec and compare it to the working copy.
swagger generate spec --scan-models --exclude-deps --output "${regenerated_swagger_spec}"
if ! diff -u "${swagger_spec}" "${regenerated_swagger_spec}" > /dev/null; then
  echo "${swagger_spec} is out of date. Please run the following command to update it:" >&2
  echo "  swagger generate spec --scan-models --exclude-deps --output ${swagger_spec}" >&2
  exit 1
fi

# Validate the Swagger spec.
swagger validate "${swagger_spec}"
