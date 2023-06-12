# Provisioning TLS certificates

As explained in the [deployment considersations](../getting_started/index.md), federation requires the use of TLS as most instances refuse to federate over unencrypted transports.

GoToSocial comes with built-in support for provisioning and renewing its own TLS certificates through Lets Encrypt. This guide looks at how you can provision your own certificates independently from GoToSocial. This can be useful if you want full control over how the certificates are provisioned, or because you're using a [reverse proxy](../getting_started/reverse_proxy/index.md) which is doing TLS termination.

There are a few different ways you can get TLS certificates:

* Buy them from a vendor, typically valid for 2 years
* Get them from your cloud provider, validity depends on their product constraints
* Get them from an [ACME](https://en.wikipedia.org/wiki/Automatic_Certificate_Management_Environment)-compatible provider like Lets Encrypt, typically valid for 3 months at a time

In this guide we'll only look at option 3, an ACME-compatible vendor.

## General approach

The way you'll provision certificates through Lets Encrypt is:

* Install an ACME client on your server
* Configure the ACME client to provision your certificates
* Configure a piece of software to use those certificates
* Enable a timer/cron to regularly renew the certificates
* Signal to the necessary applications they need to reload or restart to pick up the new certificates

Certificates are provisioned [using a challenge](https://letsencrypt.org/sv/docs/challenge-types/), a way to verify that you're requesting a certificate for a domain you control. You'll typically use one of:

* HTTP challenge
* DNS challenge

The HTTP challenge requires serving certain files on port 80 on the domains you're requesting a certificate for under the `/.well-known/acme/` path. This is the default challenge type.

The DNS challenge happens entirely out of band but requires you to update a DNS TXT record. This approach is only feasible if your DNS registrar provides an API through which you can modify DNS records so that your ACME client can complete this challenge.

## Clients

The official Lets Encrypt client is [certbot](https://certbot.eff.org/) and it's usually packaged in [your (Linux) distribution](https://repology.org/project/certbot/versions) of choice. Certain reverse proxies like Caddy and Traefik have built-in support for provisioning certificates using the ACME protocol.

A couple of other clients of note that you can consider using:

* [acme-client](https://man.openbsd.org/acme-client.1) for OpenBSD using the privilege separation features of the platform
* [lacme](https://git.guilhem.org/lacme/about/), which is built with process isolation and minimal privileges in mind in the same vein as acme-client but for Linux
* [Lego](https://github.com/go-acme/lego), an ACME client and library written in Go
* [mod_md](https://httpd.apache.org/docs/2.4/mod/mod_md.html), when using Apache 2.4.30+

### DNS challenge

For the DNS challenge, the API of your registrar needs to be supported by your ACME client. Though certbot has a few plugins for popular providers, you probably want to look at the [dns-multi](https://github.com/alexzorin/certbot-dns-multi) plugin instead. It leverages [Lego](https://github.com/go-acme/lego) under the hood which supports a much wider array of providers.

## Configuration

There are 3 configuration options that are important:

* [`letsencrypt-enabled`](../configuration/tls.md) controls if GoToSocial will try to provision its own certificates
* [`tls-certificate-chain`](../configuration/tls.md) filesystem path where GoToSocial can find the TLS certificate chain + the public key
* [`tls-certificate-key`](../configuration/tls.md) filesystem path where GoToSocial can find the associated TLS private key

### Without reverse proxy

When running GoToSocial directly exposed to the internet, but you still want to use your own certificates you can set the following options:

```yaml
letsencrypt-enabled: false
tls-certificate-chain: "/path/to/combined-certificate-chain-public.key"
tls-certificate-key: "/path/to/private.key"
```

This disables the built-in provisioning of certificates through Lets Encrypt and tells GoToSocial where to find the certificates on disk.

!!! tip
    Restart GoToSocial after renewing your certificates. It won't pick up on certificate rotation by itself when they are provided like this.

### With reverse proxy

When running GoToSocial behind a [reverse proxy](../getting_started/reverse_proxy/index.md) which you also use for TLS termination, you'll need the following instead:

```yaml
letsencrypt-enabled: false
tls-certificate-chain: ""
tls-certificate-key: ""
```

It's important to ensure the `tls-certificate-*` options are unset or set to the empty string. Otherwise GoToSocial will attempt to handle TLS itself.

!!! danger "Protocol configuration option"
    Do **not** change the [`protocol`](../configuration/general.md) configuration option to `http`. This should only ever by set to `http` for development purposes. It needs to be set to `https` even when running behind a TLS-terminating reverse proxy.

You'll also want to change the [`port`](../configuration/general.md) GoToSocial binds on, so it no longer tries to use port 443.

To configure TLS in your reverse proxy, please refer to their documentation:

* [nginx](https://docs.nginx.com/nginx/admin-guide/security-controls/terminating-ssl-http/)
* [apache](https://httpd.apache.org/docs/2.4/ssl/ssl_howto.html)
* [Traefik](https://doc.traefik.io/traefik/https/tls/)
* [Caddy](https://caddyserver.com/docs/caddyfile/directives/tls)

!!! tip
    When configuring TLS in your reverse proxy, ensure you configure a reasonably modern set of compatible versions and ciphers. You can use the "Intermediate" configuration from the [Mozilla SSL Configuration Generator](https://ssl-config.mozilla.org/).

    Check the documentation of your reverse proxy on whether you have to reload or restart it after certificates have changed. Not all reverse proxies detect this.

## Guides

There are a number of good resources on the internet explaining how to set all of this up.

* [ArchWiki entry](https://wiki.archlinux.org/title/certbot) on certbot
* [Gentoo wiki entry](https://wiki.gentoo.org/wiki/Let%27s_Encrypt) on Lets Encrypt
* [Linode guide](https://www.linode.com/docs/guides/enabling-https-using-certbot-with-nginx-on-fedora/) on certbot for Fedora, RHEL/CentOS, Debian and Ubuntu
* Digital Ocean guides on Lets Encrypt on Ubuntu 22.04 with [nginx](https://www.digitalocean.com/community/tutorials/how-to-secure-nginx-with-let-s-encrypt-on-ubuntu-22-04) or [apache](https://www.digitalocean.com/community/tutorials/how-to-secure-apache-with-let-s-encrypt-on-ubuntu-22-04)
