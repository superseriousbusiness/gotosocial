# Instance

## Settings

```yaml
###########################
##### INSTANCE CONFIG #####
###########################

# Config pertaining to instance federation settings, pages to hide/expose, etc.

# String. Federation mode to use for this instance.
#
# "blocklist" -- open federation by default. Only instances that are explicitly 
#                blocked will be denied (unless they are also explicitly allowed).
#
# "allowlist" -- closed federation by default. Only instances that are explicitly
#                allowed will be able to interact with this instance.
#
# For more details on blocklist and allowlist modes, check the documentation at:
# https://docs.gotosocial.org/en/latest/admin/federation_modes
#
# Options: ["blocklist", "allowlist"]
# Default: "blocklist"
instance-federation-mode: "blocklist"

# Bool. Allow unauthenticated users to make queries to /api/v1/instance/peers?filter=open in order
# to see a list of instances that this instance 'peers' with. Even if set to 'false', then authenticated
# users (members of the instance) will still be able to query the endpoint.
# Options: [true, false]
# Default: false
instance-expose-peers: false

# Bool. Allow unauthenticated users to make queries to /api/v1/instance/peers?filter=suspended in order
# to see a list of instances that this instance blocks/suspends. Even if set to 'false', then authenticated
# users (members of the instance) will still be able to query the endpoint.
#
# WARNING: Setting this variable to 'true' may result in your instance being scraped by blocklist scrapers.
# See: https://docs.gotosocial.org/en/latest/admin/domain_blocks/#block-announce-bots
#
# Options: [true, false]
# Default: false
instance-expose-suspended: false

# Bool. Allow unauthenticated users to view /about/suspended,
# showing the HTML rendered list of instances that this instance blocks/suspends.
# Options: [true, false]
# Default: false
instance-expose-suspended-web: false

# Bool. Allow unauthenticated users to make queries to /api/v1/timelines/public in order
# to see a list of public posts on this server. Even if set to 'false', then authenticated
# users (members of the instance) will still be able to query the endpoint.
# Options: [true, false]
# Default: false
instance-expose-public-timeline: false

# Bool. This flag tweaks whether GoToSocial will deliver ActivityPub messages
# to the shared inbox of a recipient, if one is available, instead of delivering
# each message to each actor who should receive a message individually.
#
# Shared inbox delivery can significantly reduce network load when delivering
# to multiple recipients share an inbox (eg., on large Mastodon instances).
#
# See: https://www.w3.org/TR/activitypub/#shared-inbox-delivery
#
# Options: [true, false]
# Default: true
instance-deliver-to-shared-inboxes: true

# Bool. This flag will inject a Mastodon version into the version field that
# is included in /api/v1/instance. This version is often used by Mastodon clients
# to do API feature detection. By injecting a Mastodon compatible version, it is
# possible to cajole those clients to behave correctly with GoToSocial.
#
# Options: [true, false]
# Default: false
instance-inject-mastodon-version: false
```
