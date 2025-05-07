#!/bin/sh

set -eu

# Cheeky little convenience script for fetching
# the latest snapshot build of GoToSocial from
# the Minio S3 bucket.
#
# Requires curl and jq.
#
# Change the variables below for your particular
# platform and architecture (default linux amd64). 
GTS_PLATFORM="linux"
GTS_ARCH="amd64"
GTS_FILENAME="gotosocial_${GTS_PLATFORM}_${GTS_ARCH}.tar.gz"

API_HOST="codeberg.org"
ORG="superseriousbusiness"
REPO="gotosocial"
BRANCH="main"

ENDPOINT="https://${API_HOST}/api/v1/repos/${ORG}/${REPO}/git/refs/heads/${BRANCH}"

if [ $# -ge 1 ] && [ -n "$1" ]; then
	LATEST_HASH=$1
else
	echo "fetching latest hash from endpoint '${ENDPOINT}'"
	LATEST_HASH="$(curl --silent --fail --retry 5 --retry-max-time 180 --max-time 30 "${ENDPOINT}" | jq -r 'first(.[]|.object.sha)')"
	echo "got latest hash = ${LATEST_HASH}"
fi

MINIO_HOST="s3.superseriousbusiness.org"
MINIO_BUCKET="gotosocial-snapshots"
MINIO_ENDPOINT="https://${MINIO_HOST}/${MINIO_BUCKET}/${LATEST_HASH}/${GTS_FILENAME}"

echo "fetching latest snapshot tar.gz from endpoint '${MINIO_ENDPOINT}'"
curl --silent --fail --retry 5 --retry-max-time 600 --max-time 1800 "${MINIO_ENDPOINT}" --output "./${GTS_FILENAME}"
echo "got latest snapshot!"
