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

If `example.org` is running on [Traefik](https://doc.traefik.io/traefik/), we could use labels similar to the following to setup the redirect.

```docker
  myservice:
    image: foo
    # Other stuff
    labels:
      - 'traefik.http.routers.myservice.rule=Host(`example.org`)'
      - 'traefik.http.middlewares.myservice-gts.redirectregex.permanent=true'
      - 'traefik.http.middlewares.myservice-gts.redirectregex.regex=^https://(.*)/.well-known/(webfinger|nodeinfo)$$'
      - 'traefik.http.middlewares.myservice-gts.redirectregex.replacement=https://fedi.$${1}/.well-known/$${2}'
      - 'traefik.http.routers.myservice.middlewares=myservice-gts@docker'
```

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

## nginx

This section contains a number of additional things for configuring nginx.

### Extra Hardening

If you want to harden up your NGINX deployment with advanced configuration options, there are many guides online for doing so ([for example](https://beaglesecurity.com/blog/article/nginx-server-security.html)). Try to find one that's up to date. Mozilla also publishes best-practice ssl configuration [here](https://ssl-config.mozilla.org/).

### Caching Webfinger

It's possible to use nginx to cache the webfinger responses. This may be useful in order to ensure clients still get a response on the webfinger endpoint even if GTS is (temporarily) down.

You'll need to configure two things:
* A cache path
* An additional `location` block for webfinger

First, the cache path which needs to happen in the `http` section, usually inside your `nginx.conf`:

```nginx.conf
http {
  ... there will be other things here ...
  proxy_cache_path /var/cache/nginx keys_zone=ap_webfinger:10m inactive=1w;
}
```

This configures a cache of 10MB whose entries will be kept up to one week if they're not accessed. The zone is named `ap_webfinger` but you can name it whatever you want. 10MB is a lot of cache keys, you can probably use a much smaller value on small instances.

Second, actually use the cache for webfinger:

```nginx.conf
server {
  server_name example.org;
  location /.well-known/webfinger {
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header X-Forwarded-Proto $scheme;

    proxy_cache ap_webfinger;
    proxy_cache_background_update on;
    proxy_cache_key $scheme://$host$uri$is_args$query_string;
    proxy_cache_valid 200 10m;
    proxy_cache_use_stale error timeout updating http_500 http_502 http_503 http_504 http_429;
    proxy_cache_lock on;
    add_header X-Cache-Status $upstream_cache_status;

    proxy_pass http://localhost:8080;
  }

  location / {
    proxy_pass http://localhost:8080/;
    proxy_set_header Host $host;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header X-Forwarded-Proto $scheme;
  }
  client_max_body_size 40M;

  listen [::]:443 ssl ipv6only=on; # managed by Certbot
  listen 443 ssl; # managed by Certbot
  ssl_certificate /etc/letsencrypt/live/example.org/fullchain.pem; # managed by Certbot
  ssl_certificate_key /etc/letsencrypt/live/example.org/privkey.pem; # managed by Certbot
  include /etc/letsencrypt/options-ssl-nginx.conf; # managed by Certbot
  ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem; # managed by Certbot
}
```

The `proxy_pass` and `proxy_set_header` are mostly the same, but the `proxy_cache*` entries warrant some explanation:

* `proxy_cache ap_webfinger` tells it to use the `ap_webfinger` cache zone we previously created. If you named it something else, you should change this value
* `proxy_cache_background_update on` means nginx will try and refresh a cached resource that's about to expire in the background, to ensure it has a current copy on disk
* `proxy_cache_key` is configured in such a way that it takes the query string into account for caching. So a request for `.well-known/webfinger?acct=user1@example.org` and `.well-known/webfinger?acct=user2@example.org` are not seen as the same
* `proxy_cache_valid 200 10m;` means we only cache 200 responses from GTS and for 10 minutes. You can add additional lines of these, like `proxy_cache_valid 404 1m;` to cache 404 responses for 1 minute
* `proxy_cache_use_stale` tells nginx it's allowed to use a stale cache entry (so older than 10 minutes) in certain cases
* `proxy_cache_lock on` means that if a resource is not cached and there's multiple concurrent requests for them, the queries will be queued up so that only one request goes through and the rest is then answered from cache
* `add_header X-Cache-Status $upstream_cache_status` will add an `X-Cache-Status` header to the response so you can check if things are getting cached. You can remove this.

Tweaking `proxy_cache_use_stale` is how you can ensure webfinger responses are still answered even if GTS itself is down. The provided configuration will serve a stale response in case there's an error proxying to GTS, if our connection to GTS times out, if GTS returns a 5xx status code or if GTS returns 429 (Too Many Requests). The `updating` value says that we're allowed to serve a stale entry if nginx is currently in the process of refreshing its cache. Because we configured `inactive=1w` in the `proxy_cache_path` directive, nginx may serve a response up to one week old if the conditions in `proxy_cache_use_stale` are met.

### Serving static assets

By default, GTS will serve assets like the CSS and fonts for the web UI as well as attachments for statuses. However it's very simple to have nginx do this instead and offload GTS from that responsibility. Nginx can generally do a faster job at this too since it's able to use newer functionality in the OS that the Go runtime hasn't necessarily adopted yet.

There are 2 paths that nginx can handle for us:
* `/assets` which contains fonts, CSS, images etc. for the web UI
* `/fileserver` which serves attachments for status posts when using the local storage backend

For `/assets` we'll need the value of `web-asset-base-dir` from the configuration, and for `/fileserver` we'll want `storage-local-base-path`. You can then adjust your nginx configuration like this:

```nginx.conf
server {
  server_name example.org;
  location /assets/ {
    alias web-asset-base-dir/;
    autoindex off;
    expires 5m;
    add_header Cache-Control "public";
  }

  location /fileserver/ {
    alias storage-local-base-path/;
    autoindex off;
    expires max;
    add_header Cache-Control "public, immutable";
  }

  location / {
    proxy_pass http://localhost:8080/;
    proxy_set_header Host $host;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header X-Forwarded-Proto $scheme;
  }
  client_max_body_size 40M;

  listen [::]:443 ssl ipv6only=on; # managed by Certbot
  listen 443 ssl; # managed by Certbot
  ssl_certificate /etc/letsencrypt/live/example.org/fullchain.pem; # managed by Certbot
  ssl_certificate_key /etc/letsencrypt/live/example.org/privkey.pem; # managed by Certbot
  include /etc/letsencrypt/options-ssl-nginx.conf; # managed by Certbot
  ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem; # managed by Certbot
}
```

The trailing slashes in the new `location` directives and the `alias` are significant, do not remove those. The `expires` directive adds the necessary headers to inform the client how long it may cache the resource. For assets, which may change on each release, 5 minutes is used in this example. For attachments, which should never change once they're created, `max` is used instead setting the cache expiry to the 31st of December 2037. For other options, see the nginx documentation on the [`expires` directive](https://nginx.org/en/docs/http/ngx_http_headers_module.html#expires). Nginx does not add cache headers to 4xx or 5xx response codes so a failure to fetch an asset won't get cached by clients. The `autoindex off` directive tells nginx to not serve a directory listing. This should be the default but it doesn't hurt to be explicit. The added `add_header` lines set additional options for the `Cache-Control` header:
* `public` is used to indicate that anyone may cache this resource
* `immutable` is used to indicate this resource will never change while it is fresh (it's before the end of the expires) allowing clients to forego conditional requests to revalidate the resource during that timespan
