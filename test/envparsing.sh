#!/bin/sh

set -eu

EXPECT=$(cat << "EOF"
{
    "account-domain": "peepee",
    "accounts-allow-custom-css": true,
    "accounts-custom-css-length": 5000,
    "accounts-reason-required": false,
    "accounts-registration-backlog-limit": 100,
    "accounts-registration-daily-limit": 50,
    "accounts-registration-open": true,
    "advanced-cookies-samesite": "strict",
    "advanced-csp-extra-uris": [],
    "advanced-header-filter-mode": "block",
    "advanced-rate-limit-exceptions": [
        "192.0.2.0/24",
        "127.0.0.1/32"
    ],
    "advanced-rate-limit-requests": 6969,
    "advanced-sender-multiplier": -1,
    "advanced-throttling-multiplier": -1,
    "advanced-throttling-retry-after": 10000000000,
    "application-name": "gts",
    "bind-address": "127.0.0.1",
    "cache": {
        "account-mem-ratio": 5,
        "account-note-mem-ratio": 1,
        "account-settings-mem-ratio": 0.1,
        "account-stats-mem-ratio": 2,
        "application-mem-ratio": 0.1,
        "block-ids-mem-ratio": 3,
        "block-mem-ratio": 2,
        "boost-of-ids-mem-ratio": 3,
        "client-mem-ratio": 0.1,
        "conversation-last-status-ids-mem-ratio": 2,
        "conversation-mem-ratio": 1,
        "domain-permission-draft-mem-ratio": 0.5,
        "domain-permission-subscription-mem-ratio": 0.5,
        "emoji-category-mem-ratio": 0.1,
        "emoji-mem-ratio": 3,
        "filter-keyword-mem-ratio": 0.5,
        "filter-mem-ratio": 0.5,
        "filter-status-mem-ratio": 0.5,
        "follow-ids-mem-ratio": 4,
        "follow-mem-ratio": 2,
        "follow-request-ids-mem-ratio": 2,
        "follow-request-mem-ratio": 2,
        "following-tag-ids-mem-ratio": 2,
        "in-reply-to-ids-mem-ratio": 3,
        "instance-mem-ratio": 1,
        "interaction-request-mem-ratio": 1,
        "list-ids-mem-ratio": 2,
        "list-mem-ratio": 1,
        "listed-ids-mem-ratio": 2,
        "marker-mem-ratio": 0.5,
        "media-mem-ratio": 4,
        "memory-target": 104857600,
        "mention-mem-ratio": 2,
        "move-mem-ratio": 0.1,
        "notification-mem-ratio": 2,
        "poll-mem-ratio": 1,
        "poll-vote-ids-mem-ratio": 2,
        "poll-vote-mem-ratio": 2,
        "report-mem-ratio": 1,
        "sin-bin-status-mem-ratio": 0.5,
        "status-bookmark-ids-mem-ratio": 2,
        "status-bookmark-mem-ratio": 0.5,
        "status-edit-mem-ratio": 2,
        "status-fave-ids-mem-ratio": 3,
        "status-fave-mem-ratio": 2,
        "status-mem-ratio": 5,
        "tag-mem-ratio": 2,
        "thread-mute-mem-ratio": 0.2,
        "token-mem-ratio": 0.75,
        "tombstone-mem-ratio": 0.5,
        "user-mem-ratio": 0.25,
        "user-mute-ids-mem-ratio": 3,
        "user-mute-mem-ratio": 2,
        "visibility-mem-ratio": 2,
        "web-push-subscription-ids-mem-ratio": 1,
        "web-push-subscription-mem-ratio": 1,
        "webfinger-mem-ratio": 0.1
    },
    "config-path": "internal/config/testdata/test.yaml",
    "db-address": ":memory:",
    "db-database": "gotosocial_prod",
    "db-max-open-conns-multiplier": 3,
    "db-password": "hunter2",
    "db-port": 6969,
    "db-postgres-connection-string": "",
    "db-sqlite-busy-timeout": 1000000000,
    "db-sqlite-cache-size": 0,
    "db-sqlite-journal-mode": "DELETE",
    "db-sqlite-synchronous": "FULL",
    "db-tls-ca-cert": "",
    "db-tls-mode": "disable",
    "db-type": "sqlite",
    "db-user": "sex-haver",
    "dry-run": true,
    "email": "",
    "host": "example.com",
    "http-client": {
        "allow-ips": [],
        "block-ips": [],
        "timeout": 30000000000,
        "tls-insecure-skip-verify": false
    },
    "instance-allow-backdating-statuses": true,
    "instance-deliver-to-shared-inboxes": false,
    "instance-expose-peers": true,
    "instance-expose-public-timeline": true,
    "instance-expose-suspended": true,
    "instance-expose-suspended-web": true,
    "instance-federation-mode": "allowlist",
    "instance-federation-spam-filter": true,
    "instance-inject-mastodon-version": true,
    "instance-languages": [
        "nl",
        "en-GB"
    ],
    "instance-stats-mode": "baffle",
    "instance-subscriptions-process-every": 86400000000000,
    "instance-subscriptions-process-from": "23:00",
    "landing-page-user": "admin",
    "letsencrypt-cert-dir": "/gotosocial/storage/certs",
    "letsencrypt-email-address": "",
    "letsencrypt-enabled": true,
    "letsencrypt-port": 80,
    "local-only": false,
    "log-client-ip": false,
    "log-db-queries": true,
    "log-level": "info",
    "log-timestamp-format": "banana",
    "media-cleanup-every": 86400000000000,
    "media-cleanup-from": "00:00",
    "media-description-max-chars": 5000,
    "media-description-min-chars": 69,
    "media-emoji-local-max-size": 420,
    "media-emoji-remote-max-size": 420,
    "media-ffmpeg-pool-size": 8,
    "media-image-size-hint": 5242880,
    "media-local-max-size": 420,
    "media-remote-cache-days": 30,
    "media-remote-max-size": 420,
    "media-video-size-hint": 41943040,
    "metrics-auth-enabled": false,
    "metrics-auth-password": "",
    "metrics-auth-username": "",
    "metrics-enabled": false,
    "oidc-admin-groups": [
        "steamy"
    ],
    "oidc-allowed-groups": [
        "sloths"
    ],
    "oidc-client-id": "1234",
    "oidc-client-secret": "shhhh its a secret",
    "oidc-enabled": true,
    "oidc-idp-name": "sex-haver",
    "oidc-issuer": "whoknows",
    "oidc-link-existing": true,
    "oidc-scopes": [
        "read",
        "write"
    ],
    "oidc-skip-verification": true,
    "password": "",
    "path": "",
    "port": 6969,
    "protocol": "http",
    "remote-only": false,
    "request-id-header": "X-Trace-Id",
    "smtp-disclose-recipients": true,
    "smtp-from": "queen.rip.in.piss@terfisland.org",
    "smtp-host": "example.com",
    "smtp-password": "hunter2",
    "smtp-port": 4269,
    "smtp-username": "sex-haver",
    "software-version": "",
    "statuses-max-chars": 69,
    "statuses-media-max-files": 1,
    "statuses-poll-max-options": 1,
    "statuses-poll-option-max-chars": 50,
    "storage-backend": "local",
    "storage-local-base-path": "/root/store",
    "storage-s3-access-key": "minio",
    "storage-s3-bucket": "gts",
    "storage-s3-endpoint": "localhost:9000",
    "storage-s3-proxy": true,
    "storage-s3-redirect-url": "",
    "storage-s3-secret-key": "miniostorage",
    "storage-s3-use-ssl": false,
    "syslog-address": "127.0.0.1:6969",
    "syslog-enabled": true,
    "syslog-protocol": "udp",
    "tls-certificate-chain": "",
    "tls-certificate-key": "",
    "tracing-enabled": false,
    "tracing-endpoint": "localhost:4317",
    "tracing-insecure-transport": true,
    "tracing-transport": "grpc",
    "trusted-proxies": [
        "127.0.0.1/32",
        "docker.host.local"
    ],
    "username": "",
    "web-asset-base-dir": "/root",
    "web-template-base-dir": "/root"
}
EOF
)

# Set all the environment variables to 
# ensure that these are parsed without panic
OUTPUT=$(GTS_LOG_LEVEL='info' \
GTS_LOG_TIMESTAMP_FORMAT="banana" \
GTS_LOG_DB_QUERIES=true \
GTS_LOG_CLIENT_IP=false \
GTS_APPLICATION_NAME=gts \
GTS_LANDING_PAGE_USER=admin \
GTS_HOST=example.com \
GTS_ACCOUNT_DOMAIN='peepee' \
GTS_PROTOCOL=http \
GTS_BIND_ADDRESS='127.0.0.1' \
GTS_PORT=6969 \
GTS_TRUSTED_PROXIES='127.0.0.1/32,docker.host.local' \
GTS_DB_TYPE='sqlite' \
GTS_DB_POSTGRES_CONNECTION_STRING='' \
GTS_DB_ADDRESS=':memory:' \
GTS_DB_PORT=6969 \
GTS_DB_USER='sex-haver' \
GTS_DB_PASSWORD='hunter2' \
GTS_DB_DATABASE='gotosocial_prod' \
GTS_DB_MAX_OPEN_CONNS_MULTIPLIER=3 \
GTS_DB_SQLITE_JOURNAL_MODE='DELETE' \
GTS_DB_SQLITE_SYNCHRONOUS='FULL' \
GTS_DB_SQLITE_CACHE_SIZE=0 \
GTS_DB_SQLITE_BUSY_TIMEOUT='1s' \
GTS_TLS_MODE='' \
GTS_DB_TLS_CA_CERT='' \
GTS_WEB_TEMPLATE_BASE_DIR='/root' \
GTS_WEB_ASSET_BASE_DIR='/root' \
GTS_INSTANCE_EXPOSE_PEERS=true \
GTS_INSTANCE_EXPOSE_SUSPENDED=true \
GTS_INSTANCE_EXPOSE_SUSPENDED_WEB=true \
GTS_INSTANCE_EXPOSE_PUBLIC_TIMELINE=true \
GTS_INSTANCE_FEDERATION_MODE='allowlist' \
GTS_INSTANCE_FEDERATION_SPAM_FILTER=true \
GTS_INSTANCE_DELIVER_TO_SHARED_INBOXES=false \
GTS_INSTANCE_INJECT_MASTODON_VERSION=true \
GTS_INSTANCE_LANGUAGES="nl,en-gb" \
GTS_INSTANCE_STATS_MODE="baffle" \
GTS_ACCOUNTS_ALLOW_CUSTOM_CSS=true \
GTS_ACCOUNTS_CUSTOM_CSS_LENGTH=5000 \
GTS_ACCOUNTS_REGISTRATION_BACKLOG_LIMIT=100 \
GTS_ACCOUNTS_REGISTRATION_DAILY_LIMIT=50 \
GTS_ACCOUNTS_REGISTRATION_OPEN=true \
GTS_ACCOUNTS_REASON_REQUIRED=false \
GTS_MEDIA_DESCRIPTION_MIN_CHARS=69 \
GTS_MEDIA_DESCRIPTION_MAX_CHARS=5000 \
GTS_MEDIA_IMAGE_SIZE_HINT='5MiB' \
GTS_MEDIA_LOCAL_MAX_SIZE=420 \
GTS_MEDIA_REMOTE_MAX_SIZE=420 \
GTS_MEDIA_REMOTE_CACHE_DAYS=30 \
GTS_MEDIA_EMOJI_LOCAL_MAX_SIZE=420 \
GTS_MEDIA_EMOJI_REMOTE_MAX_SIZE=420 \
GTS_MEDIA_FFMPEG_POOL_SIZE=8 \
GTS_MEDIA_VIDEO_SIZE_HINT='40MiB' \
GTS_METRICS_AUTH_ENABLED=false \
GTS_METRICS_ENABLED=false \
GTS_STORAGE_BACKEND='local' \
GTS_STORAGE_LOCAL_BASE_PATH='/root/store' \
GTS_STORAGE_S3_ACCESS_KEY='minio' \
GTS_STORAGE_S3_SECRET_KEY='miniostorage' \
GTS_STORAGE_S3_ENDPOINT='localhost:9000' \
GTS_STORAGE_S3_USE_SSL='false' \
GTS_STORAGE_S3_PROXY='true' \
GTS_STORAGE_S3_REDIRECT_URL='' \
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
GTS_OIDC_ALLOWED_GROUPS='sloths' \
GTS_OIDC_ADMIN_GROUPS='steamy' \
GTS_SMTP_HOST='example.com' \
GTS_SMTP_PORT=4269 \
GTS_SMTP_USERNAME='sex-haver' \
GTS_SMTP_PASSWORD='hunter2' \
GTS_SMTP_FROM='queen.rip.in.piss@terfisland.org' \
GTS_SMTP_DISCLOSE_RECIPIENTS=true \
GTS_SYSLOG_ENABLED=true \
GTS_SYSLOG_PROTOCOL='udp' \
GTS_SYSLOG_ADDRESS='127.0.0.1:6969' \
GTS_TRACING_ENDPOINT='localhost:4317' \
GTS_TRACING_INSECURE_TRANSPORT=true \
GTS_ADVANCED_COOKIES_SAMESITE='strict' \
GTS_ADVANCED_RATE_LIMIT_EXCEPTIONS="192.0.2.0/24,127.0.0.1/32" \
GTS_ADVANCED_RATE_LIMIT_REQUESTS=6969 \
GTS_ADVANCED_SENDER_MULTIPLIER=-1 \
GTS_ADVANCED_THROTTLING_MULTIPLIER=-1 \
GTS_ADVANCED_THROTTLING_RETRY_AFTER='10s' \
GTS_ADVANCED_HEADER_FILTER_MODE='block' \
GTS_REQUEST_ID_HEADER='X-Trace-Id' \
go run ./cmd/gotosocial/... --config-path internal/config/testdata/test.yaml debug config)

OUTPUT_OUT=$(mktemp)
echo "$OUTPUT" > "$OUTPUT_OUT"

EXPECT_OUT=$(mktemp)
echo "$EXPECT" > "$EXPECT_OUT"

DIFFCMD=$(command -v diff 2>&1)
if command -v jd >/dev/null 2>&1; then
    DIFFCMD=$(command -v jd 2>&1)
fi

if ! DIFF=$("$DIFFCMD" "$OUTPUT_OUT" "$EXPECT_OUT"); then
    echo "OUTPUT not equal EXPECTED"
    echo "$DIFF"
    exit 1
else
    echo "OK"
    exit 0
fi
