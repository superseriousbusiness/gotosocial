#!/bin/sh

set -eux

SERVER_URL="http://localhost:8080"
REDIRECT_URI="${SERVER_URL}"
CLIENT_NAME="Test Application Name"
REGISTRATION_REASON="Testing whether or not this dang diggity thing works!"
REGISTRATION_USERNAME="${1}"
REGISTRATION_EMAIL="${2}"
REGISTRATION_PASSWORD="very safe password 123"
REGISTRATION_AGREEMENT="true"
REGISTRATION_LOCALE="en"

# Step 1: create the app to register the new account
CREATE_APP_RESPONSE=$(curl --fail -s -X POST -F "client_name=${CLIENT_NAME}" -F "redirect_uris=${REDIRECT_URI}" "${SERVER_URL}/api/v1/apps")
CLIENT_ID=$(echo "${CREATE_APP_RESPONSE}" | jq -r .client_id)
CLIENT_SECRET=$(echo "${CREATE_APP_RESPONSE}" | jq -r .client_secret)
echo "Obtained client_id: ${CLIENT_ID} and client_secret: ${CLIENT_SECRET}"

# Step 2: obtain a code for that app
APP_CODE_RESPONSE=$(curl --fail -s -X POST -F "scope=read" -F "grant_type=client_credentials" -F "client_id=${CLIENT_ID}" -F "client_secret=${CLIENT_SECRET}" -F "redirect_uri=${REDIRECT_URI}" "${SERVER_URL}/oauth/token")
APP_ACCESS_TOKEN=$(echo "${APP_CODE_RESPONSE}" | jq -r .access_token)
echo "Obtained app access token: ${APP_ACCESS_TOKEN}"

# Step 3: use the code to register a new account
ACCOUNT_REGISTER_RESPONSE=$(curl --fail -s -H "Authorization: Bearer ${APP_ACCESS_TOKEN}" -F "reason=${REGISTRATION_REASON}" -F "email=${REGISTRATION_EMAIL}" -F "username=${REGISTRATION_USERNAME}" -F "password=${REGISTRATION_PASSWORD}" -F "agreement=${REGISTRATION_AGREEMENT}" -F "locale=${REGISTRATION_LOCALE}" "${SERVER_URL}/api/v1/accounts")
USER_ACCESS_TOKEN=$(echo "${ACCOUNT_REGISTER_RESPONSE}" | jq -r .access_token)
echo "Obtained user access token: ${USER_ACCESS_TOKEN}"

# # Step 4: verify the returned access token
curl -s -H "Authorization: Bearer ${USER_ACCESS_TOKEN}" "${SERVER_URL}/api/v1/accounts/verify_credentials" | jq
