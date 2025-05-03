# 贡献指引 <!-- omit in toc -->

你好！欢迎阅读 GoToSocial 的 CONTRIBUTING.md :) 感谢你的关注，为你点赞。

这些贡献指引借鉴并受到了 Gitea 的启发 (https://github.com/go-gitea/gitea/blob/main/CONTRIBUTING.md)。感谢 Gitea！

## 目录  <!-- omit in toc -->

- [介绍](#介绍)
- [错误报告与功能请求](#错误报告与功能请求)
- [合并请求](#合并请求)
  - [代码](#代码)
  - [文档](#文档)
- [开发](#开发)
  - [Golang 的分支特点](#golang-的分支特点)
  - [构建 GoToSocial](#构建-gotosocial)
    - [二进制文件](#二进制文件)
    - [Docker](#docker)
      - [使用 GoReleaser](#使用-goreleaser)
      - [手动构建](#手动构建)
  - [样式表 / Web开发](#样式表--web开发)
    - [实时加载](#实时加载)
  - [项目结构](#项目结构)
    - [浏览代码结构](#浏览代码结构)
  - [风格/代码检查/格式化](#风格代码检查格式化)
  - [测试](#测试)
    - [独立测试环境与 Pinafore](#独立测试环境与-pinafore)
    - [运行自动化测试](#运行自动化测试)
      - [SQLite](#sqlite)
      - [Postgres](#postgres)
    - [CLI 测试](#cli-测试)
    - [联合](#联合)
  - [更新 Swagger 文档](#更新-swagger-文档)
  - [CI/CD 配置](#cicd-配置)
  - [发布检查清单](#发布检查清单)
    - [如果出问题了怎么办？](#如果出问题了怎么办)

## 介绍

本文件包含一些重要信息，帮助你成功向 GoToSocial 提交的贡献。请在开启合并请求前仔细阅读！

## 错误报告与功能请求

目前，我们使用 Codeberg 的问题追踪系统来管理错误报告与功能请求。

你可以在[此处](https://codeberg.org/superseriousbusiness/gotosocial/issues "GoToSocial 的 Codeberg 问题追踪页")查看所有开放的问题。

在创建新问题之前，不论是错误还是功能请求，**请现仔细搜索所有仍处于打开状态和已被关闭的问题，以确保它尚未被解决过**。你可以使用 Codeberg 的关键字搜索来进行此操作。如果你的问题与已有问题重复，它将被关闭。

在打开功能请求之前，请考虑以下几点：

- 这个功能是否符合 GoToSocial 的范围？由于我们是小团队，我们对维护可能导致问题的[功能蔓延](https://en.wikipedia.org/wiki/Feature_creep "关于功能蔓延的维基百科文章")保持警惕。
- 这个功能是否对软件的许多用户普遍有用，还是仅适合非常具体的用例？
- 这个功能是否会对软件性能产生负面影响？如果是，这种权衡是否值得？
- 这个功能是否需要放宽 API 的安全限制？如果是，需要合理的理由。
- 这个功能是否属于 GoToSocial 的服务器后端，还是应该由客户端实现？

我们倾向于优先考虑与无障碍性、联合互通性和客户端兼容性相关的功能请求。

## 合并请求

我们欢迎新老贡献者的合并请求，但需注意以下几点：

- 你已阅读并同意我们的[行为准则](./CODE_OF_CONDUCT.md)。
- 合并请求应解决现有问题或错误（请在请求中链接相关问题），或者与文档有关。
- 如果你的合并请求引入了大量的代码或架构变更，你会愿意对这些变更的代码与架构进行一些维护工作，并解决错误。我们不欢迎引入大量维护负担的一次性合并请求！
- 合并请求质量合格。我们是小团队，时间有限，无法帮助指导合并请求或解决基本编程问题。如果你不确定，不要承担太多任务：从一个小功能或错误修复开始，将其作为你的第一个合并请求，然后逐步提高。

如果在合并请求过程中有小问题或评论，你可以[加入我们的 Matrix 空间](https://matrix.to/#/#gotosocial-space:superseriousbusiness.org "GoToSocial Matrix 空间")，地址为 `#gotosocial-space:superseriousbusiness.org`。

请阅读下面适合你计划开启的合并请求类型的相应部分。

### 代码

为了方便维护者管理，针对 GoToSocial 库的合并请求流程大致如下：

1. 为你将要解决的功能、错误或问题打开一个问题，或者在现有问题上发表评论，让大家知道你想处理它。
2. 利用开放的问题与我们讨论你的设计，收集反馈，并解决关于实现的任何问题。
3. 编写代码！确保所有现有测试通过。适当添加测试。运行代码格式化工具并更新文档。
4. 打开合并请求。如果希望对正在实现中的代码收集更多反馈，可以作为草稿提交。
5. 当你的合并请求已准备好接受审核时通知我们。
6. 等待审核。
7. 处理审核反馈，适当修改代码。如果你有合理的理由，可以对审核评论提出异议——我们都是在学习，毕竟——但请务必耐心和有礼貌。

为方便审核，请尝试将你的合并请求拆分为合理大小的提交，但不要过于追求完美：我们总是进行合并压缩。

如果你的合并请求过大，请考虑将其拆分为更小的独立合并请求以便于审核和理解。

确保你的合并请求仅包含与你尝试实现的功能或解决的错误相关的代码。不要在请求中包含对无关代码的重构：请为其创建单独的合并请求！

如果你在未遵循上述流程的情况下开启了代码合并请求，我们可能会关闭它，并要求你遵循流程。

### 文档

文档合并请求的流程比代码的稍微宽松一些。

如果你发现文档中有遗漏、错误或不明确的地方，可以自由开启合并请求进行更正；你不必先开启问题，但请在合并请求评论中解释你开启请求的原因。

我们支持基于 [Conda](https://docs.conda.io/en/latest/) 的工作流程，用于修改、构建和发布文档。以下是你可以在本地开始编辑的步骤：

* 安装 [`miniconda`](https://www.anaconda.com/docs/getting-started/miniconda/main)
* 创建你的 conda 环境：`conda env create -f ./docs/environment.yml`
* 激活环境：`conda activate gotosocial-docs`
* 在本地运行：`mkdocs serve`

然后你可以在浏览器中访问 [localhost:8000](http://127.0.0.1:8000/) 查看。

添加新页面时，需要在 [`mkdocs.yml`](mkdocs.yml) 中添加，以便它显示在侧栏的正确部分。

如果你不使用 Conda，可以阅读 `docs/environment.yml` 查看需要哪些依赖，并手动通过 `pip install` 安装这些依赖。建议在虚拟环境中进行此操作，你可以通过类似 `python3 -m venv /path-to/store-the-venv` 创建虚拟环境。之后可以调用 `/path-to/store-the-venv/bin/pip`、`/path-to/store-the-venv/bin/mkdocs` 等。

要更新依赖，在已激活的环境中使用 `conda update --update-all`。然后你可以使用以下命令更新 `environment.yml`：

```sh
conda env export -n gotosocial-docs --from-history --override-channels -c conda-forge -c nodefaults -f ./docs/environment.yml
```

注意 `conda env export` 会在 environment.yml 文件中添加 `prefix` 条目，并删除 `pip` 依赖，因此请确保移除 prefix，并重新添加 `pip` 依赖。

## 开发

### Golang 的分支特点

Golang 的一个特点是，它所依赖的源代码管理路径与 `go.mod` 中使用的路径以及各 Go 文件中的包导入路径相同。这使得使用分支版本变得有些棘手。这个问题的解决方案是先派生存储库，然后克隆上游存储库，并将上游存储库的 `origin` 设置为你派生的存储库的源。

有关更多细节，请参阅[这篇博客](https://blog.sgmansfield.com/2016/06/working-with-forks-in-go/)。

为防此文章消失，此处是步骤（有轻微修改）：

>
> 在 Codeberg 上派生存储库或设置任何其他远端 git 存储库。在这种情形下，我会转到 Codeberg 并派生存储库。
>
> 现在克隆上游存储库（而非派生的存储库）：
>
> `mkdir -p ~/go/src/code.superseriousbusiness.org && git clone git@codeberg.org:superseriousbusiness/gotosocial ~/go/src/code.superseriousbusiness.org/gotosocial`
>
> 转到你的计算机上上游存储库的顶级目录：
>
> `cd ~/go/src/code.superseriousbusiness.org/gotosocial`
>
> 将当前的 origin 远程源重命名为 upstream：
>
> `git remote rename origin upstream`
>
> 把你的派生分支添加为 origin：
>
> `git remote add origin git@codeberg.org:username/gotosocial`
>

在第一次构建项目之前，一定要运行 `git fetch`。

### 构建 GoToSocial

#### 二进制文件

要开始构建，你首先需要安装 Go。你可以在顶层目录的 `go.mod` 文件中查看需要安装的 Go 版本，然后按照[此处](https://golang.org/doc/install)的指引进行安装。

安装 Go 后，将此存储库克隆到你的 Go 路径中。通常，此路径为 `~/go/src/code.superseriousbusiness.org/gotosocial`。

安装完上述环境与依赖后，可以尝试构建项目：`./scripts/build.sh`。此命令将构建 `gotosocial` 二进制文件。

如果没有错误，太好了，你准备好了！

如果看到错误 `fatal: No names found, cannot describe anything.`，需要运行 `git fetch`。

在开发过程中，为了自动重新编译，可以使用 [nodemon](https://www.npmjs.com/package/nodemon)：

```bash
nodemon -e go --signal SIGTERM --exec "go run ./cmd/gotosocial --host localhost testrig start || exit 1"
```

#### Docker

对于以下两种方法，你需要安装 [Docker buildx](https://docs.docker.com/build/concepts/overview/#buildx)。

##### 使用 GoReleaser

GoToSocial 使用发布工具 [GoReleaser](https://goreleaser.com/intro/) 使多架构 + Docker 构建变得简单。

GoReleaser 还被 GoToSocial 用于构建和推送 Docker 镜像。

通常，这些过程由 Drone (参见 CI/CD 部分) 处理。不过，你也可以手动调用 GoReleaser 来构建快照版。

为此，首先[安装 GoReleaser](https://goreleaser.com/install/)。

接着按[样式表 / Web开发](#样式表--web开发)的说明安装 Node 和 Yarn。

最后，创建快照构建，执行：

```bash
goreleaser release --clean --snapshot
```

如果一切按计划进行，现在你应该会在 `./dist` 文件夹中找到多个架构的二进制文件和 tar，终端输出中应显示构建的快照 Docker 镜像的版本。

##### 手动构建

如果你更喜欢以简单方法构建 Docker 容器，使用更少的依赖（Node, Yarn），也可以这样构建：

```bash
./scripts/build.sh && docker buildx build -t superseriousbusiness/gotosocial:latest .
```

上述命令首先构建 `gotosocial` 二进制文件，然后调用 Docker buildx 构建容器镜像。

如果在构建过程中出现错误，提示 `"/web/assets/swagger.yaml": not found`，则需要（重新）生成 Swagger 文档，参见 [更新 Swagger 文档](#更新-swagger-文档)。

如果想为不同 CPU 架构构建 docker 镜像而不设置 buildx（例如 ARMv7 aka 32-bit ARM），首先需要通过添加以下几行到 Dockerfile 顶部来修改 Dockerfile（但不要提交此更改！）：

```dockerfile
# 使用 buildx 时，这些变量将由工具设定：
# https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope
# 但是，将它们声明为全局构建参数，允许手动使用 `--build-arg` 设置它们。
ARG BUILDPLATFORM
ARG TARGETPLATFORM
```

然后，可以使用以下命令：

```bash
GOOS=linux GOARCH=arm ./scripts/build.sh && docker build --build-arg BUILDPLATFORM=linux/amd64 --build-arg TARGETPLATFORM=linux/arm/v7 -t superseriousbusiness/gotosocial:latest .
```

另请参阅：[GOOS 和 GOARCH 值的详尽列表](https://gist.github.com/lizkes/975ab2d1b5f9d5fdee5d3fa665bcfde6)

以及：[docker 的 `--platform` 可能值的详尽列表](https://github.com/tonistiigi/binfmt/#build-test-image)

### 样式表 / Web开发

GoToSocial 使用存放于 `web/template` 文件夹下的 Gin 模板。静态资源存储于 `web/assets`。样式表和 JS 包（用于前端增强和设置界面）的源文件存储于 `web/source`，并从那里捆绑到 git 忽略的 `web/assets/dist` 文件夹。

要捆绑更改，需要 [Node.js](https://nodejs.org/en/download/) 和 [Yarn](https://classic.yarnpkg.com/en/docs/install)。

使用 [NVM](https://github.com/nvm-sh/nvm) 是安装它们的一种方便方式，还支持管理不同的 Node 版本。

安装前端依赖：

```bash
yarn --cwd ./web/source install && yarn --cwd ./web/source ts-patch install
```

`ts-patch` 步骤是必要的，因为我们使用 Typia 进行一些类型验证：参见 [Typia 安装文档](https://typia.io/docs/setup/#manual-setup)。

重新编译前端包到 `web/assets/dist`：

```bash
yarn --cwd ./web/source build
```

#### 实时加载

为了更方便的开发环境，可以在 [testrig](#测试) 中运行一个实时加载的捆绑器(bundler)。

首先用 DEBUG=1 构建 GtS 二进制文件以启用 testrig：

``` bash
DEBUG=1 ./scripts/build.sh
```

现在打开两个终端。

在第一个终端中，使用你刚构建的二进制文件在端口 8081 上运行 testrig：

```bash
DEBUG=1 GTS_PORT=8081 ./gotosocial testrig start
```

然后启动捆绑器(bundler)，它将在端口 8080 上运行，并在需要时将请求代理到 testrig 实例。

``` bash
NODE_ENV=development yarn --cwd ./web/source dev
```

然后你可以在 `http://localhost:8080/settings` 登录 GoToSocial 设置面板，并查看实时更新反映的更改。

实时加载捆绑器(bundler)*不会*更改 `dist/` 中的捆绑资源，因此完成更改并想在某处部署时，必须运行 `node web/source` 生成准备就绪的生产环境包。

### 项目结构

对于项目结构，GoToSocial 遵循 [在此处定义的标准且被广泛接受的项目布局](https://github.com/golang-standards/project-layout)。正如作者所写：

> 这是 Go 应用项目的基本布局。它不是核心 Go 开发团队定义的正式标准；然而，它是在 Go 生态系统中常见的历史和新兴项目布局模式。

在可能的情况下，我们更倾向于更短和更多的文件和包，对应用逻辑的可定义模块进行更明显的划分，而不是更少但更长的文件：如果一个 `.go` 文件接近 1000 行代码，可能就太长了。

#### 浏览代码结构

应用程序的大部分核心业务逻辑位于 `internal` 目录的各个包和子包中。以下是每个包的简要说明：

`internal/ap` - ActivityPub 工具函数和接口。

`internal/api` - 客户端与联合 (ActivityPub) API 的模型、路由和工具。在此处可以为路由器添加路由。

`internal/concurrency` - 处理器和其他队列使用的工作模式。

`internal/config` - 配置标志、CLI 标志解析及配置获取/设置的代码。

`internal/db` - 用于与 sqlite/postgres 数据库交互的数据库接口。数据库迁移代码在 `internal/db/bundb/migrations`。

`internal/email` - 通过 SMTP 发送电子邮件的功能。

`internal/federation` - ActivityPub 联合代码；实现 `go-fed` 接口。

`internal/federation/federatingdb` - 实现 `go-fed` 的数据库接口。

`internal/federation/dereferencing` - 用于从外站实例获取资源的 HTTP 调用代码。

`internal/gotosocial` - GoToSocial 服务器启动/关闭逻辑。

`internal/gtserror` - 错误模型。

`internal/gtsmodel` - 数据库和内部模型。此处包含 `bundb` 注解。

`internal/httpclient` - GoToSocial 用于发请求到外站资源的 HTTP 客户端。

`internal/id` - 生成数据库模型 ID (ULIDs) 的代码。

`internal/log` - 日志实现。

`internal/media` - 管理和处理媒体附件的代码：图像、视频、表情等。

`internal/messages` - 用于封装工作消息的模型。

`internal/middleware` - Gin Gonic 路由中间件：HTTP 签名检查、缓存控制、令牌检查等。

`internal/netutil` - HTTP/网络请求验证代码。

`internal/oauth` - OAuth 服务器实现的封装代码/接口。

`internal/oidc` - OIDC 声明和回调的封装代码/接口。

`internal/processing` - 处理联合或客户端 API 产生的消息的逻辑。GoToSocial 的核心业务逻辑大多在此处。

`internal/regexes` - 用于解析文本和匹配 URL、标签、提及的正则表达式。

`internal/router` - Gin HTTP 路由器的封装。此处包含核心 HTTP 逻辑。此路由器暴露用于附加路由的函数，由 `internal/api` 中的处理程序代码使用。

`internal/storage` - `codeberg.org/gruf/go-store` 实现的封装。此处包含本地文件存储和 S3 逻辑。

`internal/stream` - Websocket 流逻辑。

`internal/text` - 文本解析与转换。包含贴文解析逻辑——支持纯文本和 markdown。

`internal/timeline` - 贴文时间线管理代码。

`internal/trans` - 将模型导出到数据库的 JSON 备份文件，并从备份 JSON 文件导入到数据库的代码。

`internal/transport` - HTTP 传输代码和工具。

`internal/typeutils` - 在内部数据库模型和 JSON 之间进行转换，从 ActivityPub 格式到内部数据库模型格式及其反向转换的代码。基本上是序列化与反序列化。

`internal/uris` - 用于生成 GoToSocial 中使用的 URI 的工具。

`internal/util` - 零碎的工具函数，用于多个包。

`internal/validate` - 模型验证代码——目前并未真正使用。

`internal/visibility` - 贴文可见性检查和过滤。

`internal/web` - Web UI 处理程序，专门用于提供网页、登录页面、设置面板。

### 风格/代码检查/格式化

在提交代码前，建议阅读官方的简短文档 [Effective Go](https://golang.org/doc/effective_go)：这份文档是许多风格指南的基础，GoToSocial 基本遵循其建议。

我们还试图遵循的另一个风格指南是：[这个](https://github.com/bahlo/go-styleguide)。

此外，此处列举有一些符合 GtS 风格的 Uber 的 Go 风格指南亮点：

- [分组相似的声明](https://github.com/uber-go/guide/blob/master/style.md#group-similar-declarations)。
- [减少嵌套](https://github.com/uber-go/guide/blob/master/style.md#reduce-nesting)。
- [不必要的 Else](https://github.com/uber-go/guide/blob/master/style.md#unnecessary-else)。
- [局部变量声明](https://github.com/uber-go/guide/blob/master/style.md#local-variable-declarations)。
- [减少变量作用域](https://github.com/uber-go/guide/blob/master/style.md#reduce-scope-of-variables)。
- [初始化结构体](https://github.com/uber-go/guide/blob/master/style.md#initializing-structs)。

在提交代码之前，请确保执行 `go fmt ./...` 以更新空格和其他格式设置。

我们使用 [golangci-lint](https://golangci-lint.run/) 进行代码检查，通过静态代码分析捕获风格不一致和潜在的错误或安全问题。

如果你提交的 PR 未通过代码检查，将会被拒绝。因此，最好在推送或打开 PR 之前本地运行代码检查。

要做到这一点，请首先按照 [此处](https://golangci-lint.run/welcome/install/) 的说明安装代码检查工具。

然后，可以用以下命令运行代码检查：

```bash
golangci-lint run
```

如果没有输出，太好了！这说明检查通过了 :)

### 测试

GoToSocial 提供了一个 [testrig](https://codeberg.org/superseriousbusiness/gotosocial/tree/main/testrig)，包含一些可以用于集成测试的模拟包。

没有模拟的一个东西是数据库接口，因为使用内存中的 SQLite 数据库比模拟所有东西要简单得多。

#### 独立测试环境与 Pinafore

你可以启动一个在本地主机运行的独立测试服务器 testrig，可以通过 [Pinafore](https://github.com/nolanlawson/pinafore/) 连接。

要做到这一点，首先用 `DEBUG=1 ./scripts/build.sh` 构建 gotosocial 二进制文件。

然后，通过设置 `DEBUG` 环境变量启动 testrig，如下调用二进制文件：

```bash
DEBUG=1 ./gotosocial testrig start
```

要在本地开发模式下运行 Pinafore，首先克隆 [Pinafore](https://github.com/nolanlawson/pinafore/) 存储库，然后在克隆的目录中运行以下命令：

```bash
yarn # 安装依赖
yarn run dev
```

Pinafore 实例将在 `localhost:4002` 上启动。

要连接到 testrig，导航至 `http://localhost:4002`，并将在实例域名栏输入 `localhost:8080`。

在登录界面，输入电子邮件地址 `zork@example.org` 和密码 `password`。你会看到一个确认提示。接受后，你将以 Zork 身份登录。

请注意以下限制：

- 由于 testrig 使用内存数据库，因此当 testrig 停止时，数据库将被销毁。
- 如果你停止 testrig 并重新启动，则在测试期间创建的任何令牌或应用程序也会被删除。因此，你需要每次停止/启动 rig 时重新登录。
- testrig 不会进行任何实际的外部 HTTP 调用，因此联合功能无法在 testrig 工作。

#### 运行自动化测试

测试可以在 SQLite 和 Postgres 上运行。

##### SQLite

如果你想尽快运行测试，使用内存中的 SQLite 数据库，请使用：

```bash
go test ./...
```

##### Postgres

如果你想在本地运行针对 Postgres 数据库的测试，请运行：

```bash
GTS_DB_TYPE="postgres" GTS_DB_ADDRESS="localhost" go test -p 1 ./...
```

在上面的命令中，假设你使用的是默认的 Postgres 密码 `postgres`。

在 Postgres 上运行时，我们设置 `-p 1` 因为它需要串行而不是并行运行测试。

#### CLI 测试

在 [./test/envparsing.sh](./test/envparsing.sh) 中有一个测试，用于确保 CLI 标志、配置和环境变量按预期解析。

虽然此测试是 CI/CD 测试过程的一部分，但除非你在修改 `cmd/gotosocial` 中的 `main` 包或者 `internal/config` 中的 `config` 包内的代码，否则你可能不需要过多担心自行运行它。

#### 联合

通过使用从磁盘加载 TLS 文件的支持，可以启动两个或多个本地实例，其 TLS 允许（手动）测试联合。

你需要设置以下配置选项：

- `GTS_TLS_CERTIFICATE_CHAIN`：指向包含公钥证书的 PEM 编码证书链。
- `GTS_TLS_CERTIFICATE_KEY`：指向 PEM 编码的私钥。

此外，为了让 Go HTTP 客户端认可自定义 CA 签发的证书为有效，你需要设置下列变量之一：

- `SSL_CERT_FILE`：指向你的自定义 CA 的公钥。
- `SSL_CERT_DIR`：一个以 `:` 分隔的目录列表，用于加载 CA 证书。

上述 `SSL_CERT` 变量仅适用于类 Unix 系统，不包括 Mac。请参阅 https://pkg.go.dev/crypto/x509#SystemCertPool。如果你在不支持设置上述变量的架构上运行测试，可以在 `config.yaml` 文件中将 `http-client.tls-insecure-skip-verify` 设置为 `true`，以完全禁用 HTTP 客户端的 TLS 证书验证。

你还需要为两个实例名称提供功能正常的 DNS，可以通过在 `/etc/hosts` 中添加条目或运行像 [dnsmasq](https://thekelleys.org.uk/dnsmasq/doc.html) 这样的本地 DNS 服务器来实现。

### 更新 Swagger 文档

GoToSocial 使用 [go-swagger](https://goswagger.io) 根据代码注释生成 Swagger API 文档。

如果你修改了任何 API 端点上的 Swagger 注释，你可以通过运行以下命令在 `./docs/api/swagger.yaml` 生成一个新的 Swagger 文件，并通过以下命令将该文件复制到 web 资源目录中：

```bash
go run ./vendor/github.com/go-swagger/go-swagger/cmd/swagger \
generate spec --scan-models --exclude-deps -o docs/api/swagger.yaml \
&& cp docs/api/swagger.yaml web/assets/swagger.yaml
```

你无需安装 go-swagger 来运行此命令，因为 `vendor` 目录中已经包含了 go-swagger。

### CI/CD 配置

GoToSocial 使用 [Woodpecker CI](https://woodpecker-ci.org/) 进行 CI/CD 任务，如运行测试、代码检查和构建 Docker 容器。

这些运行与 Codeberg 集成，在打开拉取请求或合并到主干时执行。

`woodpecker` 流水线文件在 [此处](../../../../.woodpecker) —— 它们定义了 Woodpecker 如何运行及何时运行。

GoToSocial 的 Woodpecker 实例地址在 [此处](https://woodpecker.superseriousbusiness.org/repos/2).

Woodpecker 的文档参见 [此处](https://woodpecker-ci.org/docs/intro).