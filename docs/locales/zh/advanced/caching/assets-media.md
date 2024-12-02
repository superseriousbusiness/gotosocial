# 缓存资源与媒体

当你配置 GoToSocial 实例使用本地存储媒体时，可以使用你的[反向代理](../../getting_started/reverse_proxy/index.md)直接提供这些文件并进行缓存。这样可以避免频繁请求 GoToSocial，同时反向代理通常能比 GoToSocial 更快地提供资源。

你还可以使用反向代理来缓存 GoToSocial Web UI 的资源，比如其使用的 CSS 和图片。

当使用[分域](../host-account-domain.md)部署方式时，你需要确保在主机域上配置资源和媒体的缓存。

!!! warning "媒体修剪"
    如果你配置了媒体修剪，必须确保当磁盘上找不到媒体时，仍然将请求发送到 GoToSocial。这将保证从外站实例重新获取该媒体，之后的请求将再次由你的反向代理处理。

## 端点

有两个端点提供可服务和缓存的资源：

* `/assets` 包含字体、CSS、图像等 Web UI 的资源
* `/fileserver` 在使用本地存储后端时，服务于贴文的附件

`/assets` 的文件系统位置由 [`web-asset-base-dir`](../../configuration/web.md) 配置选项定义。`/fileserver` 下的文件从 [`storage-local-base-path`](../../configuration/storage.md) 获取。

## 配置

=== "apache2"

	`Cache-Control` 头手动设置，合并配置和 `expires` 指令的值，以避免因为两个头行而导致错误。默认情况下 `Header set` 为 `onsuccess`，因此它也不会添加到错误响应中。

	假设你的 GtS 安装在 `/opt/GtS` 根目录下，并有一个 `storage` 子目录，且 Web 服务器已被授予访问权限，可以在 vhost 中添加以下部分：

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

	这里的技巧是在基于 Apache 2 的反向代理设置中…

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

	… 默认情况下所有的请求都是通过代理的，`RewriteRule` 通过指定文件系统路径来绕过代理以重定向到特定 URL 前缀，而 `RewriteCond` 确保只有在文件确实存在时才禁用 `/fileserver/` 代理。

	你还需要运行以下命令（假设使用类似 Debian 的设置）来启用使用的模块：

	```
	$ sudo a2enmod expires
	$ sudo a2enmod headers
	$ sudo a2enmod rewrite
	```

	然后（在测试配置后）重启 Apache。

=== "nginx"

	以下是你需要在现有的 nginx 配置中添加的三个位置块的示例：

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

	`/fileserver` 位置有点特殊。当我们无法从磁盘获取媒体时，我们希望将请求代理到 GoToSocial，以便它尝试获取。`try_files` 指令本身不能使用 `proxy_pass`，所以我们创建了命名的 `@fileserver` 位置，在 `try_files` 中最后传递给它。

	!!! bug "尾部斜杠"
		`location` 指令和 `alias` 中的尾部斜杠很重要，不要移除它们。

	`expires` 指令添加了必要的头信息，以告知客户端可以缓存资源的时间：

	* 对于资源，因为可能在每次发布时更改，所以在此示例中使用了 5 分钟
	* 对于附件，因为一旦创建后永远不会更改，所以当前使用一周

	有关其他选项，请参阅 [nginx 的 `expires` 指令](https://nginx.org/en/docs/http/ngx_http_headers_module.html#expires)文档。

	Nginx 不会为 4xx 或 5xx 响应代码添加缓存头，因此抓取资源失败时不会被客户端缓存。`autoindex off` 指令告诉 nginx 不提供目录列表。这应该是默认设置，但明确设置不会有害。添加的 `add_header` 行为 `Cache-Control` 头设置了额外的选项：

	* `public` 用于指示任何人都可以缓存此资源
	* `immutable` 用于指示该资源在其新鲜期内（在 `expires` 之前）绝不会更改，允许客户端在此期间忽略条件请求以重新验证资源。
