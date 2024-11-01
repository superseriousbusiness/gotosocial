# TLS

你可以通过以下两种方式配置 TLS 支持：
* 内置支持 Lets Encrypt / ACME 兼容供应商
* 从磁盘加载 TLS 文件

不能同时启用这两种方法。

注意，当使用从磁盘加载的 TLS 文件时，你需要在文件更改时重新启动实例。文件不会自动重新加载。

## 设置

```yaml
##############################
##### LETSENCRYPT 配置 #####
##############################

# 与自动获取和使用 LetsEncrypt HTTPS 证书相关的配置。

# 布尔值。是否为服务器启用 letsencrypt。
# 如果为 false，这里的其余设置将被忽略。
# 如果你的 GoToSocial 服务部署在 nginx 或 traefik 这样的反向代理后，请保持关闭状态。
# 如果没有，请开启以便可以使用 https。
# 选项：[true, false]
# 默认值：false
letsencrypt-enabled: false

# 整数。监听 letsencrypt 证书挑战的端口。
# 如果启用了 letsencrypt，则该端口必须可达，否则你将无法获取证书。
# 如果没有启用 letsencrypt，则该端口将不会使用。
# 这 *不能* 与上面指定的 webserver/API 端口相同。
# 例子：[80, 8000, 1312]
# 默认值：80
letsencrypt-port: 80

# 字符串。存储 LetsEncrypt 证书的目录。
# 最好将其设置为存储目录中的子路径，以便于备份，
# 但如果其他服务也需要访问这些证书，你可能希望将它们移到别的地方。
# 无论如何，请确保 GoToSocial 有权限写入/读取此目录。
# 例子：["/home/gotosocial/storage/certs", "/acmecerts"]
# 默认值："/gotosocial/storage/certs"
letsencrypt-cert-dir: "/gotosocial/storage/certs"

# 字符串。注册 LetsEncrypt 证书时使用的电子邮件地址。
# 此电子邮件地址很可能是实例管理员的地址。
# LetsEncrypt 将发送关于证书到期等的通知到此地址。
# 例子：["admin@example.org"]
# 默认值：""
letsencrypt-email-address: ""

##############################
##### 手动 TLS 配置  #####
##############################

# 字符串。磁盘上 PEM 编码文件的路径，包含证书链和公钥。
# 例子：["/gotosocial/storage/certs/chain.pem"]
# 默认值：""
tls-certificate-chain: ""

# 字符串。磁盘上 PEM 编码文件的路径，包含与 tls-certificate-chain 相关的私钥。
# 例子：["/gotosocial/storage/certs/private.pem"]
# 默认值：""
tls-certificate-key: ""
```
