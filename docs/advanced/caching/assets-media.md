# Caching assets and media

When you've configured your GoToSocial instance with local storage for media, you can use your [reverse proxy](../../getting_started/reverse_proxy/index.md) to serve these files directly and cache them. This avoids hitting GoToSocial for these requests and reverse proxies can typically serve assets faster than GoToSocial.

You can also use your reverse proxy to cache the GoToSocial web UI assets, like the CSS and images it uses.

When using a [split domain](../host-account-domain.md) deployment style, you need to ensure you configure caching of the assets and media on the host domain.

!!! warning "Media pruning"
    If you've configured media pruning, you need to ensure that when media is not found on disk the request is still sent on to GoToSocial. This will ensure the media is fetched again from the remote instance and subsequent requests for this media will then be handled by your reverse proxy again.

## Endpoints

There are 2 endpoints that serve assets we can serve and cache:

* `/assets` which contains fonts, CSS, images etc. for the web UI
* `/fileserver` which serves attachments for status posts when using the local storage backend

The filesystem location of `/assets` is defined by the [`web-asset-base-dir`](../../configuration/web.md) configuration option. Files under `/fileserver` are retrieved from the [`storage-local-base-path`](../../configuration/storage.md).

## Configuration

=== "apache2"

	The `Cache-Control` header is manually set to merge the values
	from the configuration and the `expires` directive to avoid
	breakage from having two header lines. `Header set` defaults
	to ` onsuccess`, so it is also not added to error responses.

	Assuming your GtS installation is rooted in `/opt/GtS` with a
	`storage` subdirectory, and the webserver has been given access,
	add the following section to the vhost:

	```apacheconf
	<Directory /opt/GtS/web/assets>
		Options None
		AllowOverride None
		Require all granted
		ExpiresActive on
		ExpiresDefault A300
		Header set Cache-Control "public, max-age=300"
	</Directory>
	RewriteRule "^/assets/(.*)$" "/opt/GtS/web/assets/$1" [L]

	<Directory /opt/GtS/storage>
		Options None
		AllowOverride None
		Require all granted
		ExpiresActive on
		ExpiresDefault A604800
		Header set Cache-Control "private, immutable, max-age=604800"
	</Directory>
	RewriteCond "/opt/GtS/storage/$1" -f
	RewriteRule "^/fileserver/(.*)$" "/opt/GtS/storage/$1" [L]
	```

	The trick here is that, in an Apache 2-based reverse proxy setup…

	```apacheconf
	RewriteEngine On

	RewriteCond %{HTTP:Upgrade} websocket [NC]
	RewriteCond %{HTTP:Connection} upgrade [NC]
	RewriteRule ^/?(.*) "ws://localhost:8980/$1" [P,L]

	ProxyIOBufferSize 65536
	ProxyTimeout 120

	ProxyPreserveHost On
	<Location "/">
		ProxyPass http://127.0.0.1:8980/
		ProxyPassReverse http://127.0.0.1:8980/
	</Location>
	```

	… everything is proxied by default, the `RewriteRule` bypasses
	the proxy (by specifying a filesystem path to redirect to) for
	specific URL præficēs and the `RewriteCond` ensures to only
	disable the `/fileserver/` proxy if the file is, indeed, present.

	Also run the following commands (assuming a Debian-like setup)
	to enable the modules used:

	```
	$ sudo a2enmod expires
	$ sudo a2enmod headers
	$ sudo a2enmod rewrite
	```

	Then (after a configtest), restart Apache.

=== "nginx"

	Here's an example of the three location blocks you'll need to add to your existing configuration in nginx:

	```nginx
	server {
	server_name social.example.org;

	location /assets/ {
		alias web-asset-base-dir/;
		autoindex off;
		expires 5m;
		add_header Cache-Control "public";
	}

	location @fileserver {
		proxy_pass http://localhost:8080;
		proxy_set_header Host $host;
		proxy_set_header Upgrade $http_upgrade;
		proxy_set_header Connection "upgrade";
		proxy_set_header X-Forwarded-For $remote_addr;
		proxy_set_header X-Forwarded-Proto $scheme;
	}

	location /fileserver/ {
		alias storage-local-base-path/;
		autoindex off;
		expires 1w;
		add_header Cache-Control "private, immutable";
		try_files $uri @fileserver;
	}
	}
	```

	The `/fileserver` location is a bit special. When we fail to fetch the media from disk, we want to proxy the request on to GoToSocial so it can try and fetch it. The `try_files` directive can't take a `proxy_pass` itself so instead we created the named `@fileserver` location that we pass in last to `try_files`.

	!!! bug "Trailing slashes"
		The trailing slashes in the `location` directives and the `alias` are significant, do not remove those.

	The `expires` directive adds the necessary headers to inform the client how long it may cache the resource:

	* For assets, which may change on each release, 5 minutes is used in this example
	* For attachments, which should never change once they're created, we currently use one week

	For other options, see the nginx documentation on the [`expires` directive](https://nginx.org/en/docs/http/ngx_http_headers_module.html#expires).

	Nginx does not add cache headers to 4xx or 5xx response codes so a failure to fetch an asset won't get cached by clients. The `autoindex off` directive tells nginx to not serve a directory listing. This should be the default but it doesn't hurt to be explicit. The added `add_header` lines set additional options for the `Cache-Control` header:

	* `public` is used to indicate that anyone may cache this resource
	* `immutable` is used to indicate this resource will never change while it is fresh (it's before the end of the expires) allowing clients to forgo conditional requests to revalidate the resource during that timespan.
