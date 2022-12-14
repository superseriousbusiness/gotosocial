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
    # set to 127.0.0.1 instead of localhost to work around https://stackoverflow.com/a/52550758
    proxy_pass http://127.0.0.1:8080;
    proxy_set_header Host $host;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header X-Forwarded-Proto $scheme;
  }
  client_max_body_size 40M;
}
```

Change `proxy_pass` to the ip and port that you're actually serving GoToSocial on (if it's not on `127.0.0.1:8080`), and change `server_name` to your own domain name.

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
    # set to 127.0.0.1 instead of localhost to work around https://stackoverflow.com/a/52550758
    proxy_pass http://127.0.0.1:8080/;
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

A number of additional configurations for nginx, including static asset serving and caching, are documented in the [Advanced](advanced.md) section of our documentation.
