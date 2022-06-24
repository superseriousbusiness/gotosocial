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
```
