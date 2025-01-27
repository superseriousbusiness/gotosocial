# Trusted Proxies

To correctly enforce [rate limiting](../api/ratelimiting.md), GoToSocial relies on the concept of "trusted proxies" in order to accurately determine the IP address of clients accessing your server.

A "trusted proxy" is an intermediate network hop that GoToSocial can be instructed to trust to provide a correct client IP address.

For example, if you are running in a reverse proxy configuration with Docker + Nginx, then the Docker network address of Nginx should be configured as a trusted proxy, since all traffic from the wider internet will come into GoToSocial via Nginx.

Without setting `trusted-proxies` correctly, GoToSocial will see all incoming client IP addresses as the same address, which leads to rate limiting issues, since GoToSocial uses client IP addresses to bucket rate limits.

## tl;dr: How to set `trusted-proxies` correctly

If your `trusted-proxies` setting is not correctly configured, you may see the following warning on the web view of your instance (v0.18.0 and above):

> Warning! It looks like trusted-proxies is not set correctly in this instance's configuration. This may cause rate-limiting issues and, by extension, federation issues.
>
> If you are the instance admin, you should fix this by adding `SUGGESTED_IP_RANGE` to your trusted-proxies. 

To resolve this, copy the IP range in the message, and edit your `config.yaml` file to add the IP range to your `trusted-proxies`.

!!! tip "You may be getting rate limited even if you don't see the above warning!"
    If you're on a version of GoToSocial below v0.18.0, or you're running behind a CDN such as Cloudflare (not recommended), you won't see a warning message. Instead, you'll see in your GoToSocial logs that all client IPs are the same address. In this case, take the recurring client IP value as `SUGGESTED_IP_RANGE`.

In this example, we assume `SUGGESTED_IP_RANGE` to be `172.17.0.1/16` (the default Docker bridge network subnet).

Before (default config):

```yaml
trusted-proxies:
  - "127.0.0.1/32"
  - "::1"
```

After (new config):

```yaml
trusted-proxies:
  - "172.17.0.1/16"
  - "127.0.0.1/32"
  - "::1"
```

If you are using [environment variables](../configuration/index.md#environment-variables) to configure your instance, you can configure `trusted-proxies` by setting the environment variable `GTS_TRUSTED_PROXIES` to a comma-separated list of IP ranges, like so:

```env
GTS_TRUSTED_PROXIES="172.17.0.1/16,127.0.0.1/32,::1"
```

If you are using docker compose, your docker-compose.yaml file should look something like this after the change (note that yaml uses `: ` and not `=`):

```yaml
################################
# BLAH BLAH OTHER CONFIG STUFF #
################################
    environment:
      ############################
      # BLAH BLAH OTHER ENV VARS #
      ############################
      ## For reverse proxy setups:
      GTS_TRUSTED_PROXIES: "172.17.0.1/16,127.0.0.1/32,::1"
################################
# BLAH BLAH OTHER CONFIG STUFF #
################################
```

Once you have made the necessary configuration changes, **restart your instance** and refresh the home page.

If the message is gone, then the problem is resolved!

If you still see the warning message but with a different suggested IP range to add to `trusted-proxies`, then follow the same steps as above again, including the new suggested IP range in your config in addition to the one you just added.

!!! tip "Cloudflare IP Addresses"
    If you are running with a CDN/proxy such as Cloudflare in front of your GoToSocial instance (not recommended), then you may need to add one or more of the Cloudflare IP addresses to your `trusted-proxies` in order to have rate limiting work properly. You can find a list of Cloudflare IP addresses here: https://www.cloudflare.com/ips/

## I can't seem to get `trusted-proxies` configured properly, can I just disable the warning?

There are some situations where it's not practically possible to get `trusted-proxies` configured correctly to detect the real client IP of incoming requests, or where the real client IP is accurate but still shows as being within a private network.

For example, if you're running GoToSocial on your home network, behind a home internet router that cannot inject an `X-Forwarded-For` header, then your suggested entry to add to `trusted-proxies` will look something like `192.168.x.x`, but adding this to `trusted-proxies` won't resolve the issue.

Another example: you're running GoToSocial on your home network, behind a home internet router, and you are accessing the web frontend from a device that's *also* on your home network, like your laptop or phone. In this case, your router may send you directly to your GoToSocial instance without your request ever leaving the network, and so GtS will correctly see *your* client IP address as a private network address, but *other* requests coming in from the wider internet will show their real remote client IP addresses. In this scenario, the `trusted-proxies` warning does not really apply.

If you've tried editing your `trusted-proxies` setting, but you still see the warning, then it's likely that one of the above examples applies to you. You can proceed in one of two ways:

### Add specific exception for your home network (preferred)

If the suggested IP range in the `trusted-proxies` warning looks something like `192.168.x.x`, but you still see other client IPs in your GoToSocial logs that don't start with `192.168`, then try adding a rate limiting exception only for devices on your home network, while leaving rate limiting in place for outside IP addresses.

For example, if your suggestion is something like `192.168.1.128/32`, then swap the `/32` for `/24` so that the range covers `192.168.1.0` -> `192.168.1.255`, and add this to the `advanced-rate-limit-exceptions` setting in your `config.yaml` file.

Before (default config):

```yaml
advanced-rate-limit-exceptions: []
```

After (new config):

```yaml
advanced-rate-limit-exceptions:
  - "192.168.1.128/24"
```

If you are using [environment variables](../configuration/index.md#environment-variables) to configure your instance, you can configure `advanced-rate-limit-exceptions` by setting the environment variable `GTS_ADVANCED_RATE_LIMIT_EXCEPTIONS` to a comma-separated list of IP ranges, like so:

```env
GTS_ADVANCED_RATE_LIMIT_EXCEPTIONS="192.168.1.128/24"
```

If you are using docker compose, your docker-compose.yaml file should look something like this after the change (note that yaml uses `: ` and not `=`):

```yaml
################################
# BLAH BLAH OTHER CONFIG STUFF #
################################
    environment:
      ############################
      # BLAH BLAH OTHER ENV VARS #
      ############################
      GTS_ADVANCED_RATE_LIMIT_EXCEPTIONS: "192.168.1.128/24"
################################
# BLAH BLAH OTHER CONFIG STUFF #
################################
```

Once you have made the necessary configuration changes, **restart your instance** and refresh the home page.

### Turn off rate limiting entirely (last resort)

If nothing else works, you can disable rate limiting entirely, which will also disable the `trusted-proxies` check and warning.

!!! warning
    Turning off rate limiting entirely should be considered a last resort, as rate limiting helps protect your instance from spam and scrapers.

To turn off rate limiting, set `advanced-rate-limit-requests` to 0 in your `config.yaml`.

Before (default config):

```yaml
advanced-rate-limit-requests: 300
```

After (new config):

```yaml
advanced-rate-limit-requests: 0
```

If you are using [environment variables](../configuration/index.md#environment-variables) to configure your instance, you can configure `advanced-rate-limit-requests` by setting the environment variable `GTS_ADVANCED_RATE_LIMIT_REQUESTS` to 0, like so:

```env
GTS_ADVANCED_RATE_LIMIT_REQUESTS="0"
```

If you are using docker compose, your docker-compose.yaml file should look something like this after the change (note that yaml uses `: ` and not `=`):

```yaml
################################
# BLAH BLAH OTHER CONFIG STUFF #
################################
    environment:
      ############################
      # BLAH BLAH OTHER ENV VARS #
      ############################
      GTS_ADVANCED_RATE_LIMIT_REQUESTS: "0"
################################
# BLAH BLAH OTHER CONFIG STUFF #
################################
```

Once you have made the necessary configuration changes, **restart your instance** and refresh the home page.
