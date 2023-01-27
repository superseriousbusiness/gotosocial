# Request Throttling and Rate Limiting

GoToSocial applies http request throttling and rate limiting to the ActivityPub API endpoints (inboxes, user endpoints, emojis, etc).

This ensures that remote servers cannot flood a GoToSocial instance with spurious requests. Instead, remote servers making GET or POST requests to the ActivityPub API endpoints should respect 429 and 503 http codes, and take account of the `retry-after` http response header.

For more details on request throttling and rate limiting behavior, please see the [throttling](../../api/throttling.md) and [rate limiting](../../api/ratelimiting.md) documents.
