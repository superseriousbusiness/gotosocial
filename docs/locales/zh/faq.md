# 常见问题

## 用户界面在哪？

GoToSocial 的大部分内容只是一个裸服务端，用于通过第三方应用程序进行使用。可通过 [Pinafore](https://pinafore.social/) 在浏览器使用，可通过 [Tusky](https://tusky.app/) 在 Android 使用，可通过 [Feditext](https://github.com/feditext/feditext) 在 iOS、iPadOS 和 macOS 使用。这些应用程序兼容性最好。任何由 Mastodon API 提供的实例功能都应该可以工作，除非它们是 GoToSocial 尚不具备的功能。永久链接和个账户页是通过 GoToSocial 直接提供的，设置面板也是如此，但大多数交互都是通过应用程序完成的。

## 为什么我的贴文没有显示在我的账户页面上？

与 Mastodon 不同，GoToSocial 的默认贴文可见性是“不列出”。如果你希望某个内容在个人资料页面上可见，贴文必须设为公开可见。

## 为什么我的贴文没有在其他实例上显示？

首先检查上面提到的可见性。TODO: 解释如何调试常见的联合问题

## 为什么我频繁收到 HTTP 429 错误响应？

GoToSocial 默认配置了基于 IP 的[限流规则](./api/ratelimiting.md)，但在某些情况下无法准确识别外部 IP，会将所有连接视为来自相同位置。在这些情况下，需要禁用或重新配置限流。

## 为什么我频繁收到 HTTP 503 错误响应？

当你的实例负载过重且请求被限制时，会返回代码 503。可以根据需要调整此行为，或完全关闭，请参见[此处](./api/throttling.md)。

## 我总是收到 400 Bad Request 错误，且已经按照错误信息中的建议操作。该怎么办？

请确认 `host` 配置与 GoToSocial 所服务的域名（用户用来访问服务器的域名）匹配。

## 我在服务器日志中不断看到 'dial within blocked / reserved IP range'，无法从我的实例连接到一些实例，我该怎么办？

外站实例的 IP 地址可能位于 GoToSocial 出于安全原因而硬编码屏蔽的“特殊用途”IP 范围内。如果需要，你可以在配置文件中覆盖此设置。查看[HTTP 客户端文档](./configuration/httpclient.md)，并仔细阅读其中的警告！如果添加了明确的 IP 允许条目，需要重启 GoToSocial 实例以使配置生效。

## 我已部署实例并登录到客户端，但时间线为空，这是怎么回事？

要查看贴文，你需要开始关注其他人！一旦你关注了几个人，并且他们发布或转发了内容，你就会在时间线上看到这些内容。目前，GoToSocial 没有“回填”贴文的方法，即尚不能从其他实例获取之前的贴文，所以你只能看到你关注的人的新贴文。如果你想与他们的旧贴文互动，可以从他们的账户页中复制贴文链接，并将其粘贴到客户端的搜索栏中。

## 如何在某个实例上注册？

我们在 v0.16.0 中引入了注册流程。你想注册的实例必须手动启用注册功能，具体详见[此处](./admin/signups.md)。

## 为什么还在 Beta 阶段？

查看[当前 bug 列表](https://codeberg.org/superseriousbusiness/gotosocial/issues?q=is%3Aissue+is%3Aopen+label%3Abug)和[路线图](https://codeberg.org/superseriousbusiness/gotosocial/src/branch/main/docs/locales/zh/repo/ROADMAP.md)以获取更详细的信息。
