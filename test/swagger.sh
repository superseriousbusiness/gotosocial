#!/bin/sh

# Test that the Swagger spec is up to date and valid.

set -eu

swagger_cmd() {
  go run ./vendor/github.com/go-swagger/go-swagger/cmd/swagger "$@"
}
swagger_spec='docs/api/swagger.yaml'

# Temporary directory for the regenerated Swagger spec.
temp_dir=$(mktemp -d)
# Can't use mktemp directly because we need to control the file extension.
regenerated_swagger_spec="${temp_dir}/swagger.yaml"
cleanup() {
  rm -rf "${temp_dir}"
}
trap cleanup INT TERM EXIT

# Regenerate the Swagger spec and compare it to the working copy.
swagger_cmd generate spec --scan-models --exclude-deps --output "${regenerated_swagger_spec}"
if ! diff -u "${swagger_spec}" "${regenerated_swagger_spec}" > /dev/null; then
  echo "${swagger_spec} is out of date. Please run the following command to update it:" >&2
  echo "  go run github.com/go-swagger/go-swagger/cmd/swagger generate spec --scan-models --exclude-deps --output ${swagger_spec}" >&2
  exit 1
fi

# Validate the Swagger spec.
swagger_cmd validate "${swagger_spec}"
