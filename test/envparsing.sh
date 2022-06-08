#!/bin/sh

set -eu

EXPECTED='{"account-domain":"peepee","accounts-approval-required":false,"accounts-reason-required":false,"accounts-registration-open":true,"advanced-cookies-samesite":"strict","application-name":"gts","bind-address":"127.0.0.1","config-path":"./test/test.yaml","db-address":":memory:","db-database":"gotosocial_prod","db-password":"hunter2","db-port":6969,"db-tls-ca-cert":"","db-tls-mode":"disable","db-type":"sqlite","db-user":"sex-haver","email":"","host":"example.com","letsencrypt-cert-dir":"/gotosocial/storage/certs","letsencrypt-email-address":"","letsencrypt-enabled":true,"letsencrypt-port":80,"log-db-queries":true,"log-level":"info","media-description-max-chars":5000,"media-description-min-chars":69,"media-image-max-size":420,"media-remote-cache-days":30,"media-video-max-size":420,"oidc-client-id":"1234","oidc-client-secret":"shhhh its a secret","oidc-enabled":true,"oidc-idp-name":"sex-haver","oidc-issuer":"whoknows","oidc-scopes":["read","write"],"oidc-skip-verification":true,"password":"","path":"","port":6969,"protocol":"http","smtp-from":"queen@terfisland.org","smtp-host":"example.com","smtp-password":"hunter2","smtp-port":4269,"smtp-username":"sex-haver","software-version":"","statuses-cw-max-chars":420,"statuses-max-chars":69,"statuses-media-max-files":1,"statuses-poll-max-options":1,"statuses-poll-option-max-chars":50,"storage-backend":"local","storage-local-base-path":"/root/store","syslog-address":"127.0.0.1:6969","syslog-enabled":true,"syslog-protocol":"udp","trusted-proxies":["127.0.0.1/32","0.0.0.0/0"],"username":"","web-asset-base-dir":"/root","web-template-base-dir":"/root"}'

# Set all the environment variables to 
# ensure that these are parsed without panic
OUTPUT=$(GTS_LOG_LEVEL='info' \
GTS_LOG_DB_QUERIES=true \
GTS_APPLICATION_NAME=gts \
GTS_HOST=example.com \
GTS_ACCOUNT_DOMAIN='peepee' \
GTS_PROTOCOL=http \
GTS_BIND_ADDRESS='127.0.0.1' \
GTS_PORT=6969 \
GTS_TRUSTED_PROXIES='' \
GTS_DB_TYPE='sqlite' \
GTS_DB_ADDRESS=':memory:' \
GTS_DB_PORT=6969 \
GTS_DB_USER='sex-haver' \
GTS_DB_PASSWORD='hunter2' \
GTS_DB_DATABASE='gotosocial_prod' \
GTS_TLS_MODE='' \
GTS_DB_TLS_CA_CERT='' \
GTS_WEB_TEMPLATE_BASE_DIR='/root' \
GTS_WEB_ASSET_BASE_DIR='/root' \
GTS_ACCOUNTS_REGISTRATION_OPEN=true \
GTS_ACCOUNTS_APPROVAL_REQUIRED=false \
GTS_ACCOUNTS_REASON_REQUIRED=false \
GTS_MEDIA_IMAGE_MAX_SIZE=420 \
GTS_MEDIA_VIDEO_MAX_SIZE=420 \
GTS_MEDIA_DESCRIPTION_MIN_CHARS=69 \
GTS_MEDIA_DESCRIPTION_MAX_CHARS=5000 \
GTS_MEDIA_REMOTE_CACHE_DAYS=30 \
GTS_STORAGE_BACKEND='local' \
GTS_STORAGE_LOCAL_BASE_PATH='/root/store' \
GTS_STATUSES_MAX_CHARS=69 \
GTS_STATUSES_CW_MAX_CHARS=420 \
GTS_STATUSES_POLL_MAX_OPTIONS=1 \
GTS_STATUSES_POLL_OPTIONS_MAX_CHARS=69 \
GTS_STATUSES_MEDIA_MAX_FILES=1 \
GTS_LETS_ENCRYPT_ENABLED=false \
GTS_LETS_ENCRYPT_PORT=8080 \
GTS_LETS_ENCRYPT_CERT_DIR='/root/certs' \
GTS_LETS_ENCRYPT_EMAIL_ADDRESS='le@example.com' \
GTS_OIDC_ENABLED=true \
GTS_OIDC_IDP_NAME='sex-haver' \
GTS_OIDC_SKIP_VERIFICATION=true \
GTS_OIDC_ISSUER='whoknows' \
GTS_OIDC_CLIENT_ID='1234' \
GTS_OIDC_CLIENT_SECRET='shhhh its a secret' \
GTS_OIDC_SCOPES='read,write' \
GTS_SMTP_HOST='example.com' \
GTS_SMTP_PORT=4269 \
GTS_SMTP_USERNAME='sex-haver' \
GTS_SMTP_PASSWORD='hunter2' \
GTS_SMTP_FROM='queen@terfisland.org' \
GTS_SYSLOG_ENABLED=true \
GTS_SYSLOG_PROTOCOL='udp' \
GTS_SYSLOG_ADDRESS='127.0.0.1:6969' \
GTS_ADVANCED_COOKIES_SAMESITE='strict' \
go run ./cmd/gotosocial/... --config-path $(dirname ${0})/test.yaml debug config)

if [ "${OUTPUT}" != "${EXPECTED}" ]; then
    echo "OUTPUT not equal EXPECTED"
    exit 1
else
    echo "OK"
fi
