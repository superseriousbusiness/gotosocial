# Request Throttling

GoToSocial uses request throttling to limit the number of open connections to the API of your instance. This is designed to prevent your instance from accidentally being DDOS'd (aka [the hug of death](https://en.wikipedia.org/wiki/Slashdot_effect)) if a post gets boosted or replied to by an account with many thousands of followers.

Throttling means that only a limited number of HTTP requests to the API will be handled concurrently, in order to provide a snappy response to each request and move on quickly. The rationale is that it's better to handle fewer requests quickly, than to try to handle all incoming requests at once and take multiple seconds per request.

Throttling limits are applied across router groups, similar to the way that [rate limiting](./ratelimiting.md) is organized, so if one part of the API is currently being throttled, that doesn't mean they all are.

Throttling limits are calculated based on the number of CPUs available to GoToSocial, and the configuration value `advanced-throttling-multiplier`. The calculation is performed as follows:

- In-process queue limit = number of CPUs * CPU multiplier.
- Backlog queue limit = in-process queue limit * CPU multiplier.

This leads to the following values for the default multiplier (8):

```text
1 cpu = 08 in-process, 064 backlog
2 cpu = 16 in-process, 128 backlog
4 cpu = 32 in-process, 256 backlog
8 cpu = 64 in-process, 512 backlog
```

New requests that overflow the in-process limit are held in the backlog queue, and processed as soon as a spot is freed up (ie., when a currently in-process request is finished). Requests that cannot be processed, and cannot fit in the backlog queue will be responded to with http code [503 - Service Unavailable](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/503), and the `Retry-After` header will be set to `30` (seconds), to indicate that the caller should try again later.

Requests are not held in the backlog queue indefinitely: if requests in the backlog cannot be processed within 30 seconds of being received, they will also receive a code 503 and a 30s retry-after.

## Throttling FAQs

### Can I tune the request throttling?

Yes, just change the value of `advanced-throttling-multiplier` higher (if you have very powerful CPUs) or lower (if you have relatively less powerful CPUs).

### Can I disable the request throttling?

Yes. To do so, just set `advanced-throttling-multiplier` to `0` or less. This will disable HTTP request throttling entirely, and instead attempt to process all incoming requests at once. This is useful in cases where you want to do request throttling using an external service or a reverse-proxy, and you don't want GoToSocial to interfere with your setup.
