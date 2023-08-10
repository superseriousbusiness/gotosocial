# Apache HTTP Server

In order to use Apache as a reverse proxy for GoToSocial you'll need to have it installed on your server. If you intend for the Apache instance to also handle TLS, you'll need to [provision TLS certificates](../../advanced/certificates.md) too.

Apache is [packaged for many distributions](https://repology.org/project/apache/versions). It's very likely you can install it with your distribution's package manager. You can also run Apache using a container runtime with the [official Apache image](https://hub.docker.com/_/httpd) that's published to the Docker Hub.

In this guide we'll also show how to use certbot to provision the TLS certificates. It too is [packaged in many distributions](https://repology.org/project/certbot/versions) but many distributions tend to ship fairly old versions of certbot. If you run into trouble it may be worth considering using the [container image](https://hub.docker.com/r/certbot/certbot) instead.

## Configure GoToSocial

We're going to have Apache handle LetsEncrypt certificates, so you need to turn off built-in LetsEncrypt support in your GoToSocial config.

First open the file in your text editor:

```bash
sudoedit /gotosocial/config.yaml
```

Then set `letsencrypt-enabled: false`.

If the reverse proxy will be running on the same machine, set the `bind-address` to `"localhost"` so that the GoToSocial server is only accessible via loopback. Otherwise it may be possible to bypass your proxy by connecting to GoToSocial directly, which might be undesirable.

If GoToSocial is already running, restart it.

```bash
sudo systemctl restart gotosocial.service
```

Or if you don't have a systemd service just restart it manually.

## Set up Apache

### Required Apache modules

You need to ensure you have a number of Apache modules installed and enabled. All these modules *should* ship with your distribution's Apache package, but they may have been split out into separate packages.

You can check which modules you have installed with `apachectl -M`.

You'll need to have the following modules loaded:

* `proxy_http` to proxy HTTP requests to GoToSocial
* `ssl` to handle SSL/TLS
* `headers` to manipulate HTTP request and response headers
* `rewrite` to rewrite HTTP requests
* `md` for Lets Encrypt, available since 2.4.30

On Debian, Ubuntu and openSUSE, you can use the [`a2enmod`](https://manpages.debian.org/bookworm/apache2/a2enmod.8.en.html) utility to load any additional modules. For the Red Hat/CentOS family of distributions, you'll need to add a [`LoadModule` directive](https://httpd.apache.org/docs/2.4/mod/mod_so.html#loadmodule) to your Apache configuration instead.

### TLS with mod_md

!!! note
    `mod_md` is only available since Apache 2.4.30 and still considered experimental. It works well enough in practice and is the most convenient method.

Now we'll configure Apache HTTP Server to serve GoToSocial requests.

First we'll write a configuration for Apache HTTP Server and put it in `/etc/apache2/sites-available`:

```bash
sudo mkdir -p /etc/apache2/sites-available/
sudoedit /etc/apache2/sites-available/example.com.conf
```

In the above `sudoedit` command, replace `example.com` with the hostname of your GoToSocial server.

The file you're about to create should look a bit like this:

```apache
MDomain example.com auto
MDCertificateAgreement accepted

<VirtualHost *:80 >
  ServerName example.com
</VirtualHost>

<VirtualHost *:443>
  ServerName example.com

  RewriteEngine On
  RewriteCond %{HTTP:Upgrade} websocket [NC]
  RewriteCond %{HTTP:Connection} upgrade [NC]
  # set to 127.0.0.1 instead of localhost to work around https://stackoverflow.com/a/52550758
  RewriteRule ^/?(.*) "ws://127.0.0.1:8080/$1" [P,L]

  SSLEngine On
  ProxyPreserveHost On
  # set to 127.0.0.1 instead of localhost to work around https://stackoverflow.com/a/52550758
  ProxyPass / http://127.0.0.1:8080/
  ProxyPassReverse / http://127.0.0.1:8080/

  RequestHeader set "X-Forwarded-Proto" expr=https
</VirtualHost>
```

or, if you have [Apache httpd 2.4.47+](https://httpd.apache.org/docs/2.4/mod/mod_proxy.html#protoupgrade), you can get rid of both `mod_rewrite` and `mod_proxy_wstunnel` and simplify the whole config to:

```apache
MDomain example.com auto
MDCertificateAgreement accepted

<VirtualHost *:80 >
  ServerName example.com
</VirtualHost>

<VirtualHost *:443>
  ServerName example.com

  SSLEngine On
  ProxyPreserveHost On
  # set to 127.0.0.1 instead of localhost to work around https://stackoverflow.com/a/52550758
  ProxyPass / http://127.0.0.1:8080/ upgrade=websocket
  ProxyPassReverse / http://127.0.0.1:8080/

  RequestHeader set "X-Forwarded-Proto" expr=https
</VirtualHost>
```

Again, replace occurrences of `example.com` in the above config file with the hostname of your GtS server. If your domain name is `gotosocial.example.com`, then `gotosocial.example.com` would be the correct value.

You should also change `http://127.0.0.1:8080` to the correct address and port (if it's not on `127.0.0.1:8080`) of your GtS server. For example, if you're running GoToSocial on another machine with the local ip of `192.168.178.69` and on port `8080` then `http://192.168.178.69:8080/` would be the correct value.

`Rewrite*` directives are needed to ensure that Websocket streaming connections also work. See the [websocket](./websocket.md) document for more information on this.

`ProxyPreserveHost On` is essential: It guarantees that the proxy and the GoToSocial speak of the same Server name. If not, GoToSocial will build the wrong authentication headers, and all attempts at federation will be rejected with 401 Unauthorized.

By default, apache sets `X-Forwarded-For` in forwarded requests. To make this and rate limiting work, set the `trusted-proxies` configuration variable. See the [rate limiting](../../api/ratelimiting.md) and [general configuration](../../configuration/general.md) docs

Save and close the config file.

Now we'll need to link the file we just created to the folder that Apache HTTP Server reads configurations for active sites from.

```bash
sudo mkdir /etc/apache2/sites-enabled
sudo ln -s /etc/apache2/sites-available/example.com.conf /etc/apache2/sites-enabled/
```

In the above `ln` command, replace `example.com` with the hostname of your GoToSocial server.

Now check for configuration errors.

```bash
sudo apachectl -t
```

If everything is fine you should get this as output:

```text
Syntax OK
```

Everything working? Great! Then restart Apache HTTP Server to load your new config file.

```bash
sudo systemctl restart apache2
```

Now, monitor the logs to see when the new LetsEncrypt certificate arrives (`tail -F /var/log/apache2/error.log`), and then reload Apache one last time with the above `systemctl restart` command. After that you should be good to go!

Apache HTTP Server needs to be restart (or reloaded), every time `mod_md` gets a new certificate; see the module's docs for [more information](https://github.com/icing/mod_md#how-to-manage-server-reloads).

Depending on your version of Apache HTTP Server, you may see the following error: `error (specific information not available): acme problem urn:ietf:params:acme:error:invalidEmail: Error creating new account :: contact email "webmaster@localhost" has invalid domain : Domain name needs at least one dot`

If this happens, you'll need to do one (or all) of the below:

1. Update `/etc/apache2/sites-enabled/000-default.conf` and change the `ServerAdmin` value to a valid email address (then reload Apache HTTP Server).
2. Add the line `MDContactEmail your.email.address@whatever.com` below the `MDomain` line in `/etc/apache2/sites-available/example.com.conf`, replacing `your.email.address@whatever.com` with a valid email address, and `example.com` with your GtS host name.

### TLS with externally managed certificates

!!! note
    We have additional documentation on how to [provision TLS certificates](../../advanced/certificates.md) that also provides links to additional content and tutorials for different distributions that may be good to review.

If you prefer to have a manual setup or setting SSL using a different service to manage it (Certbot, etc), then you can use a simpler setup for your Apache HTTP Server.

First we'll write a configuration for Apache HTTP Server and put it in `/etc/apache2/sites-available`:

```bash
sudo mkdir -p /etc/apache2/sites-available/
sudoedit /etc/apache2/sites-available/example.com.conf
```

In the above `sudoedit` command, replace `example.com` with the hostname of your GoToSocial server.

The file you're about to create should look initially for both 80 (required) and 443 ports (optional) a bit like this:

```apache
<VirtualHost *:80>
  ServerName example.com

  RewriteEngine On
  RewriteCond %{HTTP:Upgrade} websocket [NC]
  RewriteCond %{HTTP:Connection} upgrade [NC]
  # set to 127.0.0.1 instead of localhost to work around https://stackoverflow.com/a/52550758
  RewriteRule ^/?(.*) "ws://127.0.0.1:8080/$1" [P,L]

  ProxyPreserveHost On
  # set to 127.0.0.1 instead of localhost to work around https://stackoverflow.com/a/52550758
  ProxyPass / http://127.0.0.1:8080/
  ProxyPassReverse / http://127.0.0.1:8080/

</VirtualHost>
```

Again, replace occurrences of `example.com` in the above config file with the hostname of your GtS server. If your domain name is `gotosocial.example.com`, then `gotosocial.example.com` would be the correct value.

You should also change `http://127.0.0.1:8080` to the correct address and port (if it's not on `127.0.0.1:8080`) of your GtS server. For example, if you're running GoToSocial on another machine with the local ip of `192.168.178.69` and on port `8080` then `http://192.168.178.69:8080/` would be the correct value.

`Rewrite*` directives are needed to ensure that Websocket streaming connections also work. See the [websocket](websocket.md) document for more information on this.

`ProxyPreserveHost On` is essential: It guarantees that the proxy and the GoToSocial speak of the same Server name. If not, GoToSocial will build the wrong authentication headers, and all attempts at federation will be rejected with 401 Unauthorized.

In the case of providing an initial setup for the 443 port looking for additional managing by an external tool, you could use default certificates provided by the server which you can find referenced in the `default-ssl.conf` file at `/etc/apache2/sites-available/`.

Save and close the config file.

Now we'll need to link the file we just created to the folder that Apache HTTP Server reads configurations for active sites from.

```bash
sudo mkdir /etc/apache2/sites-enabled
sudo ln -s /etc/apache2/sites-available/example.com.conf /etc/apache2/sites-enabled/
```

In the above `ln` command, replace `example.com` with the hostname of your GoToSocial server.

Now check for configuration errors.

```bash
sudo apachectl -t
```

If everything is fine you should get this as output:

```text
Syntax OK
```

Everything working? Great! Then restart Apache HTTP Server to load your new config file.

```bash
sudo systemctl restart apache2
```

## Troubleshooting

If you cannot connect to the site in your browser, the reverse proxy setup doesn't work. Compare the Apache log file (`tail -F /var/log/apache2/access.log`) with the GoToSocial log file. Requests made must show up in both places. Double check the `ProxyPass` setting.

If you can connect but your posts don't federate and your account cannot be found from elsewhere, check your logs. Federation is broken if you see messages attempting to read your profile (something like `level=INFO … method=GET statusCode=401 path=/users/your_username msg="Unauthorized: …"`) or post to your inbox (something like `level=INFO … method=POST statusCode=404 path=/your_username/inbox msg="Not Found: …"`). Double check the `ProxyPreserveHost` setting.

If you can connect but you cannot authorize your account in a Mastodon client app, make sure you initiate login from the right domain. When using a [split domain](../../advanced/host-account-domain.md) setup, you have to initiate the login from the `host` domain, not the `account-domain`. GoToSocial sets a `Content-Security-Policy` header to counter XSS and data injection attacks. This header should be left untouched, so make sure your reverse proxy doesn't modify, override or unset it.
