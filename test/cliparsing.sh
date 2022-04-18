#!/bin/sh

set -e

echo "STARTING CLI TESTS"

echo "TEST_1 Make sure defaults are set correctly."
TEST_1_EXPECTED='{"account-domain":"","accounts-approval-required":true,"accounts-reason-required":true,"accounts-registration-open":true,"application-name":"gotosocial","bind-address":"0.0.0.0","config-path":"","db-address":"","db-database":"gotosocial","db-password":"","db-port":5432,"db-tls-ca-cert":"","db-tls-mode":"disable","db-type":"postgres","db-user":"","help":false,"host":"","letsencrypt-cert-dir":"/gotosocial/storage/certs","letsencrypt-email-address":"","letsencrypt-enabled":false,"letsencrypt-port":80,"log-db-queries":false,"log-level":"info","media-description-max-chars":500,"media-description-min-chars":0,"media-image-max-size":2097152,"media-remote-cache-days":30,"media-video-max-size":10485760,"oidc-client-id":"","oidc-client-secret":"","oidc-enabled":false,"oidc-idp-name":"","oidc-issuer":"","oidc-scopes":["openid","profile","email","groups"],"oidc-skip-verification":false,"port":8080,"protocol":"https","smtp-from":"GoToSocial","smtp-host":"","smtp-password":"","smtp-port":0,"smtp-username":"","software-version":"","statuses-cw-max-chars":100,"statuses-max-chars":5000,"statuses-media-max-files":6,"statuses-poll-max-options":6,"statuses-poll-option-max-chars":50,"storage-backend":"local","storage-local-base-path":"/gotosocial/storage","syslog-address":"localhost:514","syslog-enabled":false,"syslog-protocol":"udp","trusted-proxies":["127.0.0.1/32"],"web-asset-base-dir":"./web/assets/","web-template-base-dir":"./web/template/"}'
TEST_1="$(go run ./cmd/gotosocial/... debug config)"
if [ "${TEST_1}" != "${TEST_1_EXPECTED}" ]; then
    echo "TEST_1 not equal TEST_1_EXPECTED"
    exit 1
else
    echo "TEST_1 OK"
fi

echo "TEST_2 Override db-address from default using cli flag."
TEST_2_EXPECTED='{"account-domain":"","accounts-approval-required":true,"accounts-reason-required":true,"accounts-registration-open":true,"application-name":"gotosocial","bind-address":"0.0.0.0","config-path":"","db-address":"some.db.address","db-database":"gotosocial","db-password":"","db-port":5432,"db-tls-ca-cert":"","db-tls-mode":"disable","db-type":"postgres","db-user":"","help":false,"host":"","letsencrypt-cert-dir":"/gotosocial/storage/certs","letsencrypt-email-address":"","letsencrypt-enabled":false,"letsencrypt-port":80,"log-db-queries":false,"log-level":"info","media-description-max-chars":500,"media-description-min-chars":0,"media-image-max-size":2097152,"media-remote-cache-days":30,"media-video-max-size":10485760,"oidc-client-id":"","oidc-client-secret":"","oidc-enabled":false,"oidc-idp-name":"","oidc-issuer":"","oidc-scopes":["openid","profile","email","groups"],"oidc-skip-verification":false,"port":8080,"protocol":"https","smtp-from":"GoToSocial","smtp-host":"","smtp-password":"","smtp-port":0,"smtp-username":"","software-version":"","statuses-cw-max-chars":100,"statuses-max-chars":5000,"statuses-media-max-files":6,"statuses-poll-max-options":6,"statuses-poll-option-max-chars":50,"storage-backend":"local","storage-local-base-path":"/gotosocial/storage","syslog-address":"localhost:514","syslog-enabled":false,"syslog-protocol":"udp","trusted-proxies":["127.0.0.1/32"],"web-asset-base-dir":"./web/assets/","web-template-base-dir":"./web/template/"}'
TEST_2="$(go run ./cmd/gotosocial/... --db-address some.db.address debug config)"
if [ "${TEST_2}" != "${TEST_2_EXPECTED}" ]; then
    echo "TEST_2 not equal TEST_2_EXPECTED"
    exit 1
else
    echo "TEST_2 OK"
fi

echo "TEST_3 Override db-address from default using env var."
TEST_3_EXPECTED='{"account-domain":"","accounts-approval-required":true,"accounts-reason-required":true,"accounts-registration-open":true,"application-name":"gotosocial","bind-address":"0.0.0.0","config-path":"","db-address":"some.db.address","db-database":"gotosocial","db-password":"","db-port":5432,"db-tls-ca-cert":"","db-tls-mode":"disable","db-type":"postgres","db-user":"","help":false,"host":"","letsencrypt-cert-dir":"/gotosocial/storage/certs","letsencrypt-email-address":"","letsencrypt-enabled":false,"letsencrypt-port":80,"log-db-queries":false,"log-level":"info","media-description-max-chars":500,"media-description-min-chars":0,"media-image-max-size":2097152,"media-remote-cache-days":30,"media-video-max-size":10485760,"oidc-client-id":"","oidc-client-secret":"","oidc-enabled":false,"oidc-idp-name":"","oidc-issuer":"","oidc-scopes":["openid","profile","email","groups"],"oidc-skip-verification":false,"port":8080,"protocol":"https","smtp-from":"GoToSocial","smtp-host":"","smtp-password":"","smtp-port":0,"smtp-username":"","software-version":"","statuses-cw-max-chars":100,"statuses-max-chars":5000,"statuses-media-max-files":6,"statuses-poll-max-options":6,"statuses-poll-option-max-chars":50,"storage-backend":"local","storage-local-base-path":"/gotosocial/storage","syslog-address":"localhost:514","syslog-enabled":false,"syslog-protocol":"udp","trusted-proxies":["127.0.0.1/32"],"web-asset-base-dir":"./web/assets/","web-template-base-dir":"./web/template/"}'
TEST_3="$(GTS_DB_ADDRESS=some.db.address go run ./cmd/gotosocial/... debug config)"
if [ "${TEST_3}" != "${TEST_3_EXPECTED}" ]; then
    echo "TEST_3 not equal TEST_3_EXPECTED"
    exit 1
else
    echo "TEST_3 OK"
fi

echo "TEST_4 Override db-address from default using both env var and cli flag. The cli flag should take priority."
TEST_4_EXPECTED='{"account-domain":"","accounts-approval-required":true,"accounts-reason-required":true,"accounts-registration-open":true,"application-name":"gotosocial","bind-address":"0.0.0.0","config-path":"","db-address":"some.other.db.address","db-database":"gotosocial","db-password":"","db-port":5432,"db-tls-ca-cert":"","db-tls-mode":"disable","db-type":"postgres","db-user":"","help":false,"host":"","letsencrypt-cert-dir":"/gotosocial/storage/certs","letsencrypt-email-address":"","letsencrypt-enabled":false,"letsencrypt-port":80,"log-db-queries":false,"log-level":"info","media-description-max-chars":500,"media-description-min-chars":0,"media-image-max-size":2097152,"media-remote-cache-days":30,"media-video-max-size":10485760,"oidc-client-id":"","oidc-client-secret":"","oidc-enabled":false,"oidc-idp-name":"","oidc-issuer":"","oidc-scopes":["openid","profile","email","groups"],"oidc-skip-verification":false,"port":8080,"protocol":"https","smtp-from":"GoToSocial","smtp-host":"","smtp-password":"","smtp-port":0,"smtp-username":"","software-version":"","statuses-cw-max-chars":100,"statuses-max-chars":5000,"statuses-media-max-files":6,"statuses-poll-max-options":6,"statuses-poll-option-max-chars":50,"storage-backend":"local","storage-local-base-path":"/gotosocial/storage","syslog-address":"localhost:514","syslog-enabled":false,"syslog-protocol":"udp","trusted-proxies":["127.0.0.1/32"],"web-asset-base-dir":"./web/assets/","web-template-base-dir":"./web/template/"}'
TEST_4="$(GTS_DB_ADDRESS=some.db.address go run ./cmd/gotosocial/... --db-address some.other.db.address debug config)"
if [ "${TEST_4}" != "${TEST_4_EXPECTED}" ]; then
    echo "TEST_4 not equal TEST_4_EXPECTED"
    exit 1
else
    echo "TEST_4 OK"
fi

echo "TEST_5 Test loading a config file by passing an env var."
TEST_5_EXPECTED='{"account-domain":"example.org","accounts-approval-required":true,"accounts-reason-required":true,"accounts-registration-open":true,"application-name":"gotosocial","bind-address":"0.0.0.0","config-path":"./test/test.yaml","db-address":"127.0.0.1","db-database":"postgres","db-password":"postgres","db-port":5432,"db-tls-ca-cert":"","db-tls-mode":"disable","db-type":"postgres","db-user":"postgres","help":false,"host":"gts.example.org","letsencrypt-cert-dir":"/gotosocial/storage/certs","letsencrypt-email-address":"","letsencrypt-enabled":true,"letsencrypt-port":80,"log-db-queries":false,"log-level":"info","media-description-max-chars":500,"media-description-min-chars":0,"media-image-max-size":2097152,"media-remote-cache-days":30,"media-video-max-size":10485760,"oidc-client-id":"","oidc-client-secret":"","oidc-enabled":false,"oidc-idp-name":"","oidc-issuer":"","oidc-scopes":["openid","email","profile","groups"],"oidc-skip-verification":false,"port":8080,"protocol":"https","smtp-from":"someone@example.org","smtp-host":"verycoolemailhost.mail","smtp-password":"smtp-password","smtp-port":8888,"smtp-username":"smtp-username","software-version":"","statuses-cw-max-chars":100,"statuses-max-chars":5000,"statuses-media-max-files":6,"statuses-poll-max-options":6,"statuses-poll-option-max-chars":50,"storage-backend":"local","storage-local-base-path":"/gotosocial/storage","syslog-address":"localhost:514","syslog-enabled":false,"syslog-protocol":"udp","trusted-proxies":["127.0.0.1/32","0.0.0.0/0"],"web-asset-base-dir":"./web/assets/","web-template-base-dir":"./web/template/"}'
TEST_5="$(GTS_CONFIG_PATH=./test/test.yaml go run ./cmd/gotosocial/... debug config)"
if [ "${TEST_5}" != "${TEST_5_EXPECTED}" ]; then
    echo "TEST_5 not equal TEST_5_EXPECTED"
    exit 1
else
    echo "TEST_5 OK"
fi

echo "TEST_6 Test loading a config file by passing cli flag."
TEST_6_EXPECTED='{"account-domain":"example.org","accounts-approval-required":true,"accounts-reason-required":true,"accounts-registration-open":true,"application-name":"gotosocial","bind-address":"0.0.0.0","config-path":"./test/test.yaml","db-address":"127.0.0.1","db-database":"postgres","db-password":"postgres","db-port":5432,"db-tls-ca-cert":"","db-tls-mode":"disable","db-type":"postgres","db-user":"postgres","help":false,"host":"gts.example.org","letsencrypt-cert-dir":"/gotosocial/storage/certs","letsencrypt-email-address":"","letsencrypt-enabled":true,"letsencrypt-port":80,"log-db-queries":false,"log-level":"info","media-description-max-chars":500,"media-description-min-chars":0,"media-image-max-size":2097152,"media-remote-cache-days":30,"media-video-max-size":10485760,"oidc-client-id":"","oidc-client-secret":"","oidc-enabled":false,"oidc-idp-name":"","oidc-issuer":"","oidc-scopes":["openid","email","profile","groups"],"oidc-skip-verification":false,"port":8080,"protocol":"https","smtp-from":"someone@example.org","smtp-host":"verycoolemailhost.mail","smtp-password":"smtp-password","smtp-port":8888,"smtp-username":"smtp-username","software-version":"","statuses-cw-max-chars":100,"statuses-max-chars":5000,"statuses-media-max-files":6,"statuses-poll-max-options":6,"statuses-poll-option-max-chars":50,"storage-backend":"local","storage-local-base-path":"/gotosocial/storage","syslog-address":"localhost:514","syslog-enabled":false,"syslog-protocol":"udp","trusted-proxies":["127.0.0.1/32","0.0.0.0/0"],"web-asset-base-dir":"./web/assets/","web-template-base-dir":"./web/template/"}'
TEST_6="$(go run ./cmd/gotosocial/... --config-path ./test/test.yaml debug config)"
if [ "${TEST_6}" != "${TEST_6_EXPECTED}" ]; then
    echo "TEST_6 not equal TEST_6_EXPECTED"
    exit 1
else
    echo "TEST_6 OK"
fi

echo "TEST_7 Test loading a config file and overriding one of the variables with a cli flag."
TEST_7_EXPECTED='{"account-domain":"","accounts-approval-required":true,"accounts-reason-required":true,"accounts-registration-open":true,"application-name":"gotosocial","bind-address":"0.0.0.0","config-path":"./test/test.yaml","db-address":"127.0.0.1","db-database":"postgres","db-password":"postgres","db-port":5432,"db-tls-ca-cert":"","db-tls-mode":"disable","db-type":"postgres","db-user":"postgres","help":false,"host":"gts.example.org","letsencrypt-cert-dir":"/gotosocial/storage/certs","letsencrypt-email-address":"","letsencrypt-enabled":true,"letsencrypt-port":80,"log-db-queries":false,"log-level":"info","media-description-max-chars":500,"media-description-min-chars":0,"media-image-max-size":2097152,"media-remote-cache-days":30,"media-video-max-size":10485760,"oidc-client-id":"","oidc-client-secret":"","oidc-enabled":false,"oidc-idp-name":"","oidc-issuer":"","oidc-scopes":["openid","email","profile","groups"],"oidc-skip-verification":false,"port":8080,"protocol":"https","smtp-from":"someone@example.org","smtp-host":"verycoolemailhost.mail","smtp-password":"smtp-password","smtp-port":8888,"smtp-username":"smtp-username","software-version":"","statuses-cw-max-chars":100,"statuses-max-chars":5000,"statuses-media-max-files":6,"statuses-poll-max-options":6,"statuses-poll-option-max-chars":50,"storage-backend":"local","storage-local-base-path":"/gotosocial/storage","syslog-address":"localhost:514","syslog-enabled":false,"syslog-protocol":"udp","trusted-proxies":["127.0.0.1/32","0.0.0.0/0"],"web-asset-base-dir":"./web/assets/","web-template-base-dir":"./web/template/"}'
TEST_7="$(go run ./cmd/gotosocial/... --config-path ./test/test.yaml --account-domain '' debug config)"
if [ "${TEST_7}" != "${TEST_7_EXPECTED}" ]; then
    echo "TEST_7 not equal TEST_7_EXPECTED"
    exit 1
else
    echo "TEST_7 OK"
fi

echo "TEST_8 Test loading a config file and overriding one of the variables with an env var."
TEST_8_EXPECTED='{"account-domain":"peepee","accounts-approval-required":true,"accounts-reason-required":true,"accounts-registration-open":true,"application-name":"gotosocial","bind-address":"0.0.0.0","config-path":"./test/test.yaml","db-address":"127.0.0.1","db-database":"postgres","db-password":"postgres","db-port":5432,"db-tls-ca-cert":"","db-tls-mode":"disable","db-type":"postgres","db-user":"postgres","help":false,"host":"gts.example.org","letsencrypt-cert-dir":"/gotosocial/storage/certs","letsencrypt-email-address":"","letsencrypt-enabled":true,"letsencrypt-port":80,"log-db-queries":false,"log-level":"info","media-description-max-chars":500,"media-description-min-chars":0,"media-image-max-size":2097152,"media-remote-cache-days":30,"media-video-max-size":10485760,"oidc-client-id":"","oidc-client-secret":"","oidc-enabled":false,"oidc-idp-name":"","oidc-issuer":"","oidc-scopes":["openid","email","profile","groups"],"oidc-skip-verification":false,"port":8080,"protocol":"https","smtp-from":"someone@example.org","smtp-host":"verycoolemailhost.mail","smtp-password":"smtp-password","smtp-port":8888,"smtp-username":"smtp-username","software-version":"","statuses-cw-max-chars":100,"statuses-max-chars":5000,"statuses-media-max-files":6,"statuses-poll-max-options":6,"statuses-poll-option-max-chars":50,"storage-backend":"local","storage-local-base-path":"/gotosocial/storage","syslog-address":"localhost:514","syslog-enabled":false,"syslog-protocol":"udp","trusted-proxies":["127.0.0.1/32","0.0.0.0/0"],"web-asset-base-dir":"./web/assets/","web-template-base-dir":"./web/template/"}'
TEST_8="$(GTS_ACCOUNT_DOMAIN='peepee' go run ./cmd/gotosocial/... --config-path ./test/test.yaml debug config)"
if [ "${TEST_8}" != "${TEST_8_EXPECTED}" ]; then
    echo "TEST_8 not equal TEST_8_EXPECTED"
    exit 1
else
    echo "TEST_8 OK"
fi

echo "TEST_9 Test loading a config file and overriding one of the variables with both an env var and a cli flag. The cli flag should have priority."
TEST_9_EXPECTED='{"account-domain":"","accounts-approval-required":true,"accounts-reason-required":true,"accounts-registration-open":true,"application-name":"gotosocial","bind-address":"0.0.0.0","config-path":"./test/test.yaml","db-address":"127.0.0.1","db-database":"postgres","db-password":"postgres","db-port":5432,"db-tls-ca-cert":"","db-tls-mode":"disable","db-type":"postgres","db-user":"postgres","help":false,"host":"gts.example.org","letsencrypt-cert-dir":"/gotosocial/storage/certs","letsencrypt-email-address":"","letsencrypt-enabled":true,"letsencrypt-port":80,"log-db-queries":false,"log-level":"info","media-description-max-chars":500,"media-description-min-chars":0,"media-image-max-size":2097152,"media-remote-cache-days":30,"media-video-max-size":10485760,"oidc-client-id":"","oidc-client-secret":"","oidc-enabled":false,"oidc-idp-name":"","oidc-issuer":"","oidc-scopes":["openid","email","profile","groups"],"oidc-skip-verification":false,"port":8080,"protocol":"https","smtp-from":"someone@example.org","smtp-host":"verycoolemailhost.mail","smtp-password":"smtp-password","smtp-port":8888,"smtp-username":"smtp-username","software-version":"","statuses-cw-max-chars":100,"statuses-max-chars":5000,"statuses-media-max-files":6,"statuses-poll-max-options":6,"statuses-poll-option-max-chars":50,"storage-backend":"local","storage-local-base-path":"/gotosocial/storage","syslog-address":"localhost:514","syslog-enabled":false,"syslog-protocol":"udp","trusted-proxies":["127.0.0.1/32","0.0.0.0/0"],"web-asset-base-dir":"./web/assets/","web-template-base-dir":"./web/template/"}'
TEST_9="$(GTS_ACCOUNT_DOMAIN='peepee' go run ./cmd/gotosocial/... --config-path ./test/test.yaml --account-domain '' debug config)"
if [ "${TEST_9}" != "${TEST_9_EXPECTED}" ]; then
    echo "TEST_9 not equal TEST_9_EXPECTED"
    exit 1
else
    echo "TEST_9 OK"
fi

echo "TEST_10 Test loading a config file from json."
TEST_10_EXPECTED='{"account-domain":"example.org","accounts-approval-required":true,"accounts-reason-required":true,"accounts-registration-open":true,"application-name":"gotosocial","bind-address":"0.0.0.0","config-path":"./test/test.json","db-address":"127.0.0.1","db-database":"postgres","db-password":"postgres","db-port":5432,"db-tls-ca-cert":"","db-tls-mode":"disable","db-type":"postgres","db-user":"postgres","help":false,"host":"gts.example.org","letsencrypt-cert-dir":"/gotosocial/storage/certs","letsencrypt-email-address":"","letsencrypt-enabled":true,"letsencrypt-port":80,"log-db-queries":false,"log-level":"info","media-description-max-chars":500,"media-description-min-chars":0,"media-image-max-size":2097152,"media-remote-cache-days":30,"media-video-max-size":10485760,"oidc-client-id":"","oidc-client-secret":"","oidc-enabled":false,"oidc-idp-name":"","oidc-issuer":"","oidc-scopes":["openid","email","profile","groups"],"oidc-skip-verification":false,"port":8080,"protocol":"https","smtp-from":"someone@example.org","smtp-host":"verycoolemailhost.mail","smtp-password":"smtp-password","smtp-port":8888,"smtp-username":"smtp-username","software-version":"","statuses-cw-max-chars":100,"statuses-max-chars":5000,"statuses-media-max-files":6,"statuses-poll-max-options":6,"statuses-poll-option-max-chars":50,"storage-backend":"local","storage-local-base-path":"/gotosocial/storage","syslog-address":"localhost:514","syslog-enabled":false,"syslog-protocol":"udp","trusted-proxies":["127.0.0.1/32","0.0.0.0/0"],"web-asset-base-dir":"./web/assets/","web-template-base-dir":"./web/template/"}'
TEST_10="$(go run ./cmd/gotosocial/... --config-path ./test/test.json debug config)"
if [ "${TEST_10}" != "${TEST_10_EXPECTED}" ]; then
    echo "TEST_10 not equal TEST_10_EXPECTED"
    exit 1
else
    echo "TEST_10 OK"
fi

echo "TEST_11 Test loading a partial config file. Default values should be used apart from those set in the config file."
TEST_11_EXPECTED='{"account-domain":"peepee.poopoo","accounts-approval-required":true,"accounts-reason-required":true,"accounts-registration-open":true,"application-name":"gotosocial","bind-address":"0.0.0.0","config-path":"./test/test2.yaml","db-address":"","db-database":"gotosocial","db-password":"","db-port":5432,"db-tls-ca-cert":"","db-tls-mode":"disable","db-type":"postgres","db-user":"","help":false,"host":"","letsencrypt-cert-dir":"/gotosocial/storage/certs","letsencrypt-email-address":"","letsencrypt-enabled":false,"letsencrypt-port":80,"log-db-queries":false,"log-level":"trace","media-description-max-chars":500,"media-description-min-chars":0,"media-image-max-size":2097152,"media-remote-cache-days":30,"media-video-max-size":10485760,"oidc-client-id":"","oidc-client-secret":"","oidc-enabled":false,"oidc-idp-name":"","oidc-issuer":"","oidc-scopes":["openid","profile","email","groups"],"oidc-skip-verification":false,"port":8080,"protocol":"https","smtp-from":"GoToSocial","smtp-host":"","smtp-password":"","smtp-port":0,"smtp-username":"","software-version":"","statuses-cw-max-chars":100,"statuses-max-chars":5000,"statuses-media-max-files":6,"statuses-poll-max-options":6,"statuses-poll-option-max-chars":50,"storage-backend":"local","storage-local-base-path":"/gotosocial/storage","syslog-address":"localhost:514","syslog-enabled":false,"syslog-protocol":"udp","trusted-proxies":["127.0.0.1/32"],"web-asset-base-dir":"./web/assets/","web-template-base-dir":"./web/template/"}'
TEST_11="$(go run ./cmd/gotosocial/... --config-path ./test/test2.yaml debug config)"
if [ "${TEST_11}" != "${TEST_11_EXPECTED}" ]; then
    echo "TEST_11 not equal TEST_11_EXPECTED"
    exit 1
else
    echo "TEST_11 OK"
fi

echo "FINISHED CLI TESTS"
