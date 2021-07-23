# OpenID Connect (OIDC)

GoToSocial supports [OpenID Connect](https://openid.net/connect/), which is an identification protocol built on top of [OAuth 2.0](https://oauth.net/2/), an industry standard protocol for authorization.

This means that you can connect GoToSocial to an external OIDC provider like [Gitlab](https://docs.gitlab.com/ee/integration/openid_connect_provider.html), [Google](https://cloud.google.com/identity-platform/docs/web/oidc), [Keycloak](https://www.keycloak.org/), or [Dex](https://dexidp.io/) and allow users to sign in to GoToSocial using their credentials for that provider.

This is very convenient in the following cases:

- You're running multiple services on a platform (Matrix, Peertube, GoToSocial), and you want users to be able to use the same sign in page for all of them.
- You want to delegate management of users, accounts, passwords etc. to an external service to make admin easier for yourself.
- You already have a lot of users in an external system and you don't want to have to recreate them all in GoToSocial manually.

## Settings

GoToSocial exposes the following configuration settings for OIDC, shown below with their default values.

```yaml
#######################
##### OIDC CONFIG #####
#######################

# Config for authentication with an external OIDC provider (Dex, Google, Auth0, etc).
oidc:

  # Bool. Enable authentication with external OIDC provider. If set to true, then
  # the other OIDC options must be set as well. If this is set to false, then the standard
  # internal oauth flow will be used, where users sign in to GtS with username/password.
  # Options: [true, false]
  # Default: false
  enabled: false

  # String. Name of the oidc idp (identity provider). This will be shown to users when
  # they log in.
  # Examples: ["Google", "Dex", "Auth0"]
  # Default: ""
  idpName: ""

  # Bool. Skip the normal verification flow of tokens returned from the OIDC provider, ie.,
  # don't check the expiry or signature. This should only be used in debugging or testing,
  # never ever in a production environment as it's extremely unsafe!
  # Options: [true, false]
  # Default: false
  skipVerification: false

  # String. The OIDC issuer URI. This is where GtS will redirect users to for login.
  # Typically this will look like a standard web URL.
  # Examples: ["https://auth.example.org", "https://example.org/auth"]
  # Default: ""
  issuer: ""

  # String. The ID for this client as registered with the OIDC provider.
  # Examples: ["some-client-id", "fda3772a-ad35-41c9-9a59-f1943ad18f54"]
  # Default: ""
  clientID: ""

  # String. The secret for this client as registered with the OIDC provider.
  # Examples: ["super-secret-business", "79379cf5-8057-426d-bb83-af504d98a7b0"]
  # Default: ""
  clientSecret: ""

  # Array of string. Scopes to request from the OIDC provider. The returned values will be used to
  # populate users created in GtS as a result of the authentication flow. 'openid' and 'email' are required.
  # 'profile' is used to extract a username for the newly created user.
  # 'groups' is optional and can be used to determine if a user is an admin (if they're in the group 'admin' or 'admins').
  # Examples: See eg., https://auth0.com/docs/scopes/openid-connect-scopes
  # Default: ["openid", "email", "profile", "groups"]
  scopes:
    - "openid"
    - "email"
    - "profile"
    - "groups"
```

## Behavior

When OIDC is enabled on GoToSocial, the default sign-in page redirects automatically to the sign-in page for the OIDC provider.

This means that OIDC essentially *replaces* the normal GtS email/password sign-in flow.

When a user logs in through OIDC, GoToSocial will request that user's preferred email address and username from the OIDC provider. It will then use the returned email address to either:

*If the email address is already associated with a user/account*: sign the requester in as that user/account.

Or:

*If the email address is not yet associated with a user/account*: create a new user and account with the returned credentials, and sign the requester in as that user/account.

In other words, GoToSocial completely delegates sign-in authority to the OIDC provider, and trusts whatever credentials it returns.

### Username conflicts

In some cases, such as when a server has been switched to use OIDC after already using default settings for a while, there may be an overlap between usernames returned from OIDC, and usernames that already existed in the database.

For example, let's say that someone with username `gordonbrownfan` and email address `gordon_is_best@example.org` has an account on a GtS instance that uses the default sign-in flow.

That GtS instance then switches to using OIDC login. However, in the OIDC's storage there's also a user with username `gordonbrownfan`. If this user has the email address `gordon_is_best@example.org`, then GoToSocial will assume that the two users are the same and just log `gordonbrownfan` in as though nothing had changed. No problem!

However, if the user in the OIDC storage has a different email address, GoToSocial will try to create a new user and account for this person.

Since the username `gordonbrownfan` is already taken, GoToSocial will try `gordonbrownfan1`. If this is also taken, it will try `gordonbrownfan2`, and so on, until it finds a username that's not yet taken. It will then sign the requester in as that user/account, distinct from the original `gordonbrownfan`.

### Malformed usernames

A username returned from an OIDC provider might not always fit the pattern of what GoToSocial accepts as a valid username, ie., lower-case letters, numbers, and underscores. In this case, GoToSocial will do its best to parse the returned username into something that fits the pattern.

For example, say that an OIDC provider returns the username `Marx Is Great` for a sign in, which doesn't fit the pattern because it contains upper-case letters and spaces.

In this case, GtS will convert it into `marx_is_great` by applying the following rules:

1. Trim any leading or trailing whitespace.
2. Convert all letters to lowercase.
3. Replace spaces with underscores.

Unfortunately, at this point GoToSocial doesn't know how to handle returned usernames containing special characters such as `@` or `%`, so these will return an error.

### Group membership

Most OIDC providers allow for the concept of groups and group memberships in returned claims. GoToSocial can use group membership to determine whether or not a user returned from an OIDC flow should be created as an admin account or not.

If the returned OIDC groups information for a user contains membership of the groups `admin` or `admins`, then that user will be created/signed in as though they are an admin.

## Provider Examples

### Dex

[Dex](https://dexidp.io/) is an open-source OIDC Provider that you can host yourself. The procedure for installing Dex is out of scope for this documentation, but you can check out the Dex docs [here](https://dexidp.io/docs/).

Dex is great because it's also written in Go, like GoToSocial, which means it's small and fast and works well on lower-powered systems.

To configure Dex and GoToSocial to work together, create the following client under the `staticClients` section of your Dex config:

```yaml
staticClients:
  - id: 'gotosocial_client'
    redirectURIs:
      - 'https://gotosocial.example.org/auth/callback'
    name: 'GoToSocial'
    secret: 'some-client-secret'
```

Make sure to replace `gotosocial_client` with your desired client ID, and `secret` with a reasonably long and secure secret (a UUID for example). You should also replace `gotosocial.example.org` with the `host` of your GtS instance, but leave `/auth/callback` in place.

Now, edit the `oidc` section of your GoToSocial config.yaml as follows:

```yaml
oidc:
  enabled: true
  idpName: "Dex"
  skipVerification: false
  issuer: "https://auth.example.org"
  clientID: "gotosocial_client"
  clientSecret: "some-client-secret"
  scopes:
    - "openid"
    - "email"
    - "profile"
    - "groups"
```

Make sure to replace the `issuer` variable with whatever your Dex issuer is set to. This should be the exact URI at which your Dex instance can be reached.

Now, restart both GoToSocial and Dex so that the new settings are in place.

When you next go to log in to GtS, you should be redirected to the sign in page for Dex. On a successful sign in, you'll be directed back to GoToSocial.
