# Caddy 2

## Requirements

For this guide you will need [Caddy 2](https://caddyserver.com/), there are no other dependencies. Caddy manages Lets Encrypt certificates and renewal for them.

Caddy is in the most popular package managers, or you can get a static binary. For all latest installation guides, refer to [their manual](https://caddyserver.com/docs/install).

### Debian, Ubuntu, Raspbian

```bash
# Add the keyring for their custom repository.
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list

# Update packages and install it
sudo apt update
sudo apt install caddy
```

### Fedora, Redhat, Centos

```bash
dnf install 'dnf-command(copr)'
dnf copr enable @caddy/caddy
dnf install caddy
```

### Arch

```bash
pacman -Syu caddy
```

### FreeBSD
```bash
sudo pkg install caddy
```

## Configure GoToSocial

If GoToSocial is already running, stop it.

```bash
sudo systemctl stop gotosocial
```
In your GoToSocial config turn off Lets Encrypt by setting `letsencrypt-enabled` to `false`.

If you are running GoToSocial on port 443, change the `port` value back to the default `8080`.

If the reverse proxy will be running on the same machine, set the `bind-address` to `"localhost"` so that the GoToSocial server is only accessible via loopback. Otherwise it may be possible to bypass your proxy by connecting to GoToSocial directly, which might be undesirable.

## Set up Caddy

We will configure Caddy 2 to use GoToSocial on our main domain example.org. Since Caddy takes care of obtaining the Lets Encrypt certificate, we only need to configure it properly once.

In most simple use cases Caddy defaults to a file called Caddyfile. It can reload on changes, or can be configured through an HTTP API for zero downtime, but this is out of our current scope.

```bash
sudo mkdir -p /etc/caddy
sudo vim /etc/caddy/Caddyfile
```

While editing the file above, you should replace 'example.org' with your domain. Your domain should occur twice in the current configuration. If you have chosen another port number for GoToSocial other than port 8080, change the port number on the reverse proxy line to match that.

The file you're about to create should look like this:

```Caddyfile
example.org {
	# Optional, but recommended, compress the traffic using proper protocols
	encode zstd gzip

	# The actual proxy configuration to port 8080 (unless you've chosen another port number)
	reverse_proxy * http://127.0.0.1:8080 {
		# Flush immediately, to prevent buffered response to the client
		flush_interval -1
	}
}
```

By default, caddy sets `X-Forwarded-For` in forwarded requests. To make this and rate limiting work, set the `trusted-proxies` configuration variable. See the [rate limiting](../../api/ratelimiting.md) and [general configuration](../../configuration/general.md) docs

For advanced configuration check the [reverse_proxy directive](https://caddyserver.com/docs/caddyfile/directives/reverse_proxy) at the Caddy documentation.

Now check for configuration errors.

```bash
sudo caddy validate
```

If everything is fine, you should get some info lines as output. Unless there are lines marked with *[err]* in front of them, you are all set.

Everything working? Great! Then restart caddy to load your new config file.

```bash
sudo systemctl restart caddy
```

If everything went right, you're now all set to enjoy your GoToSocial instance, so we are going to start it again.

```bash
sudo systemctl start gotosocial
```

## Results

You should now be able to open the splash page for your instance in your web browser, and will see that it runs under HTTPS!
