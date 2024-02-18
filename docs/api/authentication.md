# Authentication with the API

GoToSocial implements an API that is compatible with the Mastodon API which means that most everything to do with the API follows many of the conventions and techniques that work for the Mastodon API. That includes authentication.

Find below the steps used to authenticate with the GoToSocial API.

---

## Create a new application

We need to get a new key:pair for our application. This is done by making a `POST` request to the `/api/v1/apps` endpoint. Replace `your_app_name` in the code below:

```
curl -X POST 'https://your.instance.url/api/v1/apps' -H 'Content-Type:application/json' -d '{
  "client_name": your_app_name,
  "redirect_uris": "urn:ietf:wg:oauth:2.0:oob",
  "scopes": "read"
}'

```

Note that `scopes` can be one of:
- read
- write
- follow
- push
- admin

!!! warning
    Be **very** careful what scopes you grant to your application.
    The default is `read` if `scopes` is not set.

A successful call returns a response with a client_id and client_secret that we are going need to use in the rest of the process. It looks something like this: 
```
{
  "id": randomised_id,
  "name": your_app_name,
  "redirect_uri": "urn:ietf:wg:oauth:2.0:oob",
  "client_id": your_new_client_id,
  "client_secret": your_new_client_secret
}
```

!!! important
    Ensure you save the client_id and client_secret somewhere so you can refer to it as we go.

## authorizing your application

We've registered a new application with GoToSocial, but it isn't connected to your account just yet. Now we need to tell GoToSocial that that new application is actually going to interact as you. To do this, we need to authenticate with your instance via a browser to process the login and permission granting process.

Using the `client_id` from the previous step, create a URL with a query string like so:
```
https://your.instance.url/oauth/authorize?client_id=your_new_client_id-id&redirect_uri=urn:ietf:wg:oauth:2.0:oob&response_type=code
```

You'll get a login form to your instance and be prompted to login if you aren't already logged in. Once logged in, you will get a screen that says something like this:
!!! note
    Hi `your_username`!
 
    Application `your_app_name` would like to perform actions on your behalf, with scope *`read`*.
    The application will redirect to urn:ietf:wg:oauth:2.0:oob to continue.
 
    <span style="padding:12px;background-color:#66befe;color:black;">Allow</span>
 

Once you click `Allow`, you will get a window that looks something like this:

!!! note
    Here's your out-of-band token with scope "read", use it wisely:
     WBANQAXXHDN9KJQZXGWNQANA4V9EJWMUDHVCUUM3JHYAB3DP


## Getting your token
The next step is to send a `POST` request to the oauth/token endpoint to get an access token that you will use to authenticate your future requests. 
```
curl -X POST 'https://your.instance.url/oauth/token' -H 'Content-Type:application/json' -d '{
  "redirect_uri": "urn:ietf:wg:oauth:2.0:oob",
  "client_id": "your_new_client_id",
  "client_secret": "your_new_client_secret",
  "grant_type": "authorization_code",
  "code": "WBANQAXXHDN9KJQZXGWNQANA4V9EJWMUDHVCUUM3JHYAB3DP"
}' 
```
!!! warning
    Please note that the characters used above are just a random selection of characters and cannot be used.
    ake sure you replace it with the code *you* get from your instance.

You'll get a response that includes your access token and looks something like this:
```
{
  "access_token": "your_brand_new_access_token",
  "created_at": unixtimestamp,
  "scope": "read",
  "token_type": "Bearer"
}
```
## Verifying it all
To make sure everything went through successfully, target the `/api/v1/verify_credentials` endpoint, adding your new access token to the Header as `Authentication:Bearer your_brand_new_token`.

See this example:
```
curl -X GET 'https://your.instance.url/api/v1/accounts/verify_credentials' -H 'Content-Type:application/json' -H 'Authorization:Bearer your_brand_new_token'
```
If all goes well, you should get your user profile as a response response.

## Final notes
Thereafter, whenever you make an api call, say to query your notifications, add a Header to your request `Authorization:Bearer your_brand_new_token` like this:
```
curl --request GET \
  --url https://your.instance.url/api/v1/notifications \
  --header 'Authorization: Bearer your_brand_new_token' \
```

