#!/bin/sh

set -eu

echo "STARTING CLI TESTS"

# TEST_1
# Make sure defaults are set correctly.
TEST_1_EXPECTED='{"account-domain":"","accounts-approval-required":true,"accounts-open-registration":true,"accounts-reason-required":true,"application-name":"gotosocial","asset-basedir":"./web/assets/","bind-address":"0.0.0.0","config-path":"","db-address":"localhost","db-database":"postgres","db-password":"postgres","db-port":5432,"db-tls-ca-cert":"","db-tls-mode":"disable","db-type":"postgres","db-user":"postgres","help":false,"host":"","letsencrypt-cert-dir":"/gotosocial/storage/certs","letsencrypt-email":"","letsencrypt-enabled":true,"letsencrypt-port":80,"log-level":"info","media-max-description-chars":500,"media-max-image-size":2097152,"media-max-video-size":10485760,"media-min-description-chars":0,"oidc-client-id":"","oidc-client-secret":"","oidc-enabled":false,"oidc-idp-name":"","oidc-issuer":"","oidc-scopes":["openid","profile","email","groups"],"oidc-skip-verification":false,"port":8080,"protocol":"https","smtp-from":"GoToSocial","smtp-host":"","smtp-password":"","smtp-port":0,"smtp-username":"","software-version":"","statuses-cw-max-chars":100,"statuses-max-chars":5000,"statuses-max-media-files":6,"statuses-poll-max-options":6,"statuses-poll-option-max-chars":50,"storage-backend":"local","storage-base-path":"/gotosocial/storage","storage-serve-base-path":"/fileserver","storage-serve-host":"localhost","storage-serve-protocol":"https","template-basedir":"./web/template/","trusted-proxies":["127.0.0.1/32"]}'
TEST_1="$(go run ./cmd/gotosocial/... debug config)"
if [ "${TEST_1}" != "${TEST_1_EXPECTED}" ]; then
    echo "TEST_1 not equal TEST_1_EXPECTED"
    exit 1
else
    echo "TEST_1 OK"
fi

# TEST_2
# Override db-address from default using cli flag.
TEST_2_EXPECTED='{"account-domain":"","accounts-approval-required":true,"accounts-open-registration":true,"accounts-reason-required":true,"application-name":"gotosocial","asset-basedir":"./web/assets/","bind-address":"0.0.0.0","config-path":"","db-address":"some.db.address","db-database":"postgres","db-password":"postgres","db-port":5432,"db-tls-ca-cert":"","db-tls-mode":"disable","db-type":"postgres","db-user":"postgres","help":false,"host":"","letsencrypt-cert-dir":"/gotosocial/storage/certs","letsencrypt-email":"","letsencrypt-enabled":true,"letsencrypt-port":80,"log-level":"info","media-max-description-chars":500,"media-max-image-size":2097152,"media-max-video-size":10485760,"media-min-description-chars":0,"oidc-client-id":"","oidc-client-secret":"","oidc-enabled":false,"oidc-idp-name":"","oidc-issuer":"","oidc-scopes":["openid","profile","email","groups"],"oidc-skip-verification":false,"port":8080,"protocol":"https","smtp-from":"GoToSocial","smtp-host":"","smtp-password":"","smtp-port":0,"smtp-username":"","software-version":"","statuses-cw-max-chars":100,"statuses-max-chars":5000,"statuses-max-media-files":6,"statuses-poll-max-options":6,"statuses-poll-option-max-chars":50,"storage-backend":"local","storage-base-path":"/gotosocial/storage","storage-serve-base-path":"/fileserver","storage-serve-host":"localhost","storage-serve-protocol":"https","template-basedir":"./web/template/","trusted-proxies":["127.0.0.1/32"]}'
TEST_2="$(go run ./cmd/gotosocial/... --db-address some.db.address debug config)"
if [ "${TEST_2}" != "${TEST_2_EXPECTED}" ]; then
    echo "TEST_2 not equal TEST_2_EXPECTED"
    exit 1
else
    echo "TEST_2 OK"
fi

# TEST_3
# Override db-address from default using env var.
TEST_3_EXPECTED='{"account-domain":"","accounts-approval-required":true,"accounts-open-registration":true,"accounts-reason-required":true,"application-name":"gotosocial","asset-basedir":"./web/assets/","bind-address":"0.0.0.0","config-path":"","db-address":"some.db.address","db-database":"postgres","db-password":"postgres","db-port":5432,"db-tls-ca-cert":"","db-tls-mode":"disable","db-type":"postgres","db-user":"postgres","help":false,"host":"","letsencrypt-cert-dir":"/gotosocial/storage/certs","letsencrypt-email":"","letsencrypt-enabled":true,"letsencrypt-port":80,"log-level":"info","media-max-description-chars":500,"media-max-image-size":2097152,"media-max-video-size":10485760,"media-min-description-chars":0,"oidc-client-id":"","oidc-client-secret":"","oidc-enabled":false,"oidc-idp-name":"","oidc-issuer":"","oidc-scopes":["openid","profile","email","groups"],"oidc-skip-verification":false,"port":8080,"protocol":"https","smtp-from":"GoToSocial","smtp-host":"","smtp-password":"","smtp-port":0,"smtp-username":"","software-version":"","statuses-cw-max-chars":100,"statuses-max-chars":5000,"statuses-max-media-files":6,"statuses-poll-max-options":6,"statuses-poll-option-max-chars":50,"storage-backend":"local","storage-base-path":"/gotosocial/storage","storage-serve-base-path":"/fileserver","storage-serve-host":"localhost","storage-serve-protocol":"https","template-basedir":"./web/template/","trusted-proxies":["127.0.0.1/32"]}'
TEST_3="$(GTS_DB_ADDRESS=some.db.address go run ./cmd/gotosocial/... debug config)"
if [ "${TEST_3}" != "${TEST_3_EXPECTED}" ]; then
    echo "TEST_3 not equal TEST_3_EXPECTED"
    exit 1
else
    echo "TEST_3 OK"
fi

# TEST_4
# Override db-address from default using both env var and cli flag.
# The cli flag should take priority.
TEST_4_EXPECTED='{"account-domain":"","accounts-approval-required":true,"accounts-open-registration":true,"accounts-reason-required":true,"application-name":"gotosocial","asset-basedir":"./web/assets/","bind-address":"0.0.0.0","config-path":"","db-address":"some.other.db.address","db-database":"postgres","db-password":"postgres","db-port":5432,"db-tls-ca-cert":"","db-tls-mode":"disable","db-type":"postgres","db-user":"postgres","help":false,"host":"","letsencrypt-cert-dir":"/gotosocial/storage/certs","letsencrypt-email":"","letsencrypt-enabled":true,"letsencrypt-port":80,"log-level":"info","media-max-description-chars":500,"media-max-image-size":2097152,"media-max-video-size":10485760,"media-min-description-chars":0,"oidc-client-id":"","oidc-client-secret":"","oidc-enabled":false,"oidc-idp-name":"","oidc-issuer":"","oidc-scopes":["openid","profile","email","groups"],"oidc-skip-verification":false,"port":8080,"protocol":"https","smtp-from":"GoToSocial","smtp-host":"","smtp-password":"","smtp-port":0,"smtp-username":"","software-version":"","statuses-cw-max-chars":100,"statuses-max-chars":5000,"statuses-max-media-files":6,"statuses-poll-max-options":6,"statuses-poll-option-max-chars":50,"storage-backend":"local","storage-base-path":"/gotosocial/storage","storage-serve-base-path":"/fileserver","storage-serve-host":"localhost","storage-serve-protocol":"https","template-basedir":"./web/template/","trusted-proxies":["127.0.0.1/32"]}'
TEST_4="$(GTS_DB_ADDRESS=some.db.address go run ./cmd/gotosocial/... --db-address some.other.db.address debug config)"
if [ "${TEST_4}" != "${TEST_4_EXPECTED}" ]; then
    echo "TEST_4 not equal TEST_4_EXPECTED"
    exit 1
else
    echo "TEST_4 OK"
fi



# TEST_5
# Test loading a config file by passing a cli flag.
TEST_5_EXPECTED='{"account-domain":"aaaaaaa","accountdomain":"aaaaaaa","accounts":{"openregistration":false,"reasonrequired":false,"requireapproval":false},"accounts-approval-required":false,"accounts-open-registration":false,"accounts-reason-required":false,"application-name":"testing","applicationname":"testing","asset-basedir":"./web/assets/","bind-address":"6.6.6.6","bindaddress":"6.6.6.6","db":{"address":":memory:","database":"postgres","password":"postgres","port":5432,"tlscacert":"","tlsmode":"disable","type":"sqlite","user":"postgres"},"db-address":":memory:","db-database":"postgres","db-password":"postgres","db-port":5432,"db-tls-ca-cert":"","db-tls-mode":"disable","db-type":"sqlite","db-user":"postgres","help":false,"host":"pooooopy","letsencrypt":{"certdir":"/gotosocial/storage/certs","emailaddress":"","enabled":false,"port":80},"letsencrypt-cert-dir":"/gotosocial/storage/certs","letsencrypt-email":"","letsencrypt-enabled":false,"letsencrypt-port":80,"log-level":"warn","loglevel":"warn","media":{"maxdescriptionchars":500,"maximagesize":2097152,"maxvideosize":10485760,"mindescriptionchars":0},"media-max-description-chars":500,"media-max-image-size":2097152,"media-max-video-size":10485760,"media-min-description-chars":0,"oidc":{"clientid":"","clientsecret":"","enabled":false,"idpname":"","issuer":"","scopes":["openid","email","profile","groups"],"skipverification":false},"oidc-client-id":"","oidc-client-secret":"","oidc-enabled":false,"oidc-idp-name":"","oidc-issuer":"","oidc-scopes":["openid","email","profile","groups"],"oidc-skip-verification":false,"port":6969,"protocol":"https","smtp":{"from":"me!!!","host":"some.email.host","port":99},"smtp-from":"me!!!","smtp-host":"some.email.host","smtp-port":99,"software-version":"","softwareversion":"","statuses":{"cwmaxchars":9999,"maxchars":0,"maxmediafiles":8,"pollmaxoptions":2,"polloptionmaxchars":67},"statuses-cw-max-chars":9999,"statuses-max-chars":0,"statuses-max-media-files":8,"statuses-poll-max-options":2,"statuses-poll-option-max-chars":67,"storage":{"backend":"s3","basepath":"/gotosocial/storage","servebasepath":"/fileserver","servehost":"localhost","serveprotocol":"https"},"storage-backend":"s3","storage-base-path":"/gotosocial/storage","storage-serve-base-path":"/fileserver","storage-serve-host":"localhost","storage-serve-protocol":"https","template":{"assetbasedir":"./web/assets/","basedir":"./web/template/"},"template-basedir":"./web/template/","trusted-proxies":["127.0.0.1/32","0.0.0.0/0"],"trustedproxies":["127.0.0.1/32","0.0.0.0/0"]}'
TEST_5="$(go run ./cmd/gotosocial/... debug config --config-path ./test/test5.yaml)"
if [ "${TEST_5}" != "${TEST_5_EXPECTED}" ]; then
    echo "TEST_5 not equal TEST_5_EXPECTED"
    exit 1
else
    echo "TEST_5 OK"
fi






echo "FINISHING CLI TESTS"
