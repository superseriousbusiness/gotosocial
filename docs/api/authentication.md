# Authentication with the API

Using the client API requires authentication. This page documents the general flow for retrieving an authentication token with examples for doing this on the CLI

## Create a new application

We need to register a new application, which we can then use to request an OAuth token. This is done by making a `POST` request to the `/api/v1/apps` endpoint. Replace `your_app_name` in the code below:

```bash
curl -X POST 'https://your.instance.url/api/v1/apps' \ 
  -H 'Content-Type:application/json' \
  -d '{
      "client_name": "your_app_name",
      "redirect_uris": "urn:ietf:wg:oauth:2.0:oob",
      "scopes": "read"
    }'
```

The string "urn:ietf:wg:oauth:2.0:oob" is an indication of what is known as out-of-band authentication - a technique used in multi-factor authentication to reduce the number of ways that a bad actor can intrude on the authentication process. In this instance, it allows us to view and manually copy the tokens created to use further in this process.

Note that `scopes` can be any space-separated combination of:

- read
- write
- admin

!!! warning
    GoToSocial does not currently support scoped authorization tokens, so any token you obtain in this process will be able to perform all actions on your behalf, including admin actions if your account has admin permissions. Nevertheless, it is always good practice to grant your application the lowest tier permissions it needs to do its job. e.g. If your application won't be making posts, use scope=read.
   
    In this spirit, "read" is used in the example above, which means that in the future when scoped tokens are supported, the application will be restricted to only being able to do "read" actions.
   
    You can read more about additional planned OAuth security features [right here](https://github.com/superseriousbusiness/gotosocial/issues/2232).

A successful call returns a response with a `client_id` and `client_secret` that we are going need to use in the rest of the process. It looks something like this: 
```json
{
  "id": "randomised_id",
  "name": "your_app_name",
  "redirect_uri": "urn:ietf:wg:oauth:2.0:oob",
  "client_id": "your_new_client_id",
  "client_secret": "your_new_client_secret"
}
```

!!! tip
    Ensure you save the `client_id` and `client_secret` somewhere so you can refer to it as we go.

## Authorizing your application

We've registered a new application with GoToSocial, but it isn't connected to your account just yet. Now we need to tell GoToSocial that that new application is actually going to act on your behalf. To do this, we need to authenticate with your instance via a browser to initiate the login and permission granting process.

Using the `client_id` from the previous step, create a URL with a query string like so:
```
https://your.instance.url/oauth/authorize?client_id=your_new_client_id-id&redirect_uri=urn:ietf:wg:oauth:2.0:oob&response_type=code
```

You'll get a login form to your instance and be prompted to login if you aren't already logged in. Once logged in, you will get a screen that says something like this:
```
Hi `your_username`!

Application `your_app_name` would like to perform actions on your behalf, with scope *`read`*.
The application will redirect to urn:ietf:wg:oauth:2.0:oob to continue.
```

Once you click `Allow`, you will get a window that looks something like this:

```
Here's your out-of-band token with scope "read", use it wisely:
WBANQAXXHDN9KJQZXGWNQANA4V9EJWMUDHVCUUM3JHYAB3DP
```


## Getting your token
The next step is to send a `POST` request to the oauth/token endpoint to get an access token that you will use to authenticate your future requests. 
```bash
curl -X POST 'https://your.instance.url/oauth/token' \
  -H 'Content-Type:application/json' 
  -d '{
      "redirect_uri": "urn:ietf:wg:oauth:2.0:oob",
      "client_id": "your_new_client_id",
      "client_secret": "your_new_client_secret",
      "grant_type": "authorization_code",
      "code": "WBANQAXXHDN9KJQZXGWNQANA4V9EJWMUDHVCUUM3JHYAB3DP"
    }' 
```
!!! warning
    Please note that the characters used above are just a random selection of characters and cannot be used.
    Make sure you replace it with the code *you* get from your instance.

You'll get a response that includes your access token and looks something like this:
```json
{
  "access_token": "your_brand_new_access_token",
  "created_at": unixtimestamp,
  "scope": "read",
  "token_type": "Bearer"
}
```
## Verifying it all
To make sure everything went through successfully, query the `/api/v1/verify_credentials` endpoint, adding your new access token to the Header as `Authorization: Bearer your_brand_new_token`.

See this example:
```bash
curl -X GET 'https://your.instance.url/api/v1/accounts/verify_credentials' \
  -H 'Content-Type:application/json' 
  -H 'Authorization:Bearer your_brand_new_token'
```
If all goes well, you should get your user profile as a response response.

## Final notes
Thereafter, whenever you make an api call, say to query your notifications, add a Header to your request `Authorization:Bearer your_brand_new_token` like this:
```bash
curl --request GET \
  --url https://your.instance.url/api/v1/notifications \
  --header 'Authorization: Bearer your_brand_new_token' \
```

