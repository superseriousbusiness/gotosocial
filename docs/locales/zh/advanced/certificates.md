# 配置 TLS 证书

如[部署注意事项](../getting_started/index.md)中所述，联合需要使用 TLS，因为大多数实例拒绝通过未加密的传输进行联合。

GoToSocial 内置了通过 Lets Encrypt 进行 TLS 证书的配置和更新支持。本指南介绍如何独立于 GoToSocial 配置证书。如果你想完全控制证书的配置方式，或者因为你正在使用执行 TLS 终止的[反向代理](../getting_started/reverse_proxy/index.md)，这将很有用。

获取 TLS 证书的方式有几种：

* 从供应商购买，通常有效期为 2 年
* 从云提供商获取，具体有效期取决于其产品限制
* 从像 Lets Encrypt 这样的[ACME](https://en.wikipedia.org/wiki/Automatic_Certificate_Management_Environment)兼容提供商处获取，通常有效期为 3 个月

在本指南中，我们只讨论第三种，有关 ACME 兼容提供商的选项。

## 一般方法

通过 Lets Encrypt 配置证书的方法是：

* 在你的服务器上安装 ACME 客户端
* 配置 ACME 客户端来配置你的证书
* 配置一个软件使用这些证书
* 启用定时器/cron 定期续订证书
* 通知必要的应用程序重新加载或重启以获取新证书

证书是通过[使用质询](https://letsencrypt.org/sv/docs/challenge-types/)来配置的，这是一种验证你为自己控制的域请求证书的方法。你通常会使用以下之一：

* HTTP 质询
* DNS 质询

HTTP 质询要求在所请求证书的域上的 80 端口下提供某些文件，路径为 `/.well-known/acme/`。这是默认质询类型。

DNS 质询完全在服务器外进行，但需要你更新 DNS TXT 记录。此方法只有在你的 DNS 注册商提供 API，使你的 ACME 客户端完成此质询时才可行。

## 客户端

官方的 Lets Encrypt 客户端是 [certbot](https://certbot.eff.org/)，通常在你选择的[（Linux）发行版](https://repology.org/project/certbot/versions)中打包。某些反向代理如 Caddy 和 Traefik 内置了使用 ACME 协议配置证书的支持。

你可以考虑使用的其他一些客户端包括：

* [acme-client](https://man.openbsd.org/acme-client.1)，适用于 OpenBSD，使用平台的特权分离功能
* [lacme](https://git.guilhem.org/lacme/about/)，以进程隔离和最低特权为构建目标，类似于 acme-client 但适用于 Linux
* [Lego](https://github.com/go-acme/lego)，用 Go 编写的 ACME 客户端和库
* [mod_md](https://httpd.apache.org/docs/2.4/mod/mod_md.html)，适用于 Apache 2.4.30+

### DNS 质询

对于 DNS 质询，你的注册商的 API 需要被你的 ACME 客户端支持。尽管 certbot 对一些流行提供商有一些插件，但你可能想查看 [dns-multi](https://github.com/alexzorin/certbot-dns-multi) 插件。它在幕后使用 [Lego](https://github.com/go-acme/lego)，支持更广泛的供应商。

## 配置

有三个重要的配置选项：

* [`letsencrypt-enabled`](../configuration/tls.md) 控制 GoToSocial 是否尝试配置自己的证书
* [`tls-certificate-chain`](../configuration/tls.md) 文件系统路径，GoToSocial 可以在此找到 TLS 证书链和公钥
* [`tls-certificate-key`](../configuration/tls.md) 文件系统路径，GoToSocial 可以在此找到关联的 TLS 私钥

### 不使用反向代理

当直接将 GoToSocial 暴露到互联网，但仍想使用自己的证书时，可以设置以下选项：

```yaml
letsencrypt-enabled: false
tls-certificate-chain: "/path/to/combined-certificate-chain-public.key"
tls-certificate-key: "/path/to/private.key"
```

这将禁用通过 Lets Encrypt 内置的证书配置，并指示 GoToSocial 在磁盘上找到证书。

!!! tip
    在续订证书后应重启 GoToSocial。它在这种情况下不会自动监测证书的更换。

### 使用反向代理

当在执行 TLS 终止的[反向代理](../getting_started/reverse_proxy/index.md)后运行 GoToSocial 时，你需要如下设置：

```yaml
letsencrypt-enabled: false
tls-certificate-chain: ""
tls-certificate-key: ""
```

确保 `tls-certificate-*` 选项未设置或设置为空字符串。否则，GoToSocial 将尝试自行处理 TLS。

!!! danger "协议配置选项"
    **不要**将 [`protocol`](../configuration/general.md) 配置选项更改为 `http`。此选项仅应在开发环境中设置为 `http`。即使在 TLS 终止的反向代理后运行，也需要设置为 `https`。

你还需要更改 GoToSocial 绑定的[`port`](../configuration/general.md)，以便它不再尝试使用 443 端口。

要在反向代理中配置 TLS，请参考其文档：

* [nginx](https://docs.nginx.com/nginx/admin-guide/security-controls/terminating-ssl-http/)
* [apache](https://httpd.apache.org/docs/2.4/ssl/ssl_howto.html)
* [Traefik](https://doc.traefik.io/traefik/https/tls/)
* [Caddy](https://caddyserver.com/docs/caddyfile/directives/tls)

!!! tip
    在你的反向代理中配置 TLS 时，请确保你配置了一组较现代的兼容版本和加密套件。可以使用 [Mozilla SSL Configuration Generator](https://ssl-config.mozilla.org/) 的“中级”配置。

    检查你的反向代理文档，以了解在证书更改后是否需要重新加载或重启它。并非所有的反向代理都会自动检测到这一点。

## 指南

网上有许多优质资源解释如何设置这些内容。

* [ArchWiki 条目](https://wiki.archlinux.org/title/certbot)关于 certbot
* [Gentoo wiki 条目](https://wiki.gentoo.org/wiki/Let%27s_Encrypt)关于 Lets Encrypt
* [Linode 指南](https://www.linode.com/docs/guides/enabling-https-using-certbot-with-nginx-on-fedora/)关于 Fedora、RHEL/CentOS、Debian 和 Ubuntu 上的 certbot
* Digital Ocean 指南关于在 Ubuntu 22.04 上用 [nginx](https://www.digitalocean.com/community/tutorials/how-to-secure-nginx-with-let-s-encrypt-on-ubuntu-22-04) 或 [apache](https://www.digitalocean.com/community/tutorials/how-to-secure-apache-with-let-s-encrypt-on-ubuntu-22-04)使用 Lets Encrypt
