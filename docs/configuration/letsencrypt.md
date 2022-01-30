# LetsEncrypt

## Settings

```yaml
##############################
##### LETSENCRYPT CONFIG #####
##############################

# Config pertaining to the automatic acquisition and use of LetsEncrypt HTTPS certificates.

# Bool. Whether or not letsencrypt should be enabled for the server.
# If false, the rest of the settings here will be ignored.
# If you serve GoToSocial behind a reverse proxy like nginx or traefik, leave this turned off.
# If you don't, then turn it on so that you can use https.
# Options: [true, false]
# Default: false
letsencrypt-enabled: false

# Int. Port to listen for letsencrypt certificate challenges on.
# If letsencrypt is enabled, this port must be reachable or you won't be able to obtain certs.
# If letsencrypt is disabled, this port will not be used.
# This *must not* be the same as the webserver/API port specified above.
# Examples: [80, 8000, 1312]
# Default: 80
letsencrypt-port: 80

# String. Directory in which to store LetsEncrypt certificates.
# It is a good move to make this a sub-path within your storage directory, as it makes
# backup easier, but you might wish to move them elsewhere if they're also accessed by other services.
# In any case, make sure GoToSocial has permissions to write to / read from this directory.
# Examples: ["/home/gotosocial/storage/certs", "/acmecerts"]
# Default: "/gotosocial/storage/certs"
letsencrypt-cert-dir: "/gotosocial/storage/certs"

# String. Email address to use when registering LetsEncrypt certs.
# Most likely, this will be the email address of the instance administrator.
# LetsEncrypt will send notifications about expiring certificates etc to this address.
# Examples: ["admin@example.org"]
# Default: ""
letsencrypt-email-address: ""
```
