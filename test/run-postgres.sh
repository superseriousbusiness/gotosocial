#!/bin/bash

set -ex

# Determine available docker binary
_docker=$(command -v 'podman') || \
_docker=$(command -v 'docker') || \
{ echo 'docker not found'; exit 1; }

# Ensure test args are set.
ARGS=${@}; [ -z "$ARGS" ] && \
ARGS='./...'

# Database config.
DB_NAME='postgres'
DB_USER='postgres'
DB_PASS='postgres'
DB_IP='127.0.0.1'
DB_PORT=5432

# Start postgres container
CID=$($_docker run --detach \
    --publish "${DB_IP}:${DB_PORT}:${DB_PORT}" \
    --env "POSTGRES_DB=${DB_NAME}" \
    --env "POSTGRES_USER=${DB_USER}" \
    --env "POSTGRES_PASSWORD=${DB_PASS}" \
    --env "POSTGRES_HOST_AUTH_METHOD=trust" \
    --env "PGHOST=0.0.0.0" \
    --env "PGPORT=${DB_PORT}" \
    'docker.io/postgres:latest')

# On exit kill the container
trap "$_docker kill ${CID}" exit

sleep 5
#docker exec "$CID" psql --user "$DB_USER" --password "$DB_PASS" -c "CREATE DATABASE \"${DB_NAME}\" WITH LOCALE \"C.UTF-8\" TEMPLATE \"template0\";"
$_docker exec "$CID" psql --user "$DB_USER" --password "$DB_PASS" -c "GRANT ALL PRIVILEGES ON DATABASE \"${DB_NAME}\" TO \"${DB_USER}\";"

env \
GTS_DB_TYPE=postgres \
GTS_DB_ADDRESS=${DB_IP} \
GTS_DB_PORT=${DB_PORT} \
GTS_DB_USER=${DB_USER} \
GTS_DB_PASSWORD=${DB_PASS} \
GTS_DB_DATABASE=${DB_NAME} \
go test -p 1 ${ARGS}