# Reverse proxy with nginx

## Requirements

For this you will need certbot, the certbot nginx plugin and of course nginx.
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

In your GoToSocial config turn off letsencrypt.
First open the file in your text editor.

```bash
sudoedit /gotosocial/config.yaml
```

Then set `letsencrypt-enabled: false`.

If GoToSocial is already running, restart it.

```bash
sudo systemctl restart gotosocial.service
```

Or if you don't have a systemd service just restart it manually.

## Set up nginx

First we will set up nginx to serve GoToSocial as unsecured http and then later use certbot to automatically upgrade to https.
Please do not try to use it until that's done or you'll be transmitting passwords over clear text.

First we'll write a configuration for nginx and put it in `/etc/nginx/sites-available`.

```bash
sudo mkdir /etc/nginx/sites-available/
sudoedit /etc/nginx/sites-available/yourgotosocial.url.conf
```

The file you're about to create should look a bit like this:

```nginx.conf
server {
  listen 80;
  listen [::]:80;
  server_name example.com;
  location / {
    proxy_pass http://localhost:8080;
    proxy_set_header Host $host;
  }
}
```

**Note***: You can remove the line `listen [::]:80;` if your server is not ipv6 capable.

**Note***: `proxy_set_header Host $host;` is essential: It guarantees that the proxy and the gotosocial speak of the same Server name. If not, gotosocial will build the wrong authentication headers, and all attempts at federation will be rejected with 401.

Change `proxy_pass` to the ip and port that you're actually serving GoToSocial on and change `server_name` to your own domain name.
If your domain name is `gotosocial.example.com` then `server_name gotosocial.example.com;` would be the correct value.
If you're running GoToSocial on another machine with the local ip of 192.168.178.69 and on port 8080 then `proxy_pass http://192.168.178.69:8080;` would be the correct value.

Next we'll need to link the file we just created to the folder that nginx reads configurations for active sites from.

```bash
sudo mkdir /etc/nginx/sites-enabled
sudo ln -s /etc/nginx/sites-available/yourgotosocial.url.conf /etc/nginx/sites-enabled/
```

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
Reload it one last time and after that you should be good to go!

```bash
sudo systemctl restart nginx
```

### Results

The resulting NGINX config should look something like this:

```nginx.conf
server {
  listen 80;
  listen [::]:80;
  server_name gts.example.com;

  location /.well-known/acme-challenge/ {
    default_type "text/plain";
    root /var/www/certbot;
  }
  location / { return 301 https://$host$request_uri; }
}

server {
  listen 443 ssl http2;
  listen [::]:443 ssl http2;
  server_name gts.example.com;

  #############################################################################
  # Certificates                                                              #
  # you need a certificate to run in production. see https://letsencrypt.org/ #
  #############################################################################
  ssl_certificate     /etc/letsencrypt/live/gts.example.com/fullchain.pem;
  ssl_certificate_key /etc/letsencrypt/live/gts.example.com/privkey.pem;

  location ^~ '/.well-known/acme-challenge' {
    default_type "text/plain";
    root /var/www/certbot;
  }

  ###########################################
  # Security hardening (as of Nov 15, 2020) #
  # based on Mozilla Guideline v5.6         #
  ###########################################

  ssl_protocols             TLSv1.2 TLSv1.3;
  ssl_prefer_server_ciphers on;
  ssl_ciphers "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305";
  ssl_session_timeout       1d; # defaults to 5m
  ssl_session_cache         shared:SSL:10m; # estimated to 40k sessions
  ssl_session_tickets       off;
  ssl_stapling              on;
  ssl_stapling_verify       on;
  ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;
  # HSTS (https://hstspreload.org), requires to be copied in 'location' sections that have add_header directives
  add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload";


  location / {
    proxy_pass         http://127.0.0.1:8080;

    proxy_set_header   Host             $host;
    proxy_set_header   Connection       $http_connection;
    proxy_set_header   X-Real-IP        $remote_addr;
    proxy_set_header   X-Forwarded-For  $proxy_add_x_forwarded_for;
    proxy_set_header   X-Scheme         $scheme;
  }
}
```
