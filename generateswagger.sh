#!/bin/bash

SWAGGER_FILE="swagger.yaml"
GTS_VERSION="$(cat version)"

swagger generate spec -o "${SWAGGER_FILE}" --scan-models
sed -i "s/REPLACE_ME/${GTS_VERSION}/" "${SWAGGER_FILE}"
