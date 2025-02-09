# NGINX

要使用 NGINX 作为 GoToSocial 的反向代理，你需要在服务器上安装它。如果你打算让 NGINX 处理 TLS，你还需要[配置 TLS 证书](../../advanced/certificates.md)。

!!! tip "提示"
    通过在 `server` 块中包含 `http2 on;` 来启用 NGINX 的 HTTP/2。这样可以加快客户端的体验。请参阅 [ngx_http_v2_module 文档](https://nginx.org/en/docs/http/ngx_http_v2_module.html#example)。

NGINX 已为[多个发行版打包](https://repology.org/project/nginx/versions)。你很可能可以使用发行版的包管理器来安装它。你也可以使用 Docker Hub 上发布的[官方 NGINX 镜像](https://hub.docker.com/_/nginx)通过容器运行 NGINX。

在本指南中，我们还将展示如何使用 certbot 配置 TLS 证书。它也在[许多发行版中打包](https://repology.org/project/certbot/versions)，但许多发行版往往提供的 certbot 版本较旧。如果遇到问题，可以考虑使用[容器镜像](https://hub.docker.com/r/certbot/certbot)。

## 配置 GoToSocial

如果 GoToSocial 已在运行，先停止它。

```bash
sudo systemctl stop gotosocial
```

或者如果你没有 systemd 服务，只需手动停止它。

这样调整你的 GoToSocial 配置：

```yaml
letsencrypt-enabled: false
port: 8080
bind-address: 127.0.0.1
```

第一个设置禁用了内置的 TLS 证书配置。由于 NGINX 现在将处理这些流量，GoToSocial 不再需要绑定到 443 端口或任何特权端口。

通过将 `bind-address` 设置为 `127.0.0.1`，GoToSocial 将不再能直接从外部访问。如果你的 NGINX 和 GoToSocial 实例不在同一台服务器上，你需要绑定一个允许你的反向代理访问你的 GoToSocial 实例的 IP 地址。绑定到私有 IP 地址可以确保只有通过 NGINX 才能访问 GoToSocial。

## 设置 NGINX

我们首先设置 NGINX 为 GoToSocial 提供不安全的 http 服务，然后使用 Certbot 自动升级为 https 服务。

请勿在此完成之前尝试使用，否则你将有泄露密码的风险，或破坏联合。

首先，我们将为 NGINX 编写一个配置文件，并将其放入 `/etc/nginx/sites-available` 中。

```bash
sudo mkdir -p /etc/nginx/sites-available
sudoedit /etc/nginx/sites-available/yourgotosocial.url.conf
```

在上述命令中，将 `yourgotosocial.url` 替换为你的实际 GoToSocial 主机值。所以如果你的 `host` 设置为 `example.org`，那么文件应该命名为 `/etc/nginx/sites-available/example.org.conf`

你要创建的文件应该如下所示：

```nginx
server {
  listen 80;
  listen [::]:80;
  server_name example.org;
  location / {
    # 设置为 127.0.0.1 而不是 localhost 以解决 https://stackoverflow.com/a/52550758
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

将 `proxy_pass` 改为你实际运行 GoToSocial 的 IP 和端口（如果不是 `127.0.0.1:8080`），并将 `server_name` 改为你自己的域名。

如果你的域名是 `example.org`，那么 `server_name example.org;` 就是正确的值。

如果你在另一台本地 IP 为 192.168.178.69 的机器上运行 GoToSocial，并在端口 8080 上，那么 `proxy_pass http://192.168.178.69:8080;` 就是正确的值。

**注意**：如果你的服务器不支持 IPv6，可以删除 `listen [::]:80;` 这一行。

**注意**：`proxy_set_header Host $host;` 必不可少。它确保代理和 GoToSocial 使用相同的服务器名称。如果没有，GoToSocial 将构建错误的身份验证标头，导致所有的联合尝试以 401 被拒绝。

**注意**：`Connection` 和 `Upgrade` 头用于 WebSocket 连接。请参阅 [WebSocket 文档](websocket.md)。

**注意**：本例中 `client_max_body_size` 设置为 40M，这是 GoToSocial 的默认视频上传大小。根据需要你可以将此值设置得更大或更小。nginx 的默认值仅为 1M，太小了。

**注意**：为了使 `X-Forwarded-For` 和限流生效，请设置 `trusted-proxies` 配置变量。请参阅[限流](../../api/ratelimiting.md)和[通用配置](../../configuration/general.md)文档。

接下来我们需要将刚创建的文件链接到 nginx 从中读取活动站点配置的文件夹中。

```bash
sudo mkdir -p /etc/nginx/sites-enabled
sudo ln -s /etc/nginx/sites-available/yourgotosocial.url.conf /etc/nginx/sites-enabled/
```

再次将 `yourgotosocial.url` 替换为你的实际 GoToSocial 主机值。

现在检查配置错误。

```bash
sudo nginx -t
```

如果一切正常，你应该会看到以下输出：

```text
nginx: the configuration file /etc/nginx/nginx.conf syntax is ok
nginx: configuration file /etc/nginx/nginx.conf test is successful
```

一切正常吗？太好了！然后重启 nginx 以加载新的配置文件。

```bash
sudo systemctl restart nginx
```

## 设置 TLS

!!! warning "警告"
    我们有关于如何[配置 TLS 证书](../../advanced/certificates.md)的附加文档，还提供了有关不同发行版的附加内容和教程链接，值得一看。

你现在可以运行 certbot，它将引导你完成启用 https 的步骤。

```bash
sudo certbot --nginx
```

完成后，它应该自动编辑你的配置文件以启用 https。

最后再次重新加载 NGINX：

```bash
sudo systemctl restart nginx
```

现在重新启动 GoToSocial：

```bash
sudo systemctl start gotosocial
```

## 安全加固

如果你想通过进阶配置选项加强 NGINX 部署，网上有很多指南（[例如这个](https://beaglesecurity.com/blog/article/nginx-server-security.html)）。请尝试找到最新的指南。Mozilla 还[在此处](https://ssl-config.mozilla.org/)发布了最佳实践 SSL 配置。

## 结果

你现在应该可以在浏览器中打开实例的启动页面，并看到它运行在 https 下！

如果你再次打开 NGINX 配置，你会发现 Certbot 添加了一些额外的行。

!!! warning "警告"
    根据你设置 Certbot 时选择的选项，以及使用的 NGINX 版本，可能会有所不同。

```nginx
server {
  server_name example.org;
  location / {
    # 设置为 127.0.0.1 而不是 localhost 以解决 https://stackoverflow.com/a/52550758
    proxy_pass http://127.0.0.1:8080;
    proxy_set_header Host $host;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header X-Forwarded-Proto $scheme;
  }
  client_max_body_size 40M;

  listen [::]:443 ssl; # 由 Certbot 管理
  listen 443 ssl; # 由 Certbot 管理
  http2 on; # 由 Certbot 管理
  ssl_certificate /etc/letsencrypt/live/example.org/fullchain.pem; # 由 Certbot 管理
  ssl_certificate_key /etc/letsencrypt/live/example.org/privkey.pem; # 由 Certbot 管理
  include /etc/letsencrypt/options-ssl-nginx.conf; # 由 Certbot 管理
  ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem; # 由 Certbot 管理
}

server {
  if ($host = example.org) {
      return 301 https://$host$request_uri;
  } # 由 Certbot 管理

  listen 80;
  listen [::]:80;
  server_name example.org;
    return 404; # 由 Certbot 管理
}
```

关于 nginx 的其他配置选项（包括静态资源服务和缓存），请参阅文档的[进阶配置部分](../../advanced/index.md)。
