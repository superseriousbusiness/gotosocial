# Instance

## Settings

```yaml
###########################
##### INSTANCE CONFIG #####
###########################

# Config pertaining to instance federation settings, pages to hide/expose, etc.

# Bool. Allow unauthenticated users to make queries to /api/v1/instance/peers?filter=open in order
# to see a list of instances that this instance 'peers' with. Even if set to 'false', then authenticated
# users (members of the instance) will still be able to query the endpoint.
# Options: [true, false]
# Default: false
instance-expose-peers: false

# Bool. Allow unauthenticated users to make queries to /api/v1/instance/peers?filter=suspended in order
# to see a list of instances that this instance blocks/suspends. This will also allow unauthenticated
# users to see the list through the web UI. Even if set to 'false', then authenticated users (members
# of the instance) will still be able to query the endpoint.
# Options: [true, false]
# Default: false
instance-expose-suspended: false

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
```
