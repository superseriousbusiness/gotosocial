# 裸机

本指南将引导你在裸机上使用官方二进制发行版来运行 GoToSocial。

## 准备 VPS

在 VPS 或你的家庭服务器终端中，创建 GoToSocial 运行的目录，它将用作存储的目录，以及存放 LetsEncrypt 证书的目录。

这意味着我们需要以下目录结构：

```
.
└── gotosocial
    └── storage
        └── certs
```

你可以通过以下命令一步创建所有目录：

```bash
mkdir -p /gotosocial/storage/certs
```

如果你在机器上没有 root 权限，请使用类似 `~/gotosocial` 的路径。

## 下载发行版

在 VPS 或你的家庭服务器终端中，进入你刚创建的 GoToSocial 根目录：

```bash
cd /gotosocial
```

现在，下载与你运行的操作系统和架构相对应的最新 GoToSocial 发行版压缩包。

!!! tip
    你可以在[这里](https://github.com/superseriousbusiness/gotosocial/releases)找到按发布时间排列的发布列表，最新的发行版位于最上面。

例如，下载适用于 64 位 Linux 的版本：

```bash
GTS_VERSION=X.Y.Z # 替换此处
GTS_TARGET=linux_amd64
wget https://github.com/superseriousbusiness/gotosocial/releases/download/v${GTS_VERSION}/gotosocial_${GTS_VERSION}_${GTS_TARGET}.tar.gz
```

然后解压：

```bash
tar -xzf gotosocial_${GTS_VERSION}_${GTS_TARGET}.tar.gz
```

这将在你的当前目录放置 `gotosocial` 二进制文件，以及包含网页前端资源的 `web` 文件夹和包含示例配置文件的 `example` 文件夹。

!!! danger
    如果你想使用基于当前主分支代码的 GoToSocial 快照构建，可以从[这里](https://minio.s3.superseriousbusiness.org/browser/gotosocial-snapshots)下载最近的二进制 .tar.gz 文件（基于提交哈希）。仅在你很清楚自己的操作时使用，否则请使用稳定版。

## 编辑配置文件

基于 `example` 文件夹中的 `config.yaml` 创建一个新的配置文件。你可以复制整个文件，但请确保仅保留已更改的设置。这让检查发布升级时的配置变化更加容易。

你可能需要更改以下设置：

- 设置 `host` 为你要运行服务器的域名（例如 `example.org`）。
- 设置 `port` 为 `443`。
- 设置 `db-type` 为 `sqlite`。
- 设置 `db-address` 为 `sqlite.db`。
- 设置 `storage-local-base-path` 为你上面创建的存储目录（例如 `/gotosocial/storage`）。
- 设置 `letsencrypt-enabled` 为 `true`。
- 设置 `letsencrypt-cert-dir` 为你上面创建的证书存储目录（例如 `/gotosocial/storage/certs`）。

上述选项假设你使用 SQLite 作为数据库。如果你想使用 Postgres，请参阅[这里](../../configuration/database.md)获取配置选项。

!!! info "可选配置"
    
    `config.yaml` 文件中记录了许多其他配置选项，可以进一步自定义你的 GoToSocial 实例的行为。这些选项在可能的情况下使用合理的默认值，因此现在不必对此进行任何更改，但以下是一些你可能感兴趣的选项：
    
    - `instance-languages`: 确定实例首选语言的 [BCP47 语言标签](https://en.wikipedia.org/wiki/IETF_language_tag)数组。
    - `media-remote-cache-days`: 保存在存储中外站媒体的缓存天数。
    - `smtp-*`: 允许你的 GoToSocial 实例连接到邮件服务器并发送通知邮件的设置。

    如果你决定稍后设置/更改这些变量，请确保在进行更改后重新启动你的 GoToSocial 实例。

## 运行二进制文件

你现在可以运行二进制文件了。

使用以下命令启动 GoToSocial 服务器：

```bash
./gotosocial --config-path ./config.yaml server start
```

服务器应该现在启动，并且你应该能通过浏览器访问你的域名的启动页面。请注意，首次创建 LetsEncrypt 证书可能需要最多一分钟的时间，因此如有必要请多次刷新页面。

注意在本例中我们假设可以运行在端口 443（标准 HTTPS 端口），并且没有其他进程运行在该端口。

## 创建你的用户

你可以使用 GoToSocial 二进制文件来创建并提权你的用户账户。所有这些过程在我们的[创建用户](../user_creation.md)指南中均有记录。

## 登录

你现在应该可以使用刚创建的账户的电子邮件地址和密码登录到你的实例。

## （可选）启用 systemd 服务

如果你不喜欢每次启动时手动启动 GoToSocial，你可能希望创建一个 systemd 服务为你启动。

首先，停止你的 GoToSocial 实例。

然后为你的 GoToSocial 安装创建一个新用户和用户组：

```bash
sudo useradd -r gotosocial
sudo groupadd gotosocial
sudo usermod -a -G gotosocial gotosocial
```

然后使其成为你的 GoToSocial 安装目录的所有者，因为它们需要在其中进行读写：

```bash
sudo chown -R gotosocial:gotosocial /gotosocial
```

你可以在 [GitHub](https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/example/gotosocial.service) 或你的安装文件夹中的 `example` 文件夹中找到一个 `gotosocial.service` 文件。

将它复制到 `/etc/systemd/system/gotosocial.service`：

```bash
sudo cp /gotosocial/example/gotosocial.service /etc/systemd/system/
```

然后使用 `sudoedit /etc/systemd/system/gotosocial.service` 在编辑器中打开文件。如果你在与本指南中使用的 `/gotosocial` 路径不同的目录中安装了 GoToSocial，请根据你的安装修改 `ExecStart=` 和 `WorkingDirectory=` 行。

!!! info "运行在端口 80 和 443"
    
    如果你完全遵循本指南，你的 GoToSocial 实例将配置为绑定到端口 443 和 80，它们是已知的特权端口。要允许 GoToSocial 用户绑定到这些端口，你需要通过删除前导 `#` 来取消注释服务文件中的 `CAP_NET_BIND_SERVICE` 行。
    
    修改前：
    
    ```
    #AmbientCapabilities=CAP_NET_BIND_SERVICE
    ```
    
    修改后：
    
    ```
    AmbientCapabilities=CAP_NET_BIND_SERVICE
    ```
    
    如果你以后决定使用反向代理运行 GoToSocial（见下文），你可能希望重新注释此行以移除权限，因为反向代理将绑定到特权端口。

编辑完成后，保存并关闭文件，运行以下命令以启用服务：

```bash
sudo systemctl enable --now gotosocial.service
```

GoToSocial 现在应该已启动并运行。

## （可选）反向代理

如果你想在端口 443 上运行其他网络服务器或想添加额外的安全层，你可能希望使用[反向代理](../reverse_proxy/index.md)。我们提供了几个流行开源选项的指南，并非常欢迎提供更多指南的 pull requests。
