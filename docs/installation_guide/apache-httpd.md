# Reverse proxy with Apache httpd

## Requirements

For this you will need Apache httpd server.
That is a fairly popular package so your distro will probably have it.

### Ubuntu

```bash
sudo apt install apache2
```

### Arch

```bash
sudo pacman -S apache
```

### OpenSuse

```bash
sudo zypper install apache2
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

## Set up Apache httpd

First we will set up Apache httpd to serve GoToSocial as unsecured http and then later use certbot to automatically upgrade to https.
Please do not try to use it until that's done or you'll be transmitting passwords over clear text.

First we'll write a configuration for Apache httpd and put it in `/etc/apache2/sites-available`.

```bash
sudo mkdir /etc/apache2/sites-available/
sudoedit /etc/apache2/sites-available/yourgotosocial.url.conf
```

The file you're about to create should look a bit like this:

```apache
<VirtualHost *:80>
  ServerName example.com
  ProxyPreserveHost On
  ProxyPass / http://localhost:8080/
  ProxyPassReverse / http://localhost:8080/
</VirtualHost>
```

**Note***: `ProxyPreserveHost On` is essential: It guarantees that the proxy and the gotosocial speak of the same Server name. If not, gotosocial will build the wrong authentication headers, and all attempts at federation will be rejected with 401.

Change `ProxyPass` to the ip and port that you're actually serving GoToSocial on and change `ServerName` to your own domain name.
If your domain name is `gotosocial.example.com` then `ServerName gotosocial.example.com;` would be the correct value.
If you're running GoToSocial on another machine with the local ip of 192.168.178.69 and on port 8080 then `ProxyPass / http://192.168.178.69:8080/` would be the correct value.

Next we'll need to link the file we just created to the folder that Apache httpd reads configurations for active sites from.

```bash
sudo mkdir /etc/apache2/sites-enabled
sudo ln -s /etc/apache2/sites-available/yourgotosocial.url.conf /etc/apache2/sites-enabled/
```

Now check for configuration errors.

```bash
sudo apachectl -t
```

If everything is fine you should get this as output:

```text
Syntax OK
```

Everything working? Great! Then restart Apache httpd to load your new config file.

```bash
sudo systemctl restart apache2
```

## Setting up SSL with mod_md

To setup Apache httpd with mod_md, we'll need to load the module and then modify our vhost files:

```apache
MDomain example auto

<VirtualHost *:80>
  ServerName example.com
</VirtualHost>

<VirtualHost *:443>
  SSLEngine On
  ServerName example.com
  ProxyPreserveHost On
  ProxyPass / http://localhost:8080/
  ProxyPassReverse / http://localhost:8080/
  RequestHeader set "X-Forwarded-Proto" expr=https
</VirtualHost>
```

This allows mod_md to take care of the SSL setup, and it will also redirect from http to https.
After we put ths into place, we'll need  to restart Apache httpd

```bash
sudo systemctl restart apache2
```

and monitor the logs to see when the new certificate arrives, and then reload it one last time and after that you should be good to go!

Apache httpd needs to be restart (or reloaded), every time mod_md gets a new certificate, see the module's docs for [more information](https://github.com/icing/mod_md#how-to-manage-server-reloads) on that.
