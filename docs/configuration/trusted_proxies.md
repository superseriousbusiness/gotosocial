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

There are some situations where it's not practically possible to get `trusted-proxies` configured correctly to detect the real client IP of incoming requests For example, if you're running GoToSocial behind a home internet router that cannot inject an `X-Forwarded-For` header, then your suggested entry to add to `trusted-proxies` will look something like `192.168.x.x`, but adding this to `trusted-proxies` won't resolve the issue.

If you've tried everything, then you can disable the warning message by just turning off rate limiting entirely, ie., by setting `advanced-rate-limit-requests` to 0 in your config.yaml, or setting the environment variable `GTS_ADVANCED_RATE_LIMIT_REQUESTS` to 0. Don't forget to **restart your instance** after changing this setting.
