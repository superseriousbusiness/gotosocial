# 容器

本指南将指引你通过我们发布的官方容器镜像运行 GoToSocial。在本例中，我们将直接使用 [Docker Compose](https://docs.docker.com/compose) 和 SQLite 作为数据库。

你也可以使用容器编排系统（如 [Kubernetes](https://kubernetes.io/) 或 [Nomad](https://www.nomadproject.io/)）运行 GoToSocial，但这超出了本指南的范围。

## 创建工作目录

你需要一个工作目录来存放你的 docker-compose 文件，以及一个目录来存储 GoToSocial 的数据。请使用以下命令创建这些目录：

```bash
mkdir -p ~/gotosocial/data
```

现在切换到你创建的工作目录：

```bash
cd ~/gotosocial
```

## 获取最新的 docker-compose.yaml

使用 `wget` 下载最新的 [docker-compose.yaml](https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/example/docker-compose/docker-compose.yaml) 示例，我们将根据需要进行自定义：

```bash
wget https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/example/docker-compose/docker-compose.yaml
```

## 编辑 docker-compose.yaml

由于 GoToSocial 可以使用[环境变量](../../configuration/index.md#环境变量)进行配置，我们可以跳过在容器中挂载 config.yaml 文件，使配置更为简单。只需编辑 docker-compose.yaml 文件以更改一些内容。

首先在你的编辑器中打开 docker-compose.yaml 文件。例如：

```bash
nano docker-compose.yaml
```

### 版本

如果需要，更新 GoToSocial Docker 镜像标签到你想要使用的 GtS 版本：

* `latest`：默认值。这指向最新的稳定版本的 GoToSocial。
* `snapshot`：指向当前在主分支上的代码。不保证稳定，可能经常出错。谨慎使用。
* `vX.Y.Z`：发布标签。这指向 GoToSocial 的特定、稳定的版本。

!!! tip
    `latest` 和 `snapshot` 标签是动态标签，而 `vX.Y.Z` 标签是固定的。拉取动态标签的结果可能每天都会变化。同一系统上的 `latest` 可能与不同系统上的 `latest` 不同。建议使用 `vX.Y.Z` 标签，以便你始终确切知道运行的是 GoToSocial 的哪个版本。发布列表可以在[这里](https://github.com/superseriousbusiness/gotosocial/releases)找到，最新的发布在顶部。

### 主机

更改 `GTS_HOST` 环境变量为你运行 GoToSocial 的域名。

### 服务器时区（可选但推荐）

为确保你的 GoToSocial 服务器在贴文和日志中显示正确的时间，你可以通过编辑 `TZ` 环境变量设置服务器的时区。

1. 删除环境部分中 `TZ: UTC` 前面的 `#`。
2. 将 `UTC` 部分更改为你的时区标识符。有关这些标识符的列表，请参阅 https://en.wikipedia.org/wiki/List_of_tz_database_time_zones。

例如，如果你在明斯克运行服务器，你可以设置 `TZ: Europe/Minsk`，日本设置为 `TZ: Japan`，迪拜设置为 `TZ: Asia/Dubai`，等等。

如果不设置，将使用默认的 `UTC`。

### 用户（可选/可能不必要）

默认情况下，Docker 化的 GoToSocial 以 Linux 用户/组 `1000:1000` 运行，这在大多数情况下是可以的。如果你想以不同的用户/组运行，应相应地更改 docker-compose.yaml 中的 `user` 字段。

例如，假设你为 id 为 `1001` 的用户和组创建了 `~/gotosocial/data` 目录。如果现在不更改 `user` 字段就尝试运行 GoToSocial，将会遇到权限错误，无法在目录中打开数据库文件。在这种情况下，你需要将 docker compose 文件的 `user` 字段更改为 `1001:1001`。

### LetsEncrypt（可选）

如果你想为 TLS 证书（https）使用 [LetsEncrypt](../../configuration/tls.md)，你还应该：

1. 将 `GTS_LETSENCRYPT_ENABLED` 的值更改为 `"true"`。
2. 删除 `ports` 部分中 `- "80:80"` 前面的 `#`。
3. （可选）将 `GTS_LETSENCRYPT_EMAIL_ADDRESS` 设置为有效的电子邮件地址，以接收证书过期警告等。

!!! info "可选配置"
    
    config.yaml 文件中记录了许多其他配置选项，你可以使用这些选项进一步自定义你的 GoToSocial 实例的行为。尽可能使用合理的默认设置，因此不一定需要立即对它们进行更改，但以下几个可能会感兴趣：
    
    - `GTS_INSTANCE_LANGUAGES`：确定你实例首选语言的 [BCP47 语言标签](https://en.wikipedia.org/wiki/IETF_language_tag)数组。
    - `GTS_MEDIA_REMOTE_CACHE_DAYS`：在存储中保持外站媒体缓存的天数。
    - `GTS_SMTP_*`：允许你的 GoToSocial 实例连接到电子邮件服务器并发送通知电子邮件的设置。

    如果你决定稍后设置/更改这些变量，请确保在更改后重新创建 GoToSocial 实例容器。
    

!!! tip
    
    有关将 config.yaml 文件中的变量名称转换为环境变量的帮助，请参阅[配置部分](../../configuration/index.md#environment-variables)。

### Wazero 编译缓存（可选）

启动时，GoToSocial 会将嵌入的 WebAssembly `ffmpeg` 和 `ffprobe` 二进制文件编译为 [Wazero](https://wazero.io/) 兼容模块，用于媒体处理而无需任何外部依赖。

要加快 GoToSocial 的启动时间，你可以在重启之间缓存已编译的模块，这样 GoToSocial 就不必在每次启动时从头编译它们。

如果你希望在 Docker 容器中进行此操作，首先在工作文件夹中创建一个 `.cache` 目录以存储模块：

```bash
mkdir -p ~/gotosocial/.cache
```

然后，取消注释 docker-compose.yaml 文件中第二个卷的前面的 `#` 符号，使其从

```yaml
#- ~/gotosocial/.cache:/gotosocial/.cache
```

变为

```yaml
- ~/gotosocial/.cache:/gotosocial/.cache
```

这将指示 Docker 在 Docker 容器中将 `~/gotosocial/.cache` 目录挂载到 `/gotosocial/.cache`。

## 启动 GoToSocial

完成这些小改动后，您现在可以使用以下命令启动 GoToSocial：

```shell
docker-compose up -d
```

运行此命令后，你应该会看到如下输出：

```text
Creating network "gotosocial_gotosocial" with the default driver
Creating gotosocial ... done
```

如果你想跟踪 GoToSocial 的日志，可以使用：

```bash
docker logs -f gotosocial
```

如果一切正常，你应该会看到类似以下的内容：

```text
time=2022-04-19T09:48:35Z level=info msg=connected to SQLITE database
time=2022-04-19T09:48:35Z level=info msg=MIGRATED DATABASE TO group #1 (20211113114307, 20220214175650, 20220305130328, 20220315160814) func=doMigration
time=2022-04-19T09:48:36Z level=info msg=instance account example.org CREATED with id 01EXX0TJ9PPPXF2C4N2MMMVK50
time=2022-04-19T09:48:36Z level=info msg=created instance instance example.org with id 01PQT31C7BZJ1Q2Z4BMEV90ZCV
time=2022-04-19T09:48:36Z level=info msg=media manager cron logger: start[]
time=2022-04-19T09:48:36Z level=info msg=media manager cron logger: schedule[now 2022-04-19 09:48:36.096127852 +0000 UTC entry 1 next 2022-04-20 00:00:00 +0000 UTC]
time=2022-04-19T09:48:36Z level=info msg=started media manager remote cache cleanup job: will run next at 2022-04-20 00:00:00 +0000 UTC
time=2022-04-19T09:48:36Z level=info msg=listening on 0.0.0.0:8080
```

## 创建你的第一个用户

现在 GoToSocial 已在运行，你应该至少为自己创建一个用户。如何创建用户可以在我们的[创建用户](../user_creation.md)指南中找到。

### 完成

GoToSocial 现在应该在你的机器上运行！要验证这一点，打开浏览器，导航到你设置的 `GTS_HOST` 值。你应该会看到 GoToSocial 的登陆页面。干得不错！

## （可选）反向代理

如果你想在 443 端口上运行其他网络服务器或想增加额外的安全层，你可能需要使用[反向代理](../reverse_proxy/index.md)。我们为几个流行的开源选项提供了指南，并乐意接受更多的拉取请求以增加新的指南。
