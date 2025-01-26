# Reverse proxy

GoToSocial can be exposed directly to the internet. However, many folks prefer to have a reverse proxy handle connections from the outside instead. This can also give greater control over TLS configurations and enables some more advanced scenario's like asset caching.

## General procedure

In order to use a reverse-proxy, you'll typically want to do a few things:

* Configure some way to get TLS certificates for the host domain
* Bind GoToSocial to a local IP instead of a public IP and a non-priviledged port. Adjust the `bind-address` and `port` configuration options
* Disable Lets Encrypt in GoToSocial if you were using it. Set `letsencrypt-enabled` to `false`
* Configure the reverse proxy to handle TLS and proxy requests to GoToSocial

!!! warning
    Do not change the value of the `host` configuration option. This needs to remain the actual domain name the instance is running on as seen by other instances on the internet. Instead, change the `bind-address` and update the `port` and `trusted-proxies`.

### Container

When you deploy GoToSocial using our [example Docker Compose guide](../installation/container.md), it will bind to port `443` by default as it assumes you want to directly expose it to the internet. In order to run it behind a reverse proxy, you need to change that.

In the compose file:

* Comment out the `- "443:8080"` line in the `ports` definition
* If you had enabled Lets Encrypt support:
    * Comment out the `- "80:80"` line in the `ports` definition
    * Set `GTS_LETSENCRYPT_ENABLED` back to `"false"` or comment it out
* Uncomment the `- "127.0.0.1:8080:8080"` line instead

This now causes Docker to only forward connections on `127.0.0.1` on port `8080` to the container, effectively isolating it from the outside world. You can now tell your reverse-proxy to send requests there instead.

## Guides

We have guides available for the following servers:

* [nginx](nginx.md)
* [Apache httpd](apache-httpd.md)
* [Caddy 2](caddy.md)

## WebSockets

When using a reverse-proxy, special care must be taken to allow WebSockets to work too. This is necessary as many client applications use WebSockets to stream your timeline. WebSockets is not used as part of federation.

Make sure you read the [WebSocket](websocket.md) documentation and configure your reverse proxy accordingly.

## Trusted Proxies

When using a reverse-proxy, you may run into issues with rate limiting and `trusted-proxies`. Check the [trusted proxies](../../configuration/trusted_proxies.md) documentation if you have any problems.
