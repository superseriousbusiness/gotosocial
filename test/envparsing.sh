#!/bin/sh

set -eu

EXPECT='{"account-domain":"peepee","accounts-allow-custom-css":true,"accounts-approval-required":false,"accounts-reason-required":false,"accounts-registration-open":true,"advanced-cookies-samesite":"strict","advanced-rate-limit-requests":6969,"application-name":"gts","bind-address":"127.0.0.1","cache":{"gts":{"account-max-size":99,"account-sweep-freq":1000000000,"account-ttl":10800000000000,"block-max-size":100,"block-sweep-freq":10000000000,"block-ttl":300000000000,"domain-block-max-size":1000,"domain-block-sweep-freq":60000000000,"domain-block-ttl":86400000000000,"emoji-category-max-size":100,"emoji-category-sweep-freq":10000000000,"emoji-category-ttl":300000000000,"emoji-max-size":500,"emoji-sweep-freq":10000000000,"emoji-ttl":300000000000,"mention-max-size":500,"mention-sweep-freq":10000000000,"mention-ttl":300000000000,"notification-max-size":500,"notification-sweep-freq":10000000000,"notification-ttl":300000000000,"status-max-size":500,"status-sweep-freq":10000000000,"status-ttl":300000000000,"tombstone-max-size":100,"tombstone-sweep-freq":10000000000,"tombstone-ttl":300000000000,"user-max-size":100,"user-sweep-freq":10000000000,"user-ttl":300000000000}},"config-path":"internal/config/testdata/test.yaml","db-address":":memory:","db-database":"gotosocial_prod","db-password":"hunter2","db-port":6969,"db-tls-ca-cert":"","db-tls-mode":"disable","db-type":"sqlite","db-user":"sex-haver","dry-run":false,"email":"","host":"example.com","instance-deliver-to-shared-inboxes":false,"instance-expose-peers":true,"instance-expose-public-timeline":true,"instance-expose-suspended":true,"landing-page-user":"admin","letsencrypt-cert-dir":"/gotosocial/storage/certs","letsencrypt-email-address":"","letsencrypt-enabled":true,"letsencrypt-port":80,"log-db-queries":true,"log-level":"info","media-description-max-chars":5000,"media-description-min-chars":69,"media-emoji-local-max-size":420,"media-emoji-remote-max-size":420,"media-image-max-size":420,"media-remote-cache-days":30,"media-video-max-size":420,"oidc-client-id":"1234","oidc-client-secret":"shhhh its a secret","oidc-enabled":true,"oidc-idp-name":"sex-haver","oidc-issuer":"whoknows","oidc-link-existing":true,"oidc-scopes":["read","write"],"oidc-skip-verification":true,"password":"","path":"","port":6969,"protocol":"http","smtp-from":"queen.rip.in.piss@terfisland.org","smtp-host":"example.com","smtp-password":"hunter2","smtp-port":4269,"smtp-username":"sex-haver","software-version":"","statuses-cw-max-chars":420,"statuses-max-chars":69,"statuses-media-max-files":1,"statuses-poll-max-options":1,"statuses-poll-option-max-chars":50,"storage-backend":"local","storage-local-base-path":"/root/store","storage-s3-access-key":"minio","storage-s3-bucket":"gts","storage-s3-endpoint":"localhost:9000","storage-s3-proxy":true,"storage-s3-secret-key":"miniostorage","storage-s3-use-ssl":false,"syslog-address":"127.0.0.1:6969","syslog-enabled":true,"syslog-protocol":"udp","trusted-proxies":["127.0.0.1/32","docker.host.local"],"username":"","web-asset-base-dir":"/root","web-template-base-dir":"/root"}'

# Set all the environment variables to 
# ensure that these are parsed without panic
OUTPUT=$(GTS_LOG_LEVEL='info' \
GTS_LOG_DB_QUERIES=true \
GTS_APPLICATION_NAME=gts \
GTS_LANDING_PAGE_USER=admin \
GTS_HOST=example.com \
GTS_ACCOUNT_DOMAIN='peepee' \
GTS_PROTOCOL=http \
GTS_BIND_ADDRESS='127.0.0.1' \
GTS_PORT=6969 \
GTS_TRUSTED_PROXIES='127.0.0.1/32,docker.host.local' \
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
GTS_INSTANCE_EXPOSE_PEERS=true \
GTS_INSTANCE_EXPOSE_SUSPENDED=true \
GTS_INSTANCE_EXPOSE_PUBLIC_TIMELINE=true \
GTS_INSTANCE_DELIVER_TO_SHARED_INBOXES=false \
GTS_ACCOUNTS_ALLOW_CUSTOM_CSS=true \
GTS_ACCOUNTS_REGISTRATION_OPEN=true \
GTS_ACCOUNTS_APPROVAL_REQUIRED=false \
GTS_ACCOUNTS_REASON_REQUIRED=false \
GTS_MEDIA_IMAGE_MAX_SIZE=420 \
GTS_MEDIA_VIDEO_MAX_SIZE=420 \
GTS_MEDIA_DESCRIPTION_MIN_CHARS=69 \
GTS_MEDIA_DESCRIPTION_MAX_CHARS=5000 \
GTS_MEDIA_REMOTE_CACHE_DAYS=30 \
GTS_MEDIA_EMOJI_LOCAL_MAX_SIZE=420 \
GTS_MEDIA_EMOJI_REMOTE_MAX_SIZE=420 \
GTS_STORAGE_BACKEND='local' \
GTS_STORAGE_LOCAL_BASE_PATH='/root/store' \
GTS_STORAGE_S3_ACCESS_KEY='minio' \
GTS_STORAGE_S3_SECRET_KEY='miniostorage' \
GTS_STORAGE_S3_ENDPOINT='localhost:9000' \
GTS_STORAGE_S3_USE_SSL='false' \
GTS_STORAGE_S3_PROXY='true' \
GTS_STORAGE_S3_BUCKET='gts' \
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
GTS_OIDC_LINK_EXISTING=true \
GTS_SMTP_HOST='example.com' \
GTS_SMTP_PORT=4269 \
GTS_SMTP_USERNAME='sex-haver' \
GTS_SMTP_PASSWORD='hunter2' \
GTS_SMTP_FROM='queen.rip.in.piss@terfisland.org' \
GTS_SYSLOG_ENABLED=true \
GTS_SYSLOG_PROTOCOL='udp' \
GTS_SYSLOG_ADDRESS='127.0.0.1:6969' \
GTS_ADVANCED_COOKIES_SAMESITE='strict' \
GTS_ADVANCED_RATE_LIMIT_REQUESTS=6969 \
go run ./cmd/gotosocial/... --config-path internal/config/testdata/test.yaml debug config)

OUTPUT_OUT=$(mktemp)
echo "$OUTPUT" > "$OUTPUT_OUT"

EXPECT_OUT=$(mktemp)
echo "$EXPECT" > "$EXPECT_OUT"

if ! DIFF=$(diff "$OUTPUT_OUT" "$EXPECT_OUT"); then
    echo "OUTPUT not equal EXPECTED"
    echo "$DIFF"
    exit 1
else
    echo "OK"
    exit 0
fi
