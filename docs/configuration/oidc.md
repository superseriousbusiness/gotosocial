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

# Bool. Enable authentication with external OIDC provider. If set to true, then
# the other OIDC options must be set as well. If this is set to false, then the standard
# internal oauth flow will be used, where users sign in to GtS with username/password.
# Options: [true, false]
# Default: false
oidc-enabled: false

# String. Name of the oidc idp (identity provider). This will be shown to users when
# they log in.
# Examples: ["Google", "Dex", "Auth0"]
# Default: ""
oidc-idp-name: ""

# Bool. Skip the normal verification flow of tokens returned from the OIDC provider, ie.,
# don't check the expiry or signature. This should only be used in debugging or testing,
# never ever in a production environment as it's extremely unsafe!
# Options: [true, false]
# Default: false
oidc-skip-verification: false

# String. The OIDC issuer URI. This is where GtS will redirect users to for login.
# Typically this will look like a standard web URL.
# Examples: ["https://auth.example.org", "https://example.org/auth"]
# Default: ""
oidc-issuer: ""

# String. The ID for this client as registered with the OIDC provider.
# Examples: ["some-client-id", "fda3772a-ad35-41c9-9a59-f1943ad18f54"]
# Default: ""
oidc-client-id: ""

# String. The secret for this client as registered with the OIDC provider.
# Examples: ["super-secret-business", "79379cf5-8057-426d-bb83-af504d98a7b0"]
# Default: ""
oidc-client-secret: ""

# Array of string. Scopes to request from the OIDC provider. The returned values will be used to
# populate users created in GtS as a result of the authentication flow. 'openid' and 'email' are required.
# 'profile' is used to extract a username for the newly created user.
# 'groups' is optional and can be used to determine if a user is an admin (if they're in the group 'admin' or 'admins').
# Examples: See eg., https://auth0.com/docs/scopes/openid-connect-scopes
# Default: ["openid", "email", "profile", "groups"]
oidc-scopes:
  - "openid"
  - "email"
  - "profile"
  - "groups"

# Bool. Link OIDC authenticated users to existing ones based on their email address.
# This is mostly intended for migration purposes if you were running previous versions of GTS
# which only correlated users with their email address. Should be set to false for most usecases.
# Options: [true, false]
# Default: false
oidc-link-existing: false
```

## Behavior

When OIDC is enabled on GoToSocial, the default sign-in page redirects automatically to the sign-in page for the OIDC provider.

This means that OIDC essentially *replaces* the normal GtS email/password sign-in flow.

Due to the way the ActivityPub standard works, you _cannot_ change your username
after it has been set. This conflicts with the OIDC spec which does not
guarantee that the `preferred_username` field is stable.

To work with this, we ask the user to provide a username on their first login
attempt. The field for this is pre-filled with the value of the `preferred_username` claim.

After authenticating, GtS stores the `sub` claim supplied by the OIDC provider.
On subsequent authentication attempts, the user is looked up using this claim
exclusively.

This then allows you to change the username on a provider level without losing
access to your GtS account.

### Group membership

Most OIDC providers allow for the concept of groups and group memberships in returned claims. GoToSocial can use group membership to determine whether or not a user returned from an OIDC flow should be created as an admin account or not.

If the returned OIDC groups information for a user contains membership of the groups `admin` or `admins`, then that user will be created/signed in as though they are an admin.

## Migrating from old versions

If you're moving from an old version of GtS which used the unstable `email`
claim for unique user identification, you can set the `oidc-link-existing`
configuration to `true`. If no user can be found for the ID returned by the
provider, a lookup based on the `email` claim is performed instead. If this
succeeds, the stable id is added to the database for the matching user.

You should only use this for a limited time to avoid malicious account takeover.

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
