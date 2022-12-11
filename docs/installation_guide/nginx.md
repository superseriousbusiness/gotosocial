# Reverse proxy with NGINX

## Requirements

For this you will need [Certbot](https://certbot.eff.org/), the Certbot NGINX plugin and of course [NGINX](https://www.nginx.com/) itself.

These are popular packages so your distro will probably have them.

### Ubuntu

```bash
sudo apt install certbot python3-certbot-nginx nginx
```

### Arch

```bash
sudo pacman -S certbot certbot-nginx nginx
```

### OpenSuse

```bash
sudo zypper install nginx python3-certbot python3-certbot-nginx
```

## Configure GoToSocial

If GoToSocial is already running, stop it.

```bash
sudo systemctl stop gotosocial
```

Or if you don't have a systemd service just stop it manually.

In your GoToSocial config turn off letsencrypt by setting `letsencrypt-enabled` to `false`.

If you we running GoToSocial on port 443, change the `port` value back to the default `8080`.

If the reverse proxy will be running on the same machine, set the `bind-address` to `"localhost"` so that the GoToSocial server is only accessible via loopback. Otherwise it may be possible to bypass your proxy by connecting to GoToSocial directly, which might be undesirable.

## Set up NGINX

First we will set up NGINX to serve GoToSocial as unsecured http and then use Certbot to automatically upgrade it to serve https.

Please do not try to use it until that's done or you'll risk transmitting passwords over clear text, or breaking federation.

First we'll write a configuration for NGINX and put it in `/etc/nginx/sites-available`.

```bash
sudo mkdir -p /etc/nginx/sites-available
sudoedit /etc/nginx/sites-available/yourgotosocial.url.conf
```

In the above commands, replace `yourgotosocial.url` with your actual GoToSocial host value. So if your `host` is set to `example.org`, then the file should be called `/etc/nginx/sites-available/example.org.conf`

The file you're about to create should look like this:

```nginx.conf
server {
  listen 80;
  listen [::]:80;
  server_name example.org;
  location / {
    proxy_pass http://localhost:8080;
    proxy_set_header Host $host;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header X-Forwarded-Proto $scheme;
  }
  client_max_body_size 40M;
}
```

Change `proxy_pass` to the ip and port that you're actually serving GoToSocial on and change `server_name` to your own domain name.

If your domain name is `example.org` then `server_name example.org;` would be the correct value.

If you're running GoToSocial on another machine with the local ip of 192.168.178.69 and on port 8080 then `proxy_pass http://192.168.178.69:8080;` would be the correct value.

**Note**: You can remove the line `listen [::]:80;` if your server is not ipv6 capable.

**Note**: `proxy_set_header Host $host;` is essential. It guarantees that the proxy and GoToSocial use the same server name. If not, GoToSocial will build the wrong authentication headers, and all attempts at federation will be rejected with 401.

**Note**: The `Connection` and `Upgrade` headers are used for WebSocket connections. See the [WebSocket docs](./websocket.md).

**Note**: `client_max_body_size` is set to 40M in this example, which is the default max video upload size for GoToSocial. You can make this value larger or smaller if necessary. The nginx default is only 1M, which is rather too small.

**Note**: To make `X-Forwarded-For` and rate limiting work, set the `trusted-proxies` configuration variable. See the [rate limiting](../api/ratelimiting.md) and [general configuration](../configuration/general.md) docs

Next we'll need to link the file we just created to the folder that nginx reads configurations for active sites from.

```bash
sudo mkdir -p /etc/nginx/sites-enabled
sudo ln -s /etc/nginx/sites-available/yourgotosocial.url.conf /etc/nginx/sites-enabled/
```

Again, replace `yourgotosocial.url` with your actual GoToSocial host value.

Now check for configuration errors.

```bash
sudo nginx -t
```

If everything is fine you should get this as output:

```text
nginx: the configuration file /etc/nginx/nginx.conf syntax is ok
nginx: configuration file /etc/nginx/nginx.conf test is successful
```

Everything working? Great! Then restart nginx to load your new config file.

```bash
sudo systemctl restart nginx
```

## Setting up SSL with certbot

You should now be able to run certbot and it will guide you through the steps required to enable https for your instance.

```bash
sudo certbot --nginx
```

After you do, it should have automatically edited your configuration file to enable https.

Reload NGINX one last time:

```bash
sudo systemctl restart nginx
```

Now start GoToSocial again:

```bash
sudo systemctl start gotosocial
```

## Results

You should now be able to open the splash page for your instance in your web browser, and will see that it runs under https!

If you open the NGINX config again, you'll see that Certbot added some extra lines to it.

**Note**: This may look a bit different depending on the options you chose while setting up Certbot, and the NGINX version you're using.

```nginx.conf
server {
  server_name example.org;
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

server {
  if ($host = example.org) {
      return 301 https://$host$request_uri;
  } # managed by Certbot

  listen 80;
  listen [::]:80;
  server_name example.org;
    return 404; # managed by Certbot
}
```

## Extra Hardening

If you want to harden up your NGINX deployment with advanced configuration options, there are many guides online for doing so ([for example](https://beaglesecurity.com/blog/article/nginx-server-security.html)). Try to find one that's up to date. Mozilla also publishes best-practice ssl configuration [here](https://ssl-config.mozilla.org/).

## Caching Webfinger

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

