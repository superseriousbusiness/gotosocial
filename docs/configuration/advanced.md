# Advanced

Advanced settings options are provided for the sake of allowing admins to tune their instance to their liking.

These are set to sensible defaults, so most server admins won't need to touch them or think about them.

**Changing these settings if you don't know what you're doing may break your instance**.

## Settings

```yaml
#############################
##### ADVANCED SETTINGS #####
#############################

# Advanced settings pertaining to http timeouts, security, cookies, and more.
#
# ONLY ADJUST THESE SETTINGS IF YOU KNOW WHAT YOU ARE DOING!
#
# Most users will not need to (and should not) touch these settings, since
# they are set to sensible defaults, and may break if they are changed.
#
# Nevertheless, they are provided for the sake of allowing server admins to
# tweak their instance for performance or security reasons.

# String. Value of the SameSite attribute of cookies set by GoToSocial.
# Defaults to 'lax' to ensure that the OIDC flow does not break, which is
# fine in most cases. If you want to harden your instance against CSRF attacks
# and don't mind if some login-related things might break, you can set this
# to 'strict' instead.
#
# For an overview of what this does, see:
# https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite
#
# Options: ["lax", "strict"]
# Default: "lax"
advanced-cookies-samesite: "lax"

# Int. Amount of requests to permit per router grouping from a single IP address within
# a span of 5 minutes. If this amount is exceeded, a 429 HTTP error code will be returned.
#
# If you find yourself adjusting this limit because it's regularly being exceeded,
# you should first verify that your settings for `trusted-proxies` (above) are correct.
# In many cases, when the rate limit is exceeded it is because your instance sees all
# incoming requests as coming from the *same IP address* (you can verify this by looking
# at the client IPs in your instance logs). If this is the case, try adding that IP
# address to your `trusted-proxies` *BEFORE* you go adjusting this rate limit setting!
#
# If you set this to 0 or less, rate limiting will be disabled entirely.
#
# Examples: [1000, 500, 0]
# Default: 300
advanced-rate-limit-requests: 300

# Array of string. CIDRs to except from rate limit restrictions.
# Any IPs inside the CIDR range(s) will not have rate limiting
# applied on their requests, and rate limit headers will not be
# set for those requests.
#
# For IPv6, we only take subnets up to a /64 into account. If you
# want to open up a larger prefix, you'll need to list multiple
# prefixes instead.
#
# This can be useful in the following example cases (and probably
# a bunch of others as well):
#
# 1. You've set up an automated service that uses the API, and
#    it keeps getting rate limited, even though you trust it's
#    not abusing the instance.
#
# 2. You live with multiple people who use the same instance,
#    and you're all using the same router/NAT, so you all have
#    the same IP address, and you keep rate limiting each other.
#
# 3. You mostly use your own home internet to access your instance,
#    and you want to exempt your home internet from rate limiting.
#
# You should be careful when adjusting this setting, since you
# might inadvertently make rate limiting useless if you set too
# wide a range. If in doubt, be too restrictive rather than too
# lenient, and adjust as you go.
#
# Example: ["192.168.0.0/16", "2001:DB8:FACE:CAFE::/64"]
# Default: []
advanced-rate-limit-exceptions: []

# Int. Amount of open requests to permit per CPU, per router grouping, before applying http
# request throttling. Any requests beyond the calculated limit are held in a backlog queue for
# up to 30 seconds before either being processed or timing out. Requests that don't fit in the backlog
# queue will have status 503 returned to them, and the header 'Retry-After' will be set to 30 seconds.
#
# Open request limit is available CPUs * multiplier; backlog queue limit is limit * multiplier.
#
# Example values for multiplier 8:
#
# 1 cpu = 08 open, 064 backlog
# 2 cpu = 16 open, 128 backlog
# 4 cpu = 32 open, 256 backlog
#
# Example values for multiplier 4:
#
# 1 cpu = 04 open, 016 backlog
# 2 cpu = 08 open, 032 backlog
# 4 cpu = 16 open, 064 backlog
#
# A multiplier of 8 is a sensible default, but you may wish to increase this for instances
# running on very performant hardware, or decrease it for instances using v. slow CPUs.
#
# If you set this to 0 or less, http request throttling will be disabled entirely.
#
# Examples: [8, 4, 9, 0]
# Default: 8
advanced-throttling-multiplier: 8

# Duration. Time period to use as the "retry-after" header value in response to throttled requests.
# Minimum resolution is 1 second.
#
# Examples: [30s, 10s, 5s, 1m]
# Default: "30s"
advanced-throttling-retry-after: "30s"

# Int. CPU multiplier for the fixed number of goroutines to spawn in order to send messages via ActivityPub.
# Messages will be batched and pushed to a singular queue, from which multiplier * CPU count goroutines will
# pull and attempt deliveries. This can be tuned to limit concurrent posting to remote inboxes, preventing
# your instance CPU usage skyrocketing when accounts with many followers post statuses.
#
# If you set this to 0 or less, only 1 sender will be used regardless of CPU count. This may be
# useful in cases where you are working with very tight network or CPU constraints.
#
# Example values for multiplier 2 (default):
#
# 1 cpu = 2 concurrent senders
# 2 cpu = 4 concurrent senders
# 4 cpu = 8 concurrent senders
#
# Example values for multiplier 4:
#
# 1 cpu = 4 concurrent senders
# 2 cpu = 8 concurrent senders
# 4 cpu = 16 concurrent senders
#
# Example values for multiplier <1:
#
# 1 cpu = 1 concurrent sender
# 2 cpu = 1 concurrent sender
# 4 cpu = 1 concurrent sender
advanced-sender-multiplier: 2

# Array of string. Extra URIs to add to 'img-src' and 'media-src'
# when building the Content-Security-Policy header for your instance.
#
# This can be used to allow the browser to load resources from additional
# sources like S3 buckets and so on when viewing your instance's pages
# and profiles in the browser.
#
# Since non-proxying S3 storage will be probed on instance launch to
# generate a correct Content-Security-Policy, you probably won't need
# to ever touch this setting, but it's included in the 'spirit of more
# configurable (usually) means more good'.
# 
# See: https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP
#
# Example: ["s3.example.org", "some-bucket-name.s3.example.org"]
# Default: []
advanced-csp-extra-uris: []

# String. HTTP request header filtering mode to use for this instance.
#
# "block" -- only requests that are explicitly blocked by header filters
#            will be denied (unless they are also explicitly allowed).
#
# "allow" -- only requests that are explicitly allowed by header filters
#            will be accepted (unless they are also explicitly blocked).
#            This mode is considered experimental and will almost certainly
#            break access to your instance unless you are very careful.
#
#   ""    -- request header filtering disabled.
#
# For more details on block and allow modes, check the documentation at:
# https://docs.gotosocial.org/en/latest/admin/request_filtering_modes
#
# Options: ["block", "allow", ""]
# Default: ""
advanced-header-filter-mode: ""

# Bool. Enables a proof-of-work based deterrence against scrapers
# on profile and status web pages. This will generate a unique but
# deterministic challenge for each HTTP client to complete before
# accessing the above mentioned endpoints, on success being given
# a cookie that permits challenge-less access within a 1hr window.
#
# The outcome of this is that it should make scraping of these
# endpoints economically unfeasible, while having a negligible
# performance impact on your own instance.
#
# The downside is that it requires javascript to be enabled.
#
# For more details please check the documentation at:
# https://docs.gotosocial.org/en/latest/admin/scraper_deterrence
#
# Options: [true, false]
# Default: true
advanced-scraper-deterrence: false
```
