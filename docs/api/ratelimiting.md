# Request Rate Limiting

To mitigate abuse + scraping of your instance, IP-based HTTP rate limiting is in place.

There are separate rate limiters configured for different groupings of endpoints. In other words, being rate limited for one part of the API doesn't necessarily mean you will be rate limited for other parts. Each entry in the following list has a separate rate limiter:

- `/users/*` and `/emoji/*` - ActivityPub (s2s) endpoints.
- `/auth/*` and `/oauth/*` - Sign in + OAUTH token requests.
- `/fileserver/*` - Media attachments, emojis, etc.
- `/nodeinfo/*` - NodeInfo endpoint(s).
- `/.well-known/*` - webfinger + nodeinfo requests.

By default, each rate limiter allows a maximum of 300 requests in a 5 minute time window: 1 request per second per client IP address.

Every response will include the current status of the rate limit with the following headers:

- `X-Ratelimit-Limit`: maximum number of requests allowed per time period.
- `X-Ratelimit-Remaining`: number of remaining requests that can still be performed within.
- `X-Ratelimit-Reset`: ISO8601 timestamp indicating when the rate limit will reset.

In case the rate limit is exceeded, an [HTTP 429 Too Many Requests](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/429) error is returned to the caller.

## Rate Limiting FAQs

### My rate limit keeps being exceeded! Why?

If you find that your rate limit is regularly being exceeded (both for yourself and other callers) during normal use of your instance, it may be that GoToSocial can't tell the clients apart by IP address. You can investigate this by viewing the logs of your instance. If (almost) all logged client IP addresses appear to be the same IP address (something like `172.x.x.x`), then the rate limiting will cause problems.

This happens when your server is running inside NAT (port forwarding), or behind an HTTP proxy without the correct configuration, causing your instance to see all incoming IP addresses as the same address: namely, the IP address of your reverse proxy or gateway. This means that all incoming requests are *sharing the same rate limit*, rather than being split correctly per IP.

If you are using an HTTP proxy then it's likely that your `trusted-proxies` is not correctly configured. See the [trusted-proxies](../configuration/trusted_proxies.md) documentation for more info on how to resolve this.

If you don't have an HTTP proxy, then it's likely caused by NAT. In this case you should disable rate limiting altogether.

### Can I configure the rate limit? Can I just turn it off?

Yes! Set `advanced-rate-limit-requests: 0` in the config.

### Can I exclude one or more IP addresses from rate limiting, but leave the rest in place?

Yes! Set `advanced-rate-limit-exceptions` in the config.
