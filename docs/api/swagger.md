# API Documentation

## Rate limit

To prevent abuse of the API an IP-based HTTP rate limit is in place, a maximum of 300 requests in a 5 minutes time window are allowed, every response will include the current status of the rate limit with the following headers:

- `x-ratelimit-limit` maximum number of requests allowed per time period (fixed)
- `x-ratelimit-remaining` number of remaining requests that can still be performed
- `x-ratelimit-reset` unix timestamp when the rate limit will reset

In case the rate limit is exceeded an HTTP 429 error is returned to the caller.


GoToSocial uses [go-swagger](https://github.com/go-swagger/go-swagger) to generate a V2 [OpenAPI specification](https://swagger.io/specification/v2/) document from code annotations.

The resulting API documentation is rendered below, for quick reference.

If you'd like to do more with the spec, you can also view the [swagger.yaml](/api/swagger/swagger.yaml) directly, and then paste it into something like the [Swagger Editor](https://editor.swagger.io/) in order to autogenerate GoToSocial API clients in different languages, convert the doc to JSON or OpenAPI v3 specification, etc. See [here](https://swagger.io/tools/open-source/getting-started/) for more.

!!swagger swagger.yaml!!
