# 配置概述

GoToSocial 力求尽可能让所有属性可配置，以适应多种不同的使用场景。

我们尽量提供合理的默认值，但在运行 GoToSocial 实例时，你需要进行*一些*配置管理。

## 配置方法

配置 GoToSocial 实例有三种不同的方法，这些方法可以根据你的设置进行组合。

### 配置文件

配置 GoToSocial 最简单的方法是将配置文件传递给 `gotosocial server start` 命令，例如：

```bash
gotosocial --config-path ./config.yaml server start
```

该命令需要一个 [YAML](https://en.wikipedia.org/wiki/YAML) 或 [JSON](https://en.wikipedia.org/wiki/JSON) 格式的文件。

可以在[这里](https://codeberg.org/superseriousbusiness/gotosocial/src/branch/main/example/config.yaml)找到示例配置文件，其中包含每个配置字段的解释、默认值和示例值。此示例文件也包含在每个发行版的下载资源中。

建议创建你自己的配置文件，只更改你需要改变的设置。这可以确保在每次发布时，你不必合并默认值的更改或者增删未从默认值更改的配置设置。

#### 在容器中挂载

你可能需要在容器中挂载一个 `config.yaml`，因为某些设置不容易通过环境变量或命令行标志管理。

为此，请在主机上创建一个 `config.yaml`，将其挂载到容器中，然后告诉 GoToSocial 读取该配置文件。可以通过将容器的运行命令设置为 `--config-path /path/inside/container/to/config.yaml` 或使用 `GTS_CONFIG_PATH` 环境变量来实现这一点。

对于 docker compose，可以这样修改配置：

```yaml
services:
  gotosocial:
    command: ["--config-path", "/gotosocial/config.yaml"]
    volumes:
      - type: bind
        source: /path/on/the/host/to/config.yaml
        target: /gotosocial/config.yaml
        read_only: true
```

或者，通过环境变量来修改配置：

```yaml
services:
  gotosocial:
    environment:
      GTS_CONFIG_PATH: /gotosocial/config.yaml
    volumes:
      - type: bind
        source: /path/on/the/host/to/config.yaml
        target: /gotosocial/config.yaml
        read_only: true
```

对于 Docker 或 Podman 命令行，需要传递一个 [符合规范的挂载参数](https://docs.podman.io/en/latest/markdown/podman-run.1.html#mount-type-type-type-specific-option)。

在使用 `docker run` 或 `podman run` 时，传递 `--config-path /gotosocial/config.yaml` 作为命令，例如：

```sh
podman run \
    --mount type=bind,source=/path/on/the/host/to/config.yaml,destination=/gotosocial/config.yaml,readonly \
    docker.io/superseriousbusiness/gotosocial:latest \
    --config-path /gotosocial/config.yaml
```

使用 `GTS_CONFIG_PATH` 环境变量：

```sh
podman run \
    --mount type=bind,source=/path/on/the/host/to/config.yaml,destination=/gotosocial/config.yaml,readonly \
    --env 'GTS_CONFIG_PATH=/gotosocial/config.yaml' \
    docker.io/superseriousbusiness/gotosocial:latest
```

### 环境变量

你也可以通过设置[环境变量](https://en.wikipedia.org/wiki/Environment_variable)来配置 GoToSocial。这些环境变量遵循的格式为：

1. 在配置标志前加上 `GTS_`。
2. 全部使用大写。
3. 将短横线（`-`）替换为下划线（`_`）。

例如，如果不想在 config.yaml 中设置 `media-image-max-size` 为 `2097152`，你可以改为设置环境变量：

```text
GTS_MEDIA_IMAGE_MAX_SIZE=2097152
```

如果对于环境变量名称有疑问，只需查看你正在使用的子命令的 `--help`。

### 命令行标志

最后，你可以使用命令行标志来设置配置值，这些标志是在运行 `gotosocial` 命令时直接传递的。例如，不在 config.yaml 或环境变量中设置 `media-image-max-size`，你可以直接通过命令行传递值：

```bash
gotosocial server start --media-image-max-size 2097152 
```

如果不确定哪些标志可用，请检查 `gotosocial --help`。

## 优先级

上述配置方法按列出的顺序相互覆盖。

```text
命令行标志 > 环境变量 > 配置文件
```

也就是说，如果你在配置文件中将 `media-image-max-size` 设置为 `2097152`，但*也*设置了环境变量 `GTS_MEDIA_MAX_IMAGE_SIZE=9999999`，则最终值将为 `9999999`，因为环境变量比 config.yaml 中设置的值具有*更高的优先级*。

命令行标志具有最高优先级，因此如果你设置了 `--media-image-max-size 13121312`，无论你在其他地方设置了什么，最终值都将为 `13121312`。

这意味着在你只想尝试改变一件事，但不想编辑配置文件的情况下，可以临时使用环境变量或命令行标志来设置那个东西。

## 默认值

*大多数*配置参数都提供了合理的默认值，除了必须自定义值的情况。

请查看[示例配置文件](https://codeberg.org/superseriousbusiness/gotosocial/src/branch/main/example/config.yaml)以获取默认值，或运行 `gotosocial --help`。

## `GTS_WAZERO_COMPILATION_CACHE`

启动时，GoToSocial 会将嵌入的 WebAssembly `ffmpeg` 和 `ffprobe` 二进制文件编译为 [Wazero](https://wazero.io/) 兼容模块，这些模块用于媒体处理，无需任何外部依赖。

为了加快 GoToSocial 的启动时间，你可以在首次启动时缓存已编译的模块，这样 GoToSocial 就不必在每次启动时从头开始编译它们。

你可以通过将环境变量 `GTS_WAZERO_COMPILATION_CACHE` 设置为一个目录来指示 GoToSocial 存储 Wazero 工件，该目录将由 GtS 用于存储两个大小约为 ~50MiB 的小型工件（总计约 ~100MiB）。

要了解此方法的示例，请参见 [docker-compose.yaml](https://codeberg.org/superseriousbusiness/gotosocial/raw/branch/main/example/docker-compose/docker-compose.yaml) 和 [gotosocial.service](https://codeberg.org/superseriousbusiness/gotosocial/raw/branch/main/example/gotosocial.service) 示例文件。

如果你希望在 systemd 或 Docker 之外为 GtS 提供此值，可以在启动 GtS 服务器时通过以下方式进行：

```bash
GTS_WAZERO_COMPILATION_CACHE=~/gotosocial/.cache ./gotosocial --config-path ./config.yaml server start
```
