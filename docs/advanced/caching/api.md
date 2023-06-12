# Caching API responses

It is possible to cache certain API responses to offload the GoToSocial process from having to handle all requests. We don't recommend caching responses to requests under `/api`.

When using a [split domain](../host-account-domain.md) deployment style, you need to ensure you configure caching on the host domain. The account domain should only be issuing redirects to the host domain which clients will automatically remember.

!!! warning "There are only two hard things in computer science"
    Configuring caching incorrectly can result into all kinds of problems. Follow this guide carefully and thoroughly test your modifications. Don't cache endpoints that require authentication without taking the `Authorization` header into account.

## Endpoints

### Webfinger and hostmeta

Requests to `/.well-known/webfinger` and `/.well-known/host-meta` can be safely cached. Do be careful to ensure any caching strategy takes query parameters into account when caching webfinger requests as requests to that endpoint are of the form `?resource=acct:@username@domain.tld`.

### Public keys

Many implementations will regularly request the public key for a user in order to validate the signature on a message they received. This will happen whenever a message gets federated amongst other things. These keys are long lived, essentially eternal, and can thus be cached with a long lifetime.

## Configuration snippets

### nginx

For nginx, you'll need to start by configuring a cache zone. The cache zone must be created in the `http` section, not within `server` or `location`.

```nginx
http {
    ...
    proxy_cache_path /var/cache/nginx keys_zone=gotosocial_ap_public_responses:10m inactive=1w;
}
```

This configures a cache of 10MB whose entries will be kept up to one week if they're not accessed.

The zone is named `gotosocial_ap_public_responses` but you can name it whatever you want. 10MB is a lot of cache keys; you can probably use a smaller value on small instances.

Second, we need to update our GoToSocial nginx configuration to actually use the cache for the endpoints we want to cache.

```nginx
server {
  server_name social.example.org;
  
  location ~ /.well-known/(webfinger|host-meta)$ {
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header X-Forwarded-Proto $scheme;

    proxy_cache gotosocial_ap_public_responses;
    proxy_cache_background_update on;
    proxy_cache_key $scheme://$host$uri$is_args$query_string;
    proxy_cache_valid 200 10m;
    proxy_cache_use_stale error timeout updating http_500 http_502 http_503 http_504 http_429;
    proxy_cache_lock on;
    add_header X-Cache-Status $upstream_cache_status;

    proxy_pass http://localhost:8080;
  }

  location ~ ^\/users\/(?:[a-z0-9_\.]+)\/main-key$ {
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header X-Forwarded-Proto $scheme;

    proxy_cache gotosocial_ap_public_responses;
    proxy_cache_background_update on;
    proxy_cache_key $scheme://$host$uri;
    proxy_cache_valid 200 604800s;
    proxy_cache_use_stale error timeout updating http_500 http_502 http_503 http_504 http_429;
    proxy_cache_lock on;
    add_header X-Cache-Status $upstream_cache_status;

    proxy_pass http://localhost:8080;
  }
```

The `proxy_pass` and `proxy_set_header` are mostly the same, but the `proxy_cache*` entries warrant some explanation:

- `proxy_cache gotosocial_ap_public_responses` tells nginx to use the `gotosocial_ap_public_responses` cache zone we previously created. If you named it something else, you should change this value
- `proxy_cache_background_update on` means nginx will try and refresh a cached resource that's about to expire in the background, to ensure it has a current copy on disk
- `proxy_cache_key` is configured in such a way that it takes the query string into account for caching. So a request for `.well-known/webfinger?acct=user1@example.org` and `.well-known/webfinger?acct=user2@example.org` are not seen as the same.
- `proxy_cache_valid 200 10m;` means we only cache 200 responses from GTS and for 10 minutes. You can add additional lines of these, like `proxy_cache_valid 404 1m;` to cache 404 responses for 1 minute
- `proxy_cache_use_stale` tells nginx it's allowed to use a stale cache entry (so older than 10 minutes) in certain cases
- `proxy_cache_lock on` means that if a resource is not cached and there's multiple concurrent requests for them, the queries will be queued up so that only one request goes through and the rest is then answered from cache
- `add_header X-Cache-Status $upstream_cache_status` will add an `X-Cache-Status` header to the response so you can check if things are getting cached. You can remove this.

The provided configuration will serve a stale response in case there's an error proxying to GoToSocial, if our connection to GoToSocial times out, if GoToSocial returns a `5xx` status code or if GoToSocial returns 429 (Too Many Requests). The `updating` value says that we're allowed to serve a stale entry if nginx is currently in the process of refreshing its cache. Because we configured `inactive=1w` in the `proxy_cache_path` directive, nginx may serve a response up to one week old if the conditions in `proxy_cache_use_stale` are met.
