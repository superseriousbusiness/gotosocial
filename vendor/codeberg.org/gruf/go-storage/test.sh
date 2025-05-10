#!/bin/bash

export \
    MINIO_ADDR='127.0.0.1:8080' \
    MINIO_BUCKET='test' \
    MINIO_ROOT_USER='root' \
    MINIO_ROOT_PASSWORD='password' \
    MINIO_PID=0 \
    S3_DIR=$(mktemp -d)

# Drop the test S3 bucket and kill minio on exit
trap 'rm -rf "$S3_DIR"; [ $MINIO_PID -ne 0 ] && kill -9 $MINIO_PID' EXIT

# Create required S3 bucket dir
mkdir -p "${S3_DIR}/${MINIO_BUCKET}"

# Start the minio test S3 server instance
minio server --address "$MINIO_ADDR" "$S3_DIR" & > /dev/null 2>&1
MINIO_PID=$!; [ $? -ne 0 ] && {
    echo 'failed to start minio'
    exit 1
}

# Let server startup
sleep 1

# Run go-store tests
go test ./... -v