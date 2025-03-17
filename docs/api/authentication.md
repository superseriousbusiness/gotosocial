# Authentication with the API

Using the client API requires authentication. This page documents the general flow for retrieving an authentication token with examples for doing this on the CLI using `curl`.

!!! tip
    If you want to get an API access token via the settings panel instead, without having to use the command line, see the [Application documentation](https://docs.gotosocial.org/en/latest/user_guide/settings/#applications).

## Create a new application

We need to register a new application, which we can then use to request an OAuth token. This is done by making a `POST` request to the `/api/v1/apps` endpoint. Replace `your_app_name` in the command below with the name you want to use for your application:

```bash
curl \
  -H 'Content-Type:application/json' \
  -d '{
        "client_name": "your_app_name",
        "redirect_uris": "urn:ietf:wg:oauth:2.0:oob",
        "scopes": "read"
      }' \
  'https://example.org/api/v1/apps'
```

The string `urn:ietf:wg:oauth:2.0:oob` is an indication of what is known as out-of-band authentication - a technique used in multi-factor authentication to reduce the number of ways that a bad actor can intrude on the authentication process. In this instance, it allows us to view and manually copy the tokens created to use further in this process.

!!! tip "Scopes"
    It is always good practice to grant your application the lowest tier permissions it needs to do its job. e.g. If your application won't be making posts, use `scope=read` or even a subscope of that.
    
    In this spirit, "read" is used in the example above, which means that the application will be restricted to only being able to do "read" actions.
    
    For a list of available scopes, see [the swagger docs](https://docs.gotosocial.org/en/latest/api/swagger/).

!!! warning
    GoToSocial did not support scoped authorization tokens before version 0.19.0, so if you are using a version of GoToSocial below that, then any token you obtain in this process will be able to perform all actions on your behalf, including admin actions if your account has admin permissions.

A successful call returns a response with a `client_id` and `client_secret`, which we are going need to use in the rest of the process. It looks something like this:

```json
{
  "id": "01J1CYJ4QRNFZD6WHQMZV7248G",
  "name": "your_app_name",
  "redirect_uri": "urn:ietf:wg:oauth:2.0:oob",
  "client_id": "YOUR_CLIENT_ID",
  "client_secret": "YOUR_CLIENT_SECRET"
}
```

!!! tip
    Ensure you save the `client_id` and `client_secret` values somewhere so you can refer to them as we go.

## Authorize your application to act on your behalf

We've registered a new application with GoToSocial, but it isn't connected to your account just yet. Now we need to tell GoToSocial that that new application is actually going to act on your behalf. To do this, we need to authenticate with your instance via a browser to initiate the login and permission-granting process.

Create a URL with a query string like so, replacing `YOUR_CLIENT_ID` with the `client_id` you received in the previous step, and paste the URL into your browser:

```text
https://example.org/oauth/authorize?client_id=YOUR_CLIENT_ID&redirect_uri=urn:ietf:wg:oauth:2.0:oob&response_type=code&scope=read
```

!!! tip
    If you used different scopes to register your application, then replace `scope=read` in the URL above with a plus-separated list of the scopes you registered with. For example, if you registered your application with a `scopes` value of `read write` then you should change `scope=read` in the above URL to `scope=read+write`. 

After pasting the URL into your browser, you'll be directed to a login form for your instance which prompts you to enter your email address and password in order to connect the application to your account.

Once you've submitted your credentials, you will arrive on a page that says something like this:

```
Hi `your_username`!

Application `your_app_name` would like to perform actions on your behalf, with scope *`read`*.
The application will redirect to urn:ietf:wg:oauth:2.0:oob to continue.
```

Click `Allow`, and you will get a page that looks something like this:

```text
Here's your out-of-band token with scope "read", use it wisely:
YOUR_AUTHORIZATION_TOKEN
```

Copy the out-of-band authorization token somewhere, as you'll need it in the next step.

## Get an access token

The next step is to exchange the out-of-band authorization token you just received with a reusable access token that can be sent along with all further API requests.

You can do this with another `POST` request that looks like the following:

```bash
curl \
  -H 'Content-Type: application/json' \
  -d '{
        "redirect_uri": "urn:ietf:wg:oauth:2.0:oob",
        "client_id": "YOUR_CLIENT_ID",
        "client_secret": "YOUR_CLIENT_SECRET",
        "grant_type": "authorization_code",
        "code": "YOUR_AUTHORIZATION_TOKEN"
      }' \
  'https://example.org/oauth/token'
```

Make sure to replace:

- `YOUR_CLIENT_ID` with the client ID received in the first step.
- `YOUR_CLIENT_SECRET` with the client secret received in the first step.
- `YOUR_AUTHORIZATION_TOKEN` with the out-of-band authorization token received in the second step.

You'll get a response that includes your access token and looks something like this:

```json
{
  "access_token": "YOUR_ACCESS_TOKEN",
  "created_at": 1719577950,
  "scope": "read",
  "token_type": "Bearer"
}
```

Copy and save your access token somewhere safe.

## Verifying

To make sure everything worked, try querying the `/api/v1/verify_credentials` endpoint, adding your access token to the request header as `Authorization: Bearer YOUR_ACCESS_TOKEN`.

See this example:

```bash
curl \
  -H 'Authorization: Bearer YOUR_ACCESS_TOKEN' \
  'https://example.org/api/v1/accounts/verify_credentials'
```
If all goes well, you should get your user profile as a JSON response.

## Final notes

Now that you have an access token, you can reuse that token in every API request for authorization. You do not need to do the entire token exchange dance every time!

For example, you can issue another `GET` request to the API using the same access token to get your notifications, as follows:

```bash
curl \
  -H 'Authorization: Bearer YOUR_ACCESS_TOKEN' \
  'https://example.org/api/v1/notifications'
```
