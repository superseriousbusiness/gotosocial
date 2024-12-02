# Caddy 2

## 要求

在此指南中，你需要使用 [Caddy 2](https://caddyserver.com/)，无需其他依赖。Caddy 管理 Lets Encrypt 证书及其续订。

Caddy 可以通过大多数流行的包管理器获取，或者你可以获取一个静态二进制文件。最新的安装指南请参考[他们的手册](https://caddyserver.com/docs/install)。

### Debian, Ubuntu, Raspbian

```bash
# 为其自定义仓库添加密钥环。
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list

# 更新软件包并安装
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

## 配置 GoToSocial

如果 GoToSocial 已经在运行，先停止它。

```bash
sudo systemctl stop gotosocial
```

在你的 GoToSocial 配置中，通过将 `letsencrypt-enabled` 设置为 `false` 来关闭 Lets Encrypt。

如果你之前在 443 端口运行 GoToSocial，需将 `port` 值改回默认的 `8080`。

如果反向代理将在同一台机器上运行，将 `bind-address` 设置为 `"localhost"`，这样 GoToSocial 服务器只能通过回环地址访问。否则可能会有人直接连接到 GoToSocial 以绕过你的代理，这是不安全的。

## 设置 Caddy

我们将配置 Caddy 2 来在主域名 example.org 上使用 GoToSocial。由于 Caddy 负责获取 Lets Encrypt 证书，我们只需正确配置它一次。

在最简单的使用场景中，Caddy 默认使用名为 Caddyfile 的文件。它可以在更改时重新加载，或者通过 HTTP API 配置以实现零停机，但这超出了我们当前的讨论范围。

```bash
sudo mkdir -p /etc/caddy
sudo vim /etc/caddy/Caddyfile
```

在编辑上述文件时，你应将 'example.org' 替换为你的域名。你的域名应在当前配置中出现两次。如果你为 GoToSocial 选择了端口号 8080 以外的端口，请在反向代理行中更改端口号以匹配它。

你即将创建的文件应如下所示：

```Caddyfile
example.org {
	# 可选，但推荐，使用适当的协议压缩流量
	encode zstd gzip

	# 实际的代理配置为端口 8080（除非你选择了其他端口号）
	reverse_proxy * http://127.0.0.1:8080 {
		# 立即刷新，以防止缓冲响应给客户端
		flush_interval -1
	}
}
```

默认情况下，caddy 在转发请求中设置 `X-Forwarded-For`。为了使其与速率限制配合使用，请设置 `trusted-proxies` 配置变量。详见[速率限制](../../api/ratelimiting.md)和[通用配置](../../configuration/general.md)文档。

有关进阶配置，请查看 Caddy 文档中的[反向代理指令](https://caddyserver.com/docs/caddyfile/directives/reverse_proxy)。

现在检查配置错误。

```bash
sudo caddy validate
```

如果一切正常，你将看到一些信息行作为输出。除非前面标有 *[err]* 的行，否则你就准备好了。

一切正常吗？太好了！然后重启 caddy 以加载你的新配置文件。

```bash
sudo systemctl restart caddy
```

如果一切顺利，你现在就可以享受你的 GoToSocial 实例，所以我们将再次启动它。

```bash
sudo systemctl start gotosocial
```

## 结果

你现在应该能够在浏览器中打开你的实例的启动页面，并会看到它在 HTTPS 下运行！
