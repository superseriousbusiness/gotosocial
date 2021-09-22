#!/bin/bash

set -eu

SWAGGER_FILE="docs/api/swagger.yaml"

swagger generate spec -o "${SWAGGER_FILE}" --scan-models
sed -i "s/REPLACE_ME/${VERSION}/" "${SWAGGER_FILE}"
