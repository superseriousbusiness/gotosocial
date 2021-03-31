#!/bin/sh

# Step 1: create the app to register the new account
curl -X POST -F "client_name=ahhhhhh" -F "redirect_uris=http://localhost:8080" localhost:8080/api/v1/apps

# Step 2: obtain a code for that app
curl -X POST -F "scope=read" -F "grant_type=client_credentials" -F "client_id=bbec8b67-b389-49fb-ad9c-4a990e95d75a" -F "client_secret=da21d8b1-0705-4a1c-a38e-96060ab5553d" -F "redirect_uri=http://localhost:8080" localhost:8080/oauth/token

# Step 3: use the code to register a new account
curl -H "Authorization: Bearer MGVHMZQYYMYTNJK4OC0ZN2I3LTGWNWETMGE3ZTY2NTJKYZE4" -F "reason=seems like a good time my dude" -F "email=user7@example.org" -F "username=test_user7" -F "password=this is a big long password" -F "agreement=true" -F "locale=en" localhost:8080/api/v1/accounts

# Step 4: verify the returned access token 
curl -H "Authorization: Bearer ODGYODAWNTUTZJUZZI0ZMJK3LWEXMJYTYZA5NMVKZDYWMGQY" localhost:8080/api/v1/accounts/verify_credentials
