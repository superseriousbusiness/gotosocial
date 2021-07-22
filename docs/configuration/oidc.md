# OpenID Connect (OIDC)

GoToSocial supports [OpenID Connect](https://openid.net/connect/), which is an identification protocol built on top of [OAuth 2.0](https://oauth.net/2/), an industry standard protocol for authorization.

This means that you can connect GoToSocial to an external OIDC provider like [Gitlab](https://docs.gitlab.com/ee/integration/openid_connect_provider.html), [Google](https://cloud.google.com/identity-platform/docs/web/oidc), [Keycloak](https://www.keycloak.org/), or [Dex](https://dexidp.io/) and allow users to sign in to GoToSocial using their credentials for that provider.

This is very convenient in the following cases:

- You're running multiple services on a platform, and you want users to be able to use the same sign in page for all of them.
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

## Provider Examples

### Dex

[Dex](https://dexidp.io/) is an open-source OIDC Provider that you can host yourself.

To configure Dex and GoToSocial to work together, create the following client under the `staticClients` section of the Dex config:

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

When you next go to log in, you should be redirected to Dex.
