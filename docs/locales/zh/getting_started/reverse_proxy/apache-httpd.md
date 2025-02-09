# Apache HTTP 服务器

要将 Apache 用作 GoToSocial 的反向代理，你需要在服务器上安装它。如果你还希望 Apache 处理 TLS，就需要[配置 TLS 证书](../../advanced/certificates.md)。

Apache 已被[打包用于许多发行版](https://repology.org/project/apache/versions)。你很可能可以使用发行版的包管理器来安装它。你还可以使用发布到 Docker Hub 的[官方 Apache 镜像](https://hub.docker.com/_/httpd)通过容器运行 Apache。

本指南还将展示如何使用 certbot 来配置 TLS 证书。它同样被[打包在许多发行版](https://repology.org/project/certbot/versions)中，但许多发行版往往附带较旧版本的 certbot。如果遇到问题，可以考虑使用[容器镜像](https://hub.docker.com/r/certbot/certbot)。

## 配置 GoToSocial

我们将让 Apache 处理 LetsEncrypt 证书，所以你需要在 GoToSocial 配置中关闭内置的 LetsEncrypt 支持。

首先在文本编辑器中打开文件：

```bash
sudoedit /gotosocial/config.yaml
```

然后设置 `letsencrypt-enabled: false`。

如果反向代理将在同一台机器上运行，请将 `bind-address` 设置为 `"localhost"`，这样 GoToSocial 服务器仅通过环回才可以访问。否则可能会直接连接到 GoToSocial 以绕过你的代理，这是我们不希望的。

如果 GoToSocial 已经在运行，请重启它。

```bash
sudo systemctl restart gotosocial.service
```

或者，如果你没有配置 systemd 服务，只需手动重启。

## 设置 Apache

### 所需的 Apache 模块

你需要确保安装并启用了多个 Apache 模块。这些模块应该都在你的发行版的 Apache 包中，但可能被拆分成单独的包。

你可以通过 `apachectl -M` 查看已安装哪些模块。

你需要加载以下模块：

* `proxy_http` 来代理 HTTP 请求到 GoToSocial
* `ssl` 来处理 SSL/TLS
* `headers` 来操作 HTTP 请求和响应头
* `rewrite` 来重写 HTTP 请求
* `md` 用于 Lets Encrypt，自 2.4.30 开始可用

在 Debian、Ubuntu 和 openSUSE 中，可以使用 [`a2enmod`](https://manpages.debian.org/bookworm/apache2/a2enmod.8.en.html) 工具加载任何额外的模块。对于 Red Hat/CentOS 系列发行版，你需要在 Apache 配置中添加 [`LoadModule` 指令](https://httpd.apache.org/docs/2.4/mod/mod_so.html#loadmodule)。

### 使用 mod_md 启用 TLS

!!! note "注意"
    `mod_md` 自 Apache 2.4.30 开始可用，仍被视为实验性的。实际上，它在实践中表现良好，是最便捷的方法。

现在我们将配置 Apache HTTP 服务器来处理 GoToSocial 请求。

首先，我们将在 `/etc/apache2/sites-available` 中为 Apache HTTP 服务器编写配置：

```bash
sudo mkdir -p /etc/apache2/sites-available/
sudoedit /etc/apache2/sites-available/example.com.conf
```

在上述 `sudoedit` 命令中，将 `example.com` 替换为你的 GoToSocial 服务器的域名。

你将创建的文件应如下所示：

=== "2.4.47+"
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
      # 设置为 127.0.0.1 而不是 localhost 以解决 https://stackoverflow.com/a/52550758
      ProxyPass / http://127.0.0.1:8080/ upgrade=websocket
      ProxyPassReverse / http://127.0.0.1:8080/

      RequestHeader set "X-Forwarded-Proto" expr=https
    </VirtualHost>
    ```

=== "旧版本"
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
      # 设置为 127.0.0.1 而不是 localhost 以解决 https://stackoverflow.com/a/52550758
      RewriteRule ^/?(.*) "ws://127.0.0.1:8080/$1" [P,L]

      SSLEngine On
      ProxyPreserveHost On
      # 设置为 127.0.0.1 而不是 localhost 以解决 https://stackoverflow.com/a/52550758
      ProxyPass / http://127.0.0.1:8080/
      ProxyPassReverse / http://127.0.0.1:8080/

      RequestHeader set "X-Forwarded-Proto" expr=https
    </VirtualHost>
    ```

同样，将上述配置文件中的 `example.com` 替换为你的 GoToSocial 服务器的域名。如果你的域名是 `gotosocial.example.com`，那么用 `gotosocial.example.com` 作为正确的值。

你还应该将 `http://127.0.0.1:8080` 更改为 GoToSocial 服务器的正确地址和端口（如果它不在 `127.0.0.1:8080` 上）。例如，如果你在另一台机器上以 `192.168.178.69` 的本地 IP 运行 GoToSocial，并且端口为 `8080`，那么 `http://192.168.178.69:8080/` 就是正确的值。

需要 `Rewrite*` 指令以确保 Websocket 流连接正常工作。有关更多信息，请参阅 [websocket](./websocket.md) 文档。

`ProxyPreserveHost On` 是必要的：它保证代理和 GoToSocial 使用相同的服务器名称。否则，GoToSocial 会构建错误的身份验证标头，所有联合尝试将被拒绝并返回 401 未授权。

默认情况下，Apache 会在转发的请求中设置 `X-Forwarded-For`。为了使这个设置和限速工作，设置 `trusted-proxies` 配置变量。请参阅[限速](../../api/ratelimiting.md)和[基础配置](../../configuration/general.md)文档。

保存并关闭配置文件。

现在，我们需要将刚创建的文件链接到 Apache HTTP 服务器读取已激活站点配置的文件夹中。

```bash
sudo mkdir /etc/apache2/sites-enabled
sudo ln -s /etc/apache2/sites-available/example.com.conf /etc/apache2/sites-enabled/
```

在上述 `ln` 命令中，将 `example.com` 替换为你的 GoToSocial 服务器的域名。

现在检查配置错误。

```bash
sudo apachectl -t
```

如果一切正常，你应该看到以下输出：

```text
Syntax OK
```

一切正常？太好了！然后重启 Apache HTTP 服务器以加载新的配置文件。

```bash
sudo systemctl restart apache2
```

现在，观测日志以查看新 LetsEncrypt 证书何时送达（`tail -F /var/log/apache2/error.log`），然后使用上述 `systemctl restart` 命令再次重载 Apache。之后，你应该就可以开始了！

每当 `mod_md` 获取新证书时，需要重启（或重载）Apache HTTP 服务器；请参阅该模块的文档以了解[更多信息](https://github.com/icing/mod_md#how-to-manage-server-reloads)。

根据你使用的 Apache HTTP 服务器版本，可能会看到以下错误：`error (specific information not available): acme problem urn:ietf:params:acme:error:invalidEmail: Error creating new account :: contact email "webmaster@localhost" has invalid domain : Domain name needs at least one dot`

如果发生这种情况，你需要进行以下操作之一（或全部）：

1. 更新 `/etc/apache2/sites-enabled/000-default.conf` 并将 `ServerAdmin` 值更改为有效的电子邮件地址（然后重载 Apache HTTP 服务器）。
2. 在 `/etc/apache2/sites-available/example.com.conf` 的 `MDomain` 行下添加行 `MDContactEmail your.email.address@whatever.com`，将 `your.email.address@whatever.com` 替换为有效的电子邮件地址，并将 `example.com` 替换为你的 GoToSocial 域名。

### 使用外部管理证书启用 TLS

!!! note "注意"
    我们有关于如何[配置 TLS 证书](../../advanced/certificates.md)的额外文档，其中还提供了不同发行版的其他内容和教程链接，可能值得查看。

如果你更喜欢手动设置或使用不同服务（如 Certbot）来管理 SSL，可以为你的 Apache HTTP 服务器使用更简单的设置。

首先，我们将在 `/etc/apache2/sites-available` 中为 Apache HTTP 服务器编写配置：

```bash
sudo mkdir -p /etc/apache2/sites-available/
sudoedit /etc/apache2/sites-available/example.com.conf
```

在上述 `sudoedit` 命令中，将 `example.com` 替换为你的 GoToSocial 服务器的域名。

你将创建的文件最初应如下所示，针对 80（必需）和 443 端口（可选）：

=== "2.4.47+"
    ```apache
    <VirtualHost *:80>
      ServerName example.com

      ProxyPreserveHost On
      # 设置为 127.0.0.1 而不是 localhost 以解决 https://stackoverflow.com/a/52550758
      ProxyPass / http://127.0.0.1:8080/ upgrade=websocket
      ProxyPassReverse / http://127.0.0.1:8080/
    </VirtualHost>
    ```

=== "旧版本"
    ```apache
    <VirtualHost *:80>
      ServerName example.com

      RewriteEngine On
      RewriteCond %{HTTP:Upgrade} websocket [NC]
      RewriteCond %{HTTP:Connection} upgrade [NC]
      # 设置为 127.0.0.1 而不是 localhost 以解决 https://stackoverflow.com/a/52550758
      RewriteRule ^/?(.*) "ws://127.0.0.1:8080/$1" [P,L]

      ProxyPreserveHost On
      # 设置为 127.0.0.1 而不是 localhost 以解决 https://stackoverflow.com/a/52550758
      ProxyPass / http://127.0.0.1:8080/
      ProxyPassReverse / http://127.0.0.1:8080/

    </VirtualHost>
    ```

同样，将上述配置文件中的 `example.com` 替换为你的 GoToSocial 服务器的域名。如果你的域名是 `gotosocial.example.com`，那么用 `gotosocial.example.com` 作为正确的值。

你还应该将 `http://127.0.0.1:8080` 更改为 GoToSocial 服务器的正确地址和端口（如果它不在 `127.0.0.1:8080` 上）。例如，如果你在另一台机器上以 `192.168.178.69` 的本地 IP 运行 GoToSocial，并且端口为 `8080`，那么 `http://192.168.178.69:8080/` 就是正确的值。

需要 `Rewrite*` 指令以确保 Websocket 流连接正常工作。有关更多信息，请参阅 [websocket](websocket.md) 文档。

`ProxyPreserveHost On` 是必需的：它保证代理和 GoToSocial 使用相同的服务器名称。否则，GoToSocial 会构建错误的身份验证头，所有联合尝试将被拒绝并返回 401 未授权。

在443端口提供初始设置以供外部工具进行附加管理时，你可以使用服务器提供的默认证书，你可以在 `/etc/apache2/sites-available/` 的 `default-ssl.conf` 文件中找到引用。

保存并关闭配置文件。

现在，我们需要将刚创建的文件链接到 Apache HTTP 服务器读取已激活站点配置的文件夹中。

```bash
sudo mkdir /etc/apache2/sites-enabled
sudo ln -s /etc/apache2/sites-available/example.com.conf /etc/apache2/sites-enabled/
```

在上述 `ln` 命令中，将 `example.com` 替换为你的 GoToSocial 服务器的域名。

现在检查配置错误。

```bash
sudo apachectl -t
```

如果一切正常，你应该看到以下输出：

```text
Syntax OK
```

一切正常？太好了！然后重启 Apache HTTP 服务器以加载新的配置文件。

```bash
sudo systemctl restart apache2
```

## 故障排除

如果无法在浏览器中连接到站点，则反向代理设置不起作用。比较 Apache 日志文件（`tail -F /var/log/apache2/access.log`）和 GoToSocial 日志文件。发出的请求必须在两个地方中都显示出来。仔细检查 `ProxyPass` 设置。

如果可以连接，但贴文未能联合且账户无法从其他地方找到，请检查日志。如果你看到尝试读取你的个人资料(比如 `level=INFO … method=GET statusCode=401 path=/users/your_username msg="Unauthorized: …"`)或向你的收件箱发送贴文的信息（比如 `level=INFO … method=POST statusCode=404 path=/your_username/inbox msg="Not Found: …"`），则联合已被中断。仔细检查 `ProxyPreserveHost` 设置。

如果可以连接但无法在 Mastodon 客户端应用中授权账户，请确保从正确的域名启动登录。当使用[分域](../../advanced/host-account-domain.md)设置时，必须从 `host` 域启动登录，而不是 `account-domain`。GoToSocial 设置了 `Content-Security-Policy` 头，以抵御 XSS 和数据注入攻击。该头应保持不变，确保你的反向代理没有修改、覆盖或取消设置它。
