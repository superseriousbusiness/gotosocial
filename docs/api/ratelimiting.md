# Rate Limit

To mitigate abuse + scraping of your instance, an IP-based HTTP rate limit is in place.

This rate limit applies not just to the API, but to all requests (web, federation, etc).

By default, a maximum of 1000 requests in a 5 minute time window are allowed.

Every response will include the current status of the rate limit with the following headers:

- `X-Ratelimit-Limit`: maximum number of requests allowed per time period.
- `X-Ratelimit-Remaining`: number of remaining requests that can still be performed within.
- `X-Ratelimit-Reset`: unix timestamp indicating when the rate limit will reset.

In case the rate limit is exceeded, an [HTTP 429 Too Many Requests](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/429) error is returned to the caller.

## Rate Limiting FAQs

### My rate limit keeps being exceeded! Why?

If you find that your rate limit is regularly being exceeded (both for yourself and other callers) during normal use of your instance, it may be that GoToSocial can't tell the clients apart by IP address. You can investigate this by viewing the logs of your instance. If (almost) all logged IP addresses appear to be the same IP address (something like `172.x.x.x`), then the rate limiting will cause problems.

This happens when your server is running inside NAT (port forwarding), or behind an HTTP proxy without the correct configuration, causing your instance to see all incoming IP addresses as the same address: namely, the IP address of your reverse proxy or gateway. This means that all incoming requests are *sharing the same rate limit*, rather than being split correctly per IP.

If you are using an HTTP proxy then it's likely that your `trusted-proxies` is not correctly configured. If this is the case, try adding the IP address of your reverse proxy to the list of `trusted-proxies`, and restarting your instance.

If you don't have an HTTP proxy, then it's likely caused by NAT. In this case you should disable rate limiting altogether.

### Can I configure the rate limit? Can I just turn it off?

Yes! Set `advanced-rate-limit-requests: 0` in the config.
