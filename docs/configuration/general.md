# General

The top-level configuration for GoToSocial, including basic things like host, port, bind address, and trusted-proxies.

## Settings

```yaml
###########################
##### GENERAL CONFIG ######
###########################

# String. Log level to use throughout the application. Must be lower-case.
# Options: ["trace","debug","info","warn","error","fatal"]
# Default: "info"
log-level: "info"

# Bool. Log database queries when log-level is set to debug or trace.
# This setting produces verbose logs, so it's better to only enable it
# when you're trying to track an issue down.
# Options: [true, false]
# Default: false
log-db-queries: false

# Bool. Include the client IP in the emitted log lines
# Options: [true, false]
# Default: true
log-client-ip: true

# String. Format to use for the timestamp in log lines.
# If set to the empty string, the timestamp will be
# ommitted from the logs entirely.
#
# The format must be compatible with Go's time.Layout, as
# documented on https://pkg.go.dev/time#pkg-constants.
#
# Examples: ["2006-01-02T15:04:05.000Z07:00", ""]
# Default: "02/01/2006 15:04:05.000"
log-timestamp-format: "02/01/2006 15:04:05.000"

# String. Application name to use internally.
# Examples: ["My Application","gotosocial"]
# Default: "gotosocial"
application-name: "gotosocial"

# String. The user that will be shown instead of the landing page. if no user is set, the landing page will be shown.
# Examples: "admin"
# Default: ""
landing-page-user: ""

# String. Hostname that this server will be reachable at. Defaults to localhost for local testing,
# but you should *definitely* change this when running for real, or your server won't work at all.
# DO NOT change this after your server has already run once, or you will break things!
# Examples: ["gts.example.org","some.server.com"]
# Default: "localhost"
host: "localhost"

# String. Domain to use when federating profiles. This is useful when you want your server to be at
# eg., "gts.example.org", but you want the domain on accounts to be "example.org" because it looks better
# or is just shorter/easier to remember.
#
# To make this setting work properly, you need to redirect requests at "example.org/.well-known/webfinger"
# to "gts.example.org/.well-known/webfinger" so that GtS can handle them properly.
#
# You should also redirect requests at "example.org/.well-known/nodeinfo" in the same way.
#
# You should also redirect requests at "example.org/.well-known/host-meta" in the same way. This endpoint
# is used by a number of clients to discover the API endpoint to use when the host and account domain are
# different.
#
# An empty string (ie., not set) means that the same value as 'host' will be used.
#
# DO NOT change this after your server has already run once, or you will break things!
#
# Please read the appropriate section of the installation guide before you go messing around with this setting:
# https://docs.gotosocial.org/en/latest/advanced/host-account-domain/
#
# Examples: ["example.org","server.com"]
# Default: ""
account-domain: ""

# String. Protocol over which the server is reachable from the outside world.
#
# ONLY CHANGE THIS TO HTTP FOR LOCAL TESTING! IN 99.99% OF CASES YOU SHOULD NOT CHANGE THIS!
#
# This should be the protocol part of the URI that your server is actually reachable on.
# So even if you're running GoToSocial behind a reverse proxy that handles SSL certificates
# for you, instead of using built-in letsencrypt, it should still be https, not http.
#
# Again, ONLY CHANGE THIS TO HTTP FOR LOCAL TESTING! If you set this to `http`, start your instance,
# and then later change it to `https`, you will have already broken URI generation for any created
# users on the instance. You should only touch this setting if you 100% know what you're doing.
#
# Options: ["http","https"]
# Default: "https"
protocol: "https"

# String. Address to bind the GoToSocial server to.
# This can be an IPv4 address or an IPv6 address (surrounded in square brackets), or a hostname.
# The default value will bind to all interfaces, which makes the server
# accessible by other machines. For most setups there is no need to change this.
# If you are using GoToSocial in a reverse proxy setup with the proxy running on
# the same machine, you will want to set this to "localhost" or an equivalent,
# so that the proxy can't be bypassed.
# Examples: ["0.0.0.0", "172.128.0.16", "localhost", "[::]", "[2001:db8::fed1]"]
# Default: "0.0.0.0"
bind-address: "0.0.0.0"

# Int. Listen port for the GoToSocial webserver + API. If you're running behind a reverse proxy and/or in a docker,
# container, just set this to whatever you like (or leave the default), and make sure it's forwarded properly.
# If you are running with built-in letsencrypt enabled, and running GoToSocial directly on a host machine, you will
# probably want to set this to 443 (standard https port), unless you have other services already using that port.
# This *MUST NOT* be the same as the letsencrypt port specified below, unless letsencrypt is turned off.
# Examples: [443, 6666, 8080]
# Default: 8080
port: 8080

# Array of string. CIDRs or IP addresses of proxies that should be trusted when determining real client IP from behind a reverse proxy.
# If you're running inside a Docker container behind Traefik or Nginx, for example, add the subnet of your docker network,
# or the gateway of the docker network, and/or the address of the reverse proxy (if it's not running on the host network).
# Example: ["127.0.0.1/32", "172.20.0.1"]
# Default: ["127.0.0.1/32", "::1"] (localhost ipv4 + ipv6)
trusted-proxies:
  - "127.0.0.1/32"
  - "::1"
```
