# Advanced

Advanced configuration options for GoToSocial.

## Can I host my instance at `fedi.example.org` but have just `@example.org` in my username?

Yes, you can! This is useful when you have something like a personal page or blog at `example.org`, but you also want your fediverse account to have `example.org` in it to avoid confusing people, or just because it looks nicer than `fedi.example.org`.

Please note that you need to do this *BEFORE RUNNING GOTOSOCIAL* for the first time, or things will likely break.

### Step 1: Configure GoToSocial

This step is easy.

In the settings, GoToSocial differentiates between `host`--the address at which your instance is accessible--and `account-domain`--which is the domain you want to show in accounts.

Behold, from the example config.yaml file:

```yaml
# String. Hostname that this server will be reachable at. Defaults to localhost for local testing,
# but you should *definitely* change this when running for real, or your server won't work at all.
# DO NOT change this after your server has already run once, or you will break things!
# Examples: ["gts.example.org","some.server.com"]
# Default: "localhost"
host: "localhost"

# String. Domain to use when federating profiles. This is useful when you want your server to be at
# eg., "gts.example.org", but you want the domain on accounts to be "example.org" because it looks better
# or is just shorter/easier to remember.
#
# To make this setting work properly, you need to redirect requests at "example.org/.well-known/webfinger"
# to "gts.example.org/.well-known/webfinger" so that GtS can handle them properly.
#
# You should also redirect requests at "example.org/.well-known/nodeinfo" in the same way.
# An empty string (ie., not set) means that the same value as 'host' will be used.
#
# DO NOT change this after your server has already run once, or you will break things!
#
# Please read the appropriate section of the installation guide before you go messing around with this setting:
# https://docs.gotosocial.org/installation_guide/advanced/#can-i-host-my-instance-at-fediexampleorg-but-have-just-exampleorg-in-my-username
#
# Examples: ["example.org","server.com"]
# Default: ""
account-domain: ""
```

The first value, `host`, is simple. In our scenario of wanting to run the GtS instance at `fedi.example.org`, this should be set to, yep, `fedi.example.org`.

The second value, `account-domain` should be set to `example.org`, to indicate that that's the domain we want accounts to be displayed with.

IMPORTANT: `account-domain` must be a *parent domain* of `host`, and `host` must be a *subdomain* of `account-domain`. So if your `host` is `fedi.example.org`, your `account-domain` cannot be `somewhere.else.com` or `example.com`, it **has to be** `example.org`.

### Step 2: Redirect from `example.org` to `fedi.example.org`

The next step is more difficult: we need to ensure that when remote instances search for the user `@user@example.org` via webfinger, they end up being pointed towards `fedi.example.org`, where our instance is actually hosted.

Of course, we don't want to redirect *all* requests from `example.org` to `fedi.example.org` because that negates the purpose of having a separate domain in the first place, so we need to be specific.

In the config.yaml above, there are two endpoints mentioned, both of which we need to redirect: `/.well-known/webfinger` and `/.well-known/nodeinfo`.

Assuming we have an [nginx](https://nginx.org) reverse proxy running on `example.org`, we can get the redirect behavior we want by adding the following to the nginx config for `example.org`:

```nginx
http {
    server {
        listen 80;
        listen [::]:80;
        server_name example.org;

        location /.well-known/webfinger {
            rewrite ^.*$ https://fedi.example.org/.well-known/webfinger permanent;
        }

        location /.well-known/nodeinfo {
            rewrite ^.*$ https://fedi.example.org/.well-known/nodeinfo permanent;
        }

        # The rest of our nginx config ...
    }
}
```

The above configuration [rewrites](https://www.nginx.com/blog/creating-nginx-rewrite-rules/) queries to `example.org/.well-known/webfinger` and `example.org/.well-known/nodeinfo` to their `fedi.example.org` counterparts, which means that query information is preserved, making it easier to follow the redirect.

### Step 3: What now?

Once you've done steps 1 and 2, proceed as normal with the rest of your GoToSocial installation.

### Supplemental: how does this work?

With the configuration we put in place in the steps above, when someone from another instance looks up `@user@example.org`, their instance will perform a webfinger request to `example.org/.well-known/webfinger?resource:acct=user@example.org` in order to discover a link to an ActivityPub representation of that user's account. They will then be redirected to  `https://fedi.example.org/.well-known/webfinger?resource:acct=user@example.org`, and their query will be resolved.

The webfinger response returned by GoToSocial (and indeed Mastodon, and other ActivityPub implementations) contains the desired account domain in the `subject` part of the response, and provides links to aliases that should be used to query the account.

Here's an example of this working for the `superseriousbusiness.org` GoToSocial instance, which is hosted at `gts.superseriousbusiness.org`.

Curl query:

```bash
curl -v 'https://superseriousbusiness.org/.well-known/webfinger?resource=acct:@gotosocial@superseriousbusiness.org'
```

Response:

```text
> GET /.well-known/webfinger?resource=acct:@gotosocial@superseriousbusiness.org HTTP/2
> Host: superseriousbusiness.org
> user-agent: curl/7.68.0
> accept: */*
> 
< HTTP/2 301 
< content-type: text/html
< date: Thu, 17 Nov 2022 11:10:39 GMT
< location: https://gts.superseriousbusiness.org/.well-known/webfinger?resource=acct:@gotosocial@superseriousbusiness.org
< server: nginx/1.20.1
< content-length: 169
< 
<html>
<head><title>301 Moved Permanently</title></head>
<body>
<center><h1>301 Moved Permanently</h1></center>
<hr><center>nginx/1.20.1</center>
</body>
</html>

```

If we follow the redirect and make a query to the specified `location` as follows:

```bash
curl -v 'https://gts.superseriousbusiness.org/.well-known/webfinger?resource=acct:@gotosocial@superseriousbusiness.org'
```

Then we get the following response:

```json
{
  "subject": "acct:gotosocial@superseriousbusiness.org",
  "aliases": [
    "https://gts.superseriousbusiness.org/users/gotosocial",
    "https://gts.superseriousbusiness.org/@gotosocial"
  ],
  "links": [
    {
      "rel": "http://webfinger.net/rel/profile-page",
      "type": "text/html",
      "href": "https://gts.superseriousbusiness.org/@gotosocial"
    },
    {
      "rel": "self",
      "type": "application/activity+json",
      "href": "https://gts.superseriousbusiness.org/users/gotosocial"
    }
  ]
}
```

In the above response, note that the `subject` of the response contains the desired account-domain of `superseriousbusiness.org`, whereas the links contain the actual host value of `gts.superseriousbusiness.org`.

## Can I make my GoToSocial instance use a proxy (http, https, socks5) for outgoing requests?

Yes! GoToSocial supports canonical environment variables for doing this: `HTTP_PROXY`, `HTTPS_PROXY` and `NO_PROXY` (or the lowercase versions thereof). `HTTPS_PROXY` takes precedence over `HTTP_PROXY` for https requests.

The http client that GoToSocial uses will be initialized with the appropriate proxy.

The environment values may be either a complete URL or a `host[:port]`, in which case the "http" scheme is assumed. The schemes "http", "https", and "socks5" are supported.

## Application sandboxing

Although GoToSocial does not currently have any known vulnerabilities, it's
always a good idea to be proactive about security. One way you can help protect
your instance is to run it in a *sandbox* -- an environment that constrains the
actions a program can perform in order to limit the impact of a future exploit.

[Using Docker](../../installation_guide/docker) to run GoToSocial can work as a
(limited) sandboxing mechanism. For Linux installations, [Linux Security
Modules](https://en.wikipedia.org/wiki/Linux_Security_Modules) such as
[AppArmor](https://www.apparmor.net/) and
[SELinux](https://en.wikipedia.org/wiki/Security-Enhanced_Linux) work as a
complementary mechanism that typically provide stronger protections. You should
use

- **AppArmor** if you're running GoToSocial on Debian, Ubuntu, or OpenSUSE, and
- **SELinux** if you're using CentOS, RHEL, or Rocky Linux.

For other Linux distributions, you will need to look up what Linux Security
Modules are supported by your kernel.

!!! note
    GoToSocial is currently alpha software, and as more features are implemented
    these security policies may quickly become outdated. You may find that using
    AppArmor or SELinux causes GoToSocial to fail in unexpected ways until GTS
    becomes stable.

!!! caution
    Sandboxing is an _additional_ security mechanism to help defend against
    certain kinds of attacks; it _is not_ a replacement for good security
    practices.

### AppArmor

For Linux distributions supporting AppArmor, there is an AppArmor profile
available in `example/apparmor/gotosocial` that you can use to confine your
GoToSocial instance. If you're using a server (such as a VPS) to deploy
GoToSocial, you can install the AppArmor profile by downloading it and copying
it into the `/etc/apparmor.d/` directory:

```bash
wget https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/example/apparmor/gotosocial
sudo install -o root -g root gotosocial /etc/apparmor.d/gotosocial
sudo apparmor_parser -Kr /etc/apparmor.d/gotosocial
```

If you're using Docker Compose, you should add the following `security_opt`
section to your Compose configuration file:

```yaml
services:
  gotosocial:
    ...
    security_opt:
      - apparmor=gotosocial
```

If you're running GoToSocial as a Systemd service, you should instead add this
line under `[Service]`:

```ini
[Service]
...
AppArmorProfile=gotosocial
```

For other deployment methods (e.g. a managed Kubernetes cluster), you should
review your platform's documentation for how to deploy an application with an
AppArmor profile.

#### Disabling the AppArmor profile

If enabling the AppArmor profile causes your instance to experience issues, you
can uninstall it from the system as follows:

```
sudo apparmor_parser -R /etc/apparmor.d/gotosocial
sudo rm -vi /etc/apparmor.d/gotosocial
```

You will also want to remove any changes you made to your Compose configuration
or Systemd service file to enable the profile.

### SELinux

!!! note
    Currently, this SELinux policy only works for the [binary installation
    method](../../installation_guide/binary).

If SELinux is available on your system, you can optionally install [SELinux
policy](https://github.com/lzap/gotosocial-selinux) to further improve security.
