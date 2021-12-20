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
  server_name example.com;
  location / {
    proxy_pass http://localhost:8080;
    proxy_set_header Host $host;
  }
}
```

Change `proxy_pass` to the ip and port that you're actually serving GoToSocial on and change `server_name` to your own domain name.
If your domain name is `gotosocial.example.com` then `server_name gotosocial.example.com;` would be the correct value.
If you're running GoToSocial on another machine with the local ip of 192.168.178.69 and on port 8080 then `proxy_pass http://192.168.178.69:8080` would be the correct value.

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

```
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
Just reload it one last time and after that you should be good to go!

```bash
sudo systemctl restart nginx
```
