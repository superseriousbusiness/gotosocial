<!--overview-start-->
# GoToSocial <!-- omit in toc -->

**有关企业赞助的更新：我们欢迎与符合我们价值观的组织建立赞助关系；请查看下述条件**

GoToSocial 是一个用 Golang 编写的 [ActivityPub](https://activitypub.rocks/) 社交网络服务端。

通过 GoToSocial，你可以与朋友保持联系，发帖、阅读和分享图片及文章，且不会被追踪或广告打扰！

<p align="middle">
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/sloth.webp" width="300"/>
</p>

**GoToSocial 仍然是 [BETA 软件](https://en.wikipedia.org/wiki/Software_release_life_cycle#Beta)**。它已经可被部署和使用，并能与许多其他 Fediverse 服务端顺利联合（但还不是与所有服务端）。然而，许多功能尚未实现，而且还有不少漏洞！我们在 2024 年 9 月/10 月离开了 Alpha 阶段，并计划于 2026 年结束 Beta。

文档位于 [docs.gotosocial.org](https://docs.gotosocial.org/zh-cn/)。你可以直接跳至 [API 文档](https://docs.gotosocial.org/zh-cn/latest/api/swagger/)。

要从源代码构建，请查看 [CONTRIBUTING.md](https://github.com/superseriousbusiness/gotosocial/blob/main/docs/locales/zh/repo/CONTRIBUTING.md) 文件。

这是实例首页的截图！

![GoToSocial 实例 goblin.technology 的首页截图。它展示了实例的基本信息，如用户数和贴文数等。](https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/instancesplash.png)
<!--overview-end-->

## 目录 <!-- omit in toc -->

- [什么是 GoToSocial？](#什么是-gotosocial)
  - [联合](#联合)
  - [历史与现状](#历史与现状)
- [功能](#功能)
  - [兼容 Mastodon API](#兼容-mastodon-api)
  - [精细的贴文可见性设置](#精细的贴文可见性设置)
  - [回复控制](#回复控制)
  - [仅本站贴文](#仅本站贴文)
  - [RSS 源](#rss-源)
  - [富文本格式化](#富文本格式化)
  - [主题与自定义 CSS](#主题与自定义-css)
  - [易于运行](#易于运行)
  - [隐私+安全功能](#隐私安全功能)
  - [多种联合模式](#多种联合模式)
  - [OIDC 集成](#oidc-集成)
  - [后端优先设计](#后端优先设计)
- [已知问题](#已知问题)
- [安装 GoToSocial](#安装-gotosocial)
  - [支持的平台](#支持的平台)
    - [FreeBSD](#freebsd)
    - [32位](#32位)
    - [OpenBSD](#openbsd)
  - [稳定版本](#稳定版本)
  - [快照版本](#快照版本)
    - [Docker](#docker)
    - [二进制发布 .tar.gz](#二进制发布-targz)
  - [从源代码构建](#从源代码构建)
  - [第三方打包](#第三方打包)
- [参与贡献](#参与贡献)
- [联系我们](#联系我们)
- [致谢](#致谢)
  - [库](#库)
  - [图像归属与许可](#图像归属与许可)
  - [团队成员](#团队成员)
  - [特别鸣谢](#特别鸣谢)
- [赞助与资金支持](#赞助与资金支持)
  - [众筹](#众筹)
  - [企业赞助](#企业赞助)
  - [NLnet](#nlnet)
- [许可](#许可)

<!--body-1-start-->
## 什么是 GoToSocial？

GoToSocial 提供了一个轻量级、可定制且注重安全的进入 [联邦宇宙](https://en.wikipedia.org/wiki/Fediverse) 的入口，它类似但不同于像 [Mastodon](https://joinmastodon.org/)、[Pleroma](https://pleroma.social/)、[Friendica](https://friendi.ca) 和 [PixelFed](https://pixelfed.org/) 这样的现有项目。

如果你曾使用过 Twitter 或 Tumblr（甚至是 Myspace）等服务，GoToSocial 可能会让你感到熟悉：你可以关注他人并拥有粉丝，发布贴文，点赞、回复和分享他人的帖子，并通过时间线浏览你关注的人的贴文。你可以撰写长篇或短篇贴文，或者仅发布图片，一切随你选择。当然，你也可以屏蔽他人，或通过选择仅向朋友发布来限制不想要的互动。

![GoToSocial 中的网页版账户页截图，显示了头像、简介和粉丝/关注人数。](https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/profile1.png)

**GoToSocial 不使用推荐算法，也不收集你的数据来推荐内容或“改善你的体验”**。时间线是按时间顺序排列的：你在时间线顶部看到的内容是*刚刚发布的*，而不是根据你的个人资料选择的“有趣”或“有争议”的内容。

GoToSocial 并不是为拥有成千上万粉丝的“必追”网红设计的，也不是设计被用来让人上瘾的。你的时间线和体验由你关注的人和你与他人的互动方式决定，而不是你的参与度的相关指标！

GoToSocial 不会宣称比其他应用更“好”，但它提供了一些可能特别*适合你*的东西。

### 联合

因为 GoToSocial 使用 [ActivityPub](https://activitypub.rocks/)，你不仅可以与本站上的人交流，还可以无缝与 [联邦宇宙](https://en.wikipedia.org/wiki/Fediverse) 上的人交流。

![activitypub 标志](https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/ap_logo.svg)

联合意味着你的实例是一个遍布世界的、使用相同协议通信的服务器网络的一部分。你的数据不再集中在一家公司服务器上，而是在你自己的服务器上，根据你的意愿，跨越由其他人运行的服务器组成的弹性网络实现共享。

这种联合方式也意味着你不必受制于可能远在千里之外的庞大公司设定的任意规则。你的实例有自己的规则和文化；你的实例的居民是你的网上邻居；你很可能会认识你的服务器管理员和站务，或者自己成为管理员。

GoToSocial 的愿景是让许多小而特别的实例遍布联邦宇宙，让人们感到宾至如归，而不是让联邦宇宙被少数大的通用的实例占据，在那里一个人的声音可能会在大量其它账号的声音中迷失。

### 历史与现状

该项目于 2021 年 2/3 月因对其他联合式微博/社交媒体应用的安全和隐私功能的不满而起步，并希望实现一些不同的东西。

它最初是一个个人项目，然后随着更多开发者的兴趣和加入而加速发展。

我们在 2021 年 11 月进行了首次 Alpha 发布。我们于 2024 年 9 月/10 月离开 Alpha，进入 Beta 阶段。

要详细了解已实现和未实现的内容，以及 [稳定发布](https://en.wikipedia.org/wiki/Software_release_life_cycle#Stable_release) 的进展，请查看 [路线图](https://github.com/superseriousbusiness/gotosocial/blob/main/docs/locales/zh/repo/ROADMAP.md)。

---

## 功能

### 兼容 Mastodon API

Mastodon API 已成为客户端与联邦宇宙服务端通信的事实标准，因此 GoToSocial 实现并在自定义功能上扩展了该 API。

大多数实现 Mastodon API 的应用程序都应该可以使用 GoToSocial，但以下这些优秀的应用程序已经过测试，可与 GoToSocial 可靠地配合使用：

* [Tusky](https://tusky.app/) 适用于 Android
* [Semaphore](https://semaphore.social/) 适用于浏览器
* [Feditext](https://github.com/feditext/feditext) (beta) 适用于 iOS, iPadOS 和 macOS

如果你之前通过第三方应用来使用 Mastodon，使用 GoToSocial 将是轻而易举的。

### 精细的贴文可见性设置

发布内容时，选择谁能看到很重要。

GoToSocial 提供公开、不列出/悄悄公开、仅粉丝和私信（最好让对方事先同意）的贴文选项。

### 回复控制

GoToSocial 允许你通过 [互动规则](https://docs.gotosocial.org/zh-cn/latest/user_guide/settings/#default-interaction-policies) 选择谁可以回复你的贴文。你可以选择允许任何人回复贴文，仅允许朋友回复，等等。

![互动规则设置](https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/user-settings-interaction-policy-1.png)

### 仅本站贴文

有时你只想与同一实例中的人们交流。GoToSocial 通过仅本站可见贴文支持这一点，确保贴文仅保留在你的实例中。（当前，仅本站可见贴文能否使用取决于客户端支持。）

### RSS 源

GoToSocial 允许你选择将个人资料暴露为 RSS 订阅源，这样人们可以订阅你的公开源而不会错过任何贴文。

### 富文本格式化

使用 GoToSocial，你可以使用流行且易用的 Markdown 标记语言来撰写帖子，从而生成丰富的 HTML 贴文，支持引用段落、语法高亮代码块、列表、内嵌链接等。

![markdown 格式化贴文](https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/markdown-post.png)

### 主题与自定义 CSS

用户可以为他们的账户页 [选择多种有趣的主题](https://docs.gotosocial.org/zh-cn/latest/user_guide/settings/#select-theme)，或甚至编写自己的 [自定义 CSS](https://docs.gotosocial.org/zh-cn/latest/user_guide/settings/#custom-css)。

管理员也可以轻松地为用户 [添加自定义主题](https://docs.gotosocial.org/zh-cn/latest/admin/themes/) 供用户选择。

<details>
<summary>显示主题示例</summary>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-blurple-dark.png"/>
  <figcaption>Blurple dark</figcaption>
</figure>
<hr/>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-blurple-light.png"/>
  <figcaption>Blurple light</figcaption>
</figure>
<hr/>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-brutalist-light.png"/>
  <figcaption>Brutalist light</figcaption>
</figure>
<hr/>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-brutalist-dark.png"/>
  <figcaption>Brutalist dark</figcaption>
</figure>
<hr/>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-ecks-pee.png"/>
  <figcaption>Ecks pee</figcaption>
</figure>
<hr/>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-midnight-trip.png"/>
  <figcaption>Midnight trip</figcaption>
</figure>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-moonlight-hunt.png"/>
  <figcaption>Moonlight hunt</figcaption>
</figure>
<hr/>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-rainforest.png"/>
  <figcaption>Rainforest</figcaption>
</figure>
<hr/>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-soft.png"/>
  <figcaption>Soft</figcaption>
</figure>
<hr/>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-solarized-dark.png"/>
  <figcaption>Solarized dark</figcaption>
</figure>
<hr/>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-solarized-light.png"/>
  <figcaption>Solarized light</figcaption>
</figure>
<hr/>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-sunset.png"/>
  <figcaption>Sunset</figcaption>
</figure>
<hr/>
</details>

### 易于运行

GoToSocial 仅需约 250-350MiB 的 RAM，并且只要求极少的 CPU 频率，因此非常适合单板计算机、旧笔记本和每月 5 美元的小 VPS。

![Grafana 图标显示 GoToSocial 堆占用约为 250MB，偶尔飙升至 400MB-500MB。](https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/getting-started-memory-graph.png)

除数据库外无需其他依赖（也可以仅使用 SQLite！）。

只需下载二进制文件和对应资源（或 Docker 镜像），调整配置并运行。

### 隐私+安全功能

- 内置 [Let's Encrypt](https://letsencrypt.org/) 的自动使用 HTTPS 支持。
- 严格执行贴文可见性和屏蔽逻辑。
- 导入与导出允许联合实例列表和拒绝联合实例列表。订阅社区创建的屏蔽列表（类似于用于实例间联合的广告拦截器！）（功能仍在进行中）。
- HTTP 签名认证：GoToSocial 在发送和接收消息时要求 [HTTP 签名](https://datatracker.ietf.org/doc/html/draft-cavage-http-signatures-12)，以确保消息不能被篡改，身份不能被伪造。

### 多种联合模式

GoToSocial 对联合并不采取一刀切的方法。你的实例应该与谁联合应由你决定。

- “屏蔽列表”模式（默认）：发现新实例；屏蔽你不喜欢的实例。
- “允许列表”模式（实验性）；只选择与信任的实例联合。
- “零”联合模式；保持你的服务器私密（尚未实现）。

[查看文档了解更多信息](https://docs.gotosocial.org/zh-cn/latest/admin/federation_modes)。

### OIDC 集成

GoToSocial 支持 [OpenID Connect (OIDC)](https://openid.net/connect/) 身份提供商，这意味着你可以将其与现有的用户管理服务（如 [Auth0](https://auth0.com/)、[Gitlab](https://docs.gitlab.com/ee/integration/openid_connect_provider.html) 等）集成，或者部署你自己的 OIDC 服务并与之相连（我们推荐使用 [Dex](https://dexidp.io/)）。

### 后端优先设计

与其他联邦宇宙服务端项目不同，GoToSocial 不附带集成的客户端前端（例如，网页端应用）。

相反，与 Matrix.org 的 [Synapse](https://github.com/matrix-org/synapse) 项目类似，它提供了一个相对通用的后端服务器实现，一些用于展示账户和贴文的美观的页面，以及一个[具有完善文档的 API](https://docs.gotosocial.org/zh-cn/latest/api/swagger/)。

在该 API 基础上，GoToSocial 鼓励开发者构建任何他们想要的前端实现或移动应用，无论它们是类似于 Tumblr、Facebook、Twitter，还是完全不同的东西。

---

## 已知问题

由于 GoToSocial 仍处于测试阶段，存在很多错误。我们使用 [GitHub issues](https://github.com/superseriousbusiness/gotosocial/issues?q=is%3Aissue+is%3Aopen+label%3Abug) 跟踪这些问题。

由于每个 ActivityPub 服务端实现对协议的解释略有不同，有些服务端尚未与 GoToSocial 正常联合。我们在 [这个项目](https://github.com/superseriousbusiness/gotosocial/projects/4) 中跟踪这些问题。最终，我们希望确保任何可以与 Mastodon 正确联合的 ActivityPub 实现也能够与 GoToSocial 联合。

---

## 安装 GoToSocial

查看我们的 [入门文档](https://docs.gotosocial.org/zh-cn/latest/getting_started/)，并浏览我们的 [发布页面](https://github.com/superseriousbusiness/gotosocial/releases)。

<!--releases-start-->
### 支持的平台

虽然我们尽力支持合理数量的架构和操作系统，但由于库的限制或性能问题，对特定平台的支持有时是不可能实现的。

某些平台不被我们正式支持，但仍*可能*工作，我们无法测试或保证其性能或稳定性。

以下是 GoToSocial 当前针对不同平台的支持状态（如果某个平台未列出，则表示我们尚未检查，因此我们不清楚）：

| 操作系统 | 架构                  | 支持程度                            | 二进制文件 | Docker 容器     |
| ------- | --------------------- | ---------------------------------- | ---------- | --------------- |
| Linux   | x86-64/AMD64 (64位)   | 🟢 完全支持                         | 是         | 是              |
| Linux   | Armv8/ARM64 (64位)    | 🟢 完全支持                         | 是         | 是              |
| FreeBSD | x86-64/AMD64 (64位)   | 🟢 完全支持<sup>[1](#freebsd)</sup> | 是         | 否              |
| Linux   | x86-32/i386 (32位)    | 🟡 部分支持<sup>[2](#32-bit)</sup>  | 是         | 是              |
| Linux   | Armv7/ARM32 (32位)    | 🟡 部分支持<sup>[2](#32-bit)</sup>  | 是         | 是              |
| Linux   | Armv6/ARM32 (32位)    | 🟡 部分支持<sup>[2](#32-bit)</sup>  | 是         | 是              |
| OpenBSD | 任何架构                  | 🔴 不支持<sup>[3](#openbsd)</sup>   | 否         | 否              |

#### FreeBSD

大多数情况下可用，只是在 WASM SQLite 上有一些问题；在 FreeBSD 上安装时请仔细查看发行说明。如果使用 Postgres，则不应出现问题。

#### 32位

GtS 在像 i386 或 Armv6/v7 这样的 32 位系统上表现不佳，这主要是媒体解码性能的问题。

我们不建议在 32 位系统上运行 GtS，但你可以尝试关闭外站媒体处理功能，或使用完全**不受支持、实验性**的 [nowasm](https://docs.gotosocial.org/zh-cn/latest/advanced/builds/nowasm/) 标签自行构建二进制文件。

有关更多指导，请在尝试在 32 位系统上安装时检查发行说明。

#### OpenBSD

由于性能问题（空闲时的高内存占用，在处理媒体时崩溃），此系统被标记为不支持。

虽然我们不支持在 OpenBSD 上运行 GtS，但你可以尝试使用完全**不受支持、实验性**的 [nowasm](https://docs.gotosocial.org/zh-cn/latest/advanced/builds/nowasm/) 标签自行构建二进制文件。

### 稳定版本

我们为二进制构建和 Docker 容器打包稳定版本，这样你就不需要自己从源代码构建。

Docker 镜像 `superseriousbusiness/gotosocial:latest` 始终对应于最新稳定版本。由于此标签经常被覆盖，你可能希望使用 Docker CLI 标志 `--pull always` 确保每次运行此标签时都有最新的镜像，或者也可在使用前手动运行 `docker pull superseriousbusiness/gotosocial:latest`。

### 快照版本

我们还会在每次将代码合并到主分支时进行快照版的构建，因此如果你愿意，可以从主分支的代码运行。

请注意，风险自负！我们会尝试确保主分支正常工作，但不能做出任何保证。如果不确定，请选择稳定版。

#### Docker

要使用 Docker 从主分支运行，请使用 `snapshot` Docker 标签。Docker 镜像 `superseriousbusiness/gotosocial:snapshot` 始终对应主分支上的最新提交。由于此标签经常被覆盖，你可能希望使用 Docker CLI 标志 `--pull always` 确保每次运行此标签时都有最新的镜像，或者也可在使用前手动运行 `docker pull superseriousbusiness/gotosocial:snapshot`。

#### 二进制发布 .tar.gz

要使用二进制发布从主分支运行，请从我们的 [自托管 Minio S3 仓库](https://minio.s3.superseriousbusiness.org/browser/gotosocial-snapshots)下载适合你架构的 .tar.gz 文件。

S3 存储桶中的快照版二进制发布由 Github 提交哈希控制。要获取最新的，请按上次修改时间排序，或者查看 [这里的提交列表](https://github.com/superseriousbusiness/gotosocial/commits/main)，复制最新的 SHA，并在 Minio 控制台过滤器中粘贴。快照二进制发布会在 28 天后过期，以降低我们的托管成本。

### 从源代码构建

有关从源代码构建 GoToSocial 的说明，请参见 [CONTRIBUTING.md](https://github.com/superseriousbusiness/gotosocial/blob/main/docs/locales/zh/repo/CONTRIBUTING.md) 文件。

### 第三方打包

非常感谢那些将时间和精力投入到打包 GoToSocial 的人！

这些包不是由 GoToSocial 维护的，因此请将问题和反馈发往对应的存储库维护者（并考虑向他们捐款！）。

[![打包状态](https://repology.org/badge/vertical-allrepos/gotosocial.svg)](https://repology.org/project/gotosocial/versions)

你还可以通过以下方式部署自己的 GoToSocial 实例：

- [YunoHost 上的 GoToSocial 打包](https://github.com/YunoHost-Apps/gotosocial_ynh)：作者 [OniriCorpe](https://github.com/OniriCorpe)。
- [Ansible Playbook (MASH)](https://github.com/mother-of-all-self-hosting/mash-playbook)：该 Playbook 支持包括 GoToSocial 在内的多项服务。[文档](https://github.com/mother-of-all-self-hosting/mash-playbook/blob/main/docs/services/gotosocial.md)
- [GoToSocial Helm Chart](https://github.com/fSocietySocial/charts/tree/main/charts/gotosocial)：作者 [0hlov3](https://github.com/0hlov3)。

<!--releases-end-->
---

## 参与贡献

你想为 GtS 作出贡献吗？太好了！❤️❤️❤️ 请查看问题页面，看看是否有你想参与的内容，并阅读 [CONTRIBUTING.md](https://github.com/superseriousbusiness/gotosocial/blob/main/docs/locales/zh/repo/CONTRIBUTING.md) 文件以获取指南并配置开发环境。

---

## 联系我们

如果你有问题或反馈，可以[加入我们的 Matrix 空间](https://matrix.to/#/#gotosocial-space:superseriousbusiness.org)，地址是 `#gotosocial-space:superseriousbusiness.org`。这是联系开发人员的最快方式。你也可以发送邮件至 [admin@gotosocial.org](mailto:admin@gotosocial.org)。

对于错误和功能请求，请先查看是否[已有相应问题](https://github.com/superseriousbusiness/gotosocial/issues)，如果没有，可以开一个新问题工单(issue)，或者使用上述渠道提出请求（如果你没有 Github 账户的话）。

---

## 致谢
<!--body-1-end-->

### 库

GoToSocial 使用以下开源库、框架和工具，在此声明并致谢 💕

- [buckket/go-blurhash](https://github.com/buckket/go-blurhash); 用于生成图像模糊哈希。 [GPL-3.0 许可证](https://spdx.org/licenses/GPL-3.0-only.html)。
- [coreos/go-oidc](https://github.com/coreos/go-oidc); OIDC 客户端库。 [Apache-2.0 许可证](https://spdx.org/licenses/Apache-2.0.html)。
- [DmitriyVTitov/size](https://github.com/DmitriyVTitov/size); 运行时模型内存大小计算。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
- Gin:
  - [gin-contrib/cors](https://github.com/gin-contrib/cors); Gin CORS 中间件。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
  - [gin-contrib/gzip](https://github.com/gin-contrib/gzip); Gin gzip 中间件。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
  - [gin-contrib/sessions](https://github.com/gin-contrib/sessions); Gin 会话中间件。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
  - [gin-gonic/gin](https://github.com/gin-gonic/gin); 高速路由引擎。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
- [google/uuid](https://github.com/google/uuid); UUID 生成。 [BSD-3-Clause 许可证](https://spdx.org/licenses/BSD-3-Clause.html)。
- Go-Playground:
  - [go-playground/form](https://github.com/go-playground/form); 表单映射支持。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
  - [go-playground/validator](https://github.com/go-playground/validator); 结构验证。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
- Gorilla:
  - [gorilla/feeds](https://github.com/gorilla/feeds); RSS + Atom 提要生成。 [BSD-2-Clause 许可证](https://spdx.org/licenses/BSD-2-Clause.html)。
  - [gorilla/websocket](https://github.com/gorilla/websocket); WebSocket 连接。 [BSD-2-Clause 许可证](https://spdx.org/licenses/BSD-2-Clause.html)。
- [go-swagger/go-swagger](https://github.com/go-swagger/go-swagger); Swagger OpenAPI 规范生成。 [Apache-2.0 许可证](https://spdx.org/licenses/Apache-2.0.html)。
- gruf:
  - [gruf/go-bytesize](https://codeberg.org/gruf/go-bytesize); 字节大小解析/格式化。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
  - [gruf/go-cache](https://codeberg.org/gruf/go-cache); LRU 和 TTL 缓存。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
  - [gruf/go-debug](https://codeberg.org/gruf/go-debug); 调试构建标记。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
  - [gruf/go-errors](https://codeberg.org/gruf/go-errors); 类似上下文的错误与值包装。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
  - [gruf/go-fastcopy](https://codeberg.org/gruf/go-fastcopy); 高性能 I/O 复制（缓冲池）。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
  - [gruf/go-ffmpreg](https://codeberg.org/gruf/go-ffmpreg); 嵌入式 ffmpeg / ffprobe WASM 二进制文件。 [GPL-3.0 许可证](https://spdx.org/licenses/GPL-3.0-only.html)。
  - [gruf/go-kv](https://codeberg.org/gruf/go-kv); 日志字段格式化。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
  - [gruf/go-list](https://codeberg.org/gruf/go-list); 通用双向链表。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
  - [gruf/go-mutexes](https://codeberg.org/gruf/go-mutexes); 安全互斥锁和互斥图。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
  - [gruf/go-runners](https://codeberg.org/gruf/go-runners); 同步工具。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
  - [gruf/go-sched](https://codeberg.org/gruf/go-sched); 任务调度器。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
  - [gruf/go-storage](https://codeberg.org/gruf/go-storage); 文件存储后端（本地及 s3）。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
  - [gruf/go-structr](https://codeberg.org/gruf/go-structr); 结构缓存+队列及按字段索引。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
- jackc:
  - [jackc/pgconn](https://github.com/jackc/pgconn); Postgres 驱动程序。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
  - [jackc/pgx](https://github.com/jackc/pgx); Postgres 驱动程序及工具包。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
- [KimMachineGun/automemlimit](https://github.com/KimMachineGun/automemlimit); cgroups 内存限制检查。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
- [k3a/html2text](https://github.com/k3a/html2text); HTML 转文本转换。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
- [mcuadros/go-syslog](https://github.com/mcuadros/go-syslog); Syslog 服务器库。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
- [microcosm-cc/bluemonday](https://github.com/microcosm-cc/bluemonday); HTML 用户输入清理。 [BSD-3-Clause 许可证](https://spdx.org/licenses/BSD-3-Clause.html)。
- [miekg/dns](https://github.com/miekg/dns); DNS 工具。 [Go 许可证](https://go.dev/LICENSE)。
- [minio/minio-go](https://github.com/minio/minio-go); S3 客户端 SDK。 [Apache-2.0 许可证](https://spdx.org/licenses/Apache-2.0.html)。
- [mitchellh/mapstructure](https://github.com/mitchellh/mapstructure); Go 接口 => 结构解析。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
- [modernc.org/sqlite](https://gitlab.com/cznic/sqlite); 简明的 SQLite。 [其他许可证](https://gitlab.com/cznic/sqlite/-/blob/master/LICENSE)。
- [mvdan.cc/xurls](https://github.com/mvdan/xurls); URL 解析正则表达式。 [BSD-3-Clause 许可证](https://spdx.org/licenses/BSD-3-Clause.html)。
- [oklog/ulid](https://github.com/oklog/ulid); 顺序友好的数据库 ID 生成。 [Apache-2.0 许可证](https://spdx.org/licenses/Apache-2.0.html)。
- [open-telemetry/opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go); OpenTelemetry API + SDK。 [Apache-2.0 许可证](https://spdx.org/licenses/Apache-2.0.html)。
- spf13:
  - [spf13/cobra](https://github.com/spf13/cobra); 命令行工具。 [Apache-2.0 许可证](https://spdx.org/licenses/Apache-2.0.html)。
  - [spf13/viper](https://github.com/spf13/viper); 配置管理。 [Apache-2.0 许可证](https://spdx.org/licenses/Apache-2.0.html)。
- [stretchr/testify](https://github.com/stretchr/testify); 测试框架。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
- superseriousbusiness:
  - [superseriousbusiness/activity](https://github.com/superseriousbusiness/activity) 从 [go-fed/activity](https://github.com/go-fed/activity) 派生; Golang ActivityPub/ActivityStreams 库。 [BSD-3-Clause 许可证](https://spdx.org/licenses/BSD-3-Clause.html)。
  - [superseriousbusiness/exif-terminator](https://codeberg.org/superseriousbusiness/exif-terminator); EXIF 数据擦除。 [GNU AGPL v3 许可证](https://spdx.org/licenses/AGPL-3.0-or-later.html)。
  - [superseriousbusiness/httpsig](https://github.com/superseriousbusiness/httpsig) 从 [go-fed/httpsig](https://github.com/go-fed/httpsig) 派生; 安全 HTTP 签名库。 [BSD-3-Clause 许可证](https://spdx.org/licenses/BSD-3-Clause.html)。
  - [superseriousbusiness/oauth2](https://github.com/superseriousbusiness/oauth2) 从 [go-oauth2/oauth2](https://github.com/go-oauth2/oauth2) 派生; OAuth 服务器框架和令牌处理。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
- [tdewolff/minify](https://github.com/tdewolff/minify); Markdown 帖文的 HTML 压缩。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
- [uber-go/automaxprocs](https://github.com/uber-go/automaxprocs); GOMAXPROCS 自动化。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
- [ulule/limiter](https://github.com/ulule/limiter); http 流量限制中间件。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
- [uptrace/bun](https://github.com/uptrace/bun); 数据库 ORM。 [BSD-2-Clause 许可证](https://spdx.org/licenses/BSD-2-Clause.html)。
- [wagslane/go-password-validator](https://github.com/wagslane/go-password-validator); 密码强度验证。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。
- [yuin/goldmark](https://github.com/yuin/goldmark); Markdown 解析器。 [MIT 许可证](https://spdx.org/licenses/MIT.html)。

<!--body-2-start-->
### 图像归属与许可

树懒标志由 [Anna Abramek](https://abramek.art/) 设计。

<a rel="license" href="http://creativecommons.org/licenses/by-sa/4.0/"><img alt="Creative Commons License" style="border-width:0" src="https://i.creativecommons.org/l/by-sa/4.0/88x31.png" /></a><br />GoToSocial 的树懒吉祥物采用 <a rel="license" href="http://creativecommons.org/licenses/by-sa/4.0/">知识共享署名-相同方式共享 4.0 国际许可协议</a>。

该许可具体适用于以下存储库内的文件和子目录：

- [树懒标志 png](https://github.com/superseriousbusiness/gotosocial/blob/main/web/assets/logo.png)
- [树懒标志 webp](https://github.com/superseriousbusiness/gotosocial/blob/main/web/assets/logo.webp)
- [树懒标志 svg](https://github.com/superseriousbusiness/gotosocial/blob/main/web/assets/logo.svg)
- [所有默认头像](https://github.com/superseriousbusiness/gotosocial/blob/main/web/assets/default_avatars)

根据许可证条款，你可以：

- 分享 — 在任何媒介或格式中复制、传播上述材料。
- 演绎 — 混合、转换与再创作上述材料，并用于任何目的，包括商业用途。

### 团队成员

按字母顺序（... 和气味顺序）排列：

- daenney
- f0x \[[通过 liberapay 捐赠](https://liberapay.com/f0x)\]
- kim \[在 @ [codeberg](https://codeberg.org/gruf) 查看我的代码, 或在 @ [@kim](https://k.iim.gay/@kim) 找到我\]
- tobi \[[通过 liberapay 捐赠](https://liberapay.com/GoToSocial/)\]
- vyr

### 特别鸣谢

特别感谢来自 [go-fed](https://github.com/go-fed/activity) 的 CJ：没有你的工作，GoToSocial 不可能实现。

感谢所有使用 GtS 的人，包括提交问题的，提出改进建议的，提供资金支持的，以及以其他方式支持或鼓励该项目的人！

---

## 赞助与资金支持

**有关企业赞助的更新：我们欢迎与符合我们价值观的组织进行赞助合作；请参阅以下条件**

### 众筹

![open collective 标准树懒 徽章](https://opencollective.com/gotosocial/tiers/standard-sloth/badge.svg?label=Standard%20Sloth&color=brightgreen) ![open collective 稳定树懒 徽章](https://opencollective.com/gotosocial/tiers/stable-sloth/badge.svg?label=Stable%20Sloth&color=green) ![open collective 特别树懒 徽章](https://opencollective.com/gotosocial/tiers/special-sloth/badge.svg?label=Special%20Sloth&color=yellowgreen) ![open collective 糖果树懒 徽章](https://opencollective.com/gotosocial/tiers/sugar-sloth/badge.svg?label=Sugar%20Sloth&color=blue)

如果你希望为 GoToSocial 捐款以支持开发，[你可以通过我们的 OpenCollective 捐助](https://opencollective.com/gotosocial#support)！

![LiberaPay 赞助人](https://img.shields.io/liberapay/patrons/GoToSocial.svg?logo=liberapay) ![通过 LiberaPay 接收捐赠](https://img.shields.io/liberapay/receives/GoToSocial.svg?logo=liberapay)

如果你喜欢通过 LiberaPay 赞助，我们也有一个 LiberaPay 帐户！你可以在[这里找到我们](https://liberapay.com/GoToSocial/)。

通过我们 OpenCollective 和 Liberapay 账户的众筹捐款将用于支付核心团队的工资、服务器成本以及 GtS 的艺术、设计等其他开支。

💕 🦥 💕 谢谢你们！

### 企业赞助

GoToSocial 欢迎与符合我们价值观的组织进行合作。在此对您的支持表示感谢，我们将在存储库和文档中展示你的 Logo、网站及简短的标语。赞助有以下限制：

1. GoToSocial 的项目方向始终由核心团队完全掌控，永远不会受到企业赞助的支配或影响。这是不可商量的。当然，企业同样可以像任何其他用户一样建议/请求功能，但不会获得特殊待遇。

2. 企业赞助取决于你的组织是否符合我们团队的伦理准则。这不是一套具体的规则，而是“你的公司是否造成了伤害？”的问题。例如，国防行业的不需要申请，因为答案显然是肯定的！

如果在阅读后您仍有兴趣赞助我们，那太好了！请通过 admin@gotosocial.org 与我们联系以进一步讨论 :)

### NLnet

<img src="https://nlnet.nl/logo/NGI/NGIZero-green.hex.svg" width="75" alt="NGIZero logo"/>

结合以上众筹来源，2023 年 GoToSocial Alpha 阶段的开发得到了 [NGI0 Entrust Fund](https://nlnet.nl/entrust/) 旗下的 [NLnet](https://nlnet.nl/) 提供的 50,000 欧元资助。详情请见[此处](https://nlnet.nl/project/GoToSocial/#ack)。成功的资助申请存档在[此处](https://github.com/superseriousbusiness/gotosocial/blob/main/archive/nlnet/2022-next-generation-internet-zero.md)。

2024 年 GoToSocial Beta 阶段的开发将从 [NGI0 Entrust Fund](https://nlnet.nl/entrust/) 旗下的 [NLnet](https://nlnet.nl/) 那里再获得 50,000 欧元的资助。

---

## 许可

![GNU AGPL 徽标](https://www.gnu.org/graphics/agplv3-155x51.png)

GoToSocial 是自由软件，采用 [GNU AGPL v3 许可](https://github.com/superseriousbusiness/gotosocial/blob/main/LICENSE)。我们鼓励你对代码进行派生和修改，进行各种实验。

有关 AGPL 和 GPL 许可之间的区别，请参阅[这里](https://www.gnu.org/licenses/why-affero-gpl.html)，关于 GPL 许可（包括 AGPL）的常见问题解答，请参阅[这里](https://www.gnu.org/licenses/gpl-faq.html)。

如果你修改了 GoToSocial 的源码，并以网络可访问的方式运行修改后的代码，你*必须*根据许可的指引提供你对源码的修改副本：

> 如果你修改了程序，并且你的修改版本支持通过计算机网络与用户进行远程交互，你的版本必须显著地向所有这些用户提供获得你的版本对应源码的机会，方式需为通过网络服务器以不收费的方式，或通过某种标准或习惯方式提供以便于复制软件。

版权所有 (C) 全体 GoToSocial 开发者

<!--I'm adding this here to take the crown of having the 1000th commit ~ kim-->
<!--body-2-end-->
