#!/bin/bash

set -ex

HOST='localhost:8080'
ADDR='127.0.0.1'
PORT='8080'
DB=$(mktemp)
STORAGE=$(mktemp -d)
EXPORT="$(mktemp --suffix '.json')"

_kill() {
    pid=$(pidof gotosocial) \
    || return # i.e. not running
    kill -$1 $pid
}

gotosocial() {
    env \
    GTS_HOST=$HOST \
    GTS_PROTOCOL='http' \
    GTS_BIND_ADDRESS=$ADDR \
    GTS_PORT=$PORT \
    GTS_DB_TYPE='sqlite' \
    GTS_DB_ADDRESS=$DB \
    GTS_STORAGE_BACKEND='local' \
    GTS_STORAGE_LOCAL_BASE_PATH=$STORAGE \
    GTS_WEB_ASSET_BASE_DIR='./web/assets' \
    GTS_WEB_TEMPLATE_BASE_DIR='./web/template' \
    go run ./cmd/gotosocial ${@}
}

# Cleanup instance, database file, export file, storage on exit
trap "_kill 9; rm -f $DB; rm -f $EXPORT; rm -rf $STORAGE" exit

# Iterate list of usernames
for username in scoobert \
                shaggy \
                daphne \
                velma \
                fred \
                scrappy; do # hahaha

    # Create account for each username
    gotosocial admin account create \
        --username ${username} \
        --password 'sh1t_tiEr_Password Zomg123!' \
        --email ${username}@example.com

    # Confirm each of the accounts
    gotosocial admin account confirm \
        --username ${username}
done

# Ensure server runs once
gotosocial server start &

# Kill server
sleep 5
_kill 15
sleep 5

# Export current gts database to file
gotosocial admin export --path "$EXPORT" \
    || echo 'failed exporting to file'

# Clear database
rm -f $DB
touch $DB

# Import current gts database from file
gotosocial admin import --path "$EXPORT" \
    || echo 'failed importing from file'

# Ensure server can restart
{ gotosocial server start \
    || echo 'post-import failure'
} &

# Kill server
sleep 5
_kill 15
sleep 5

exit 0