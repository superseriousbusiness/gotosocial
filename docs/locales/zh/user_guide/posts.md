# 贴文

## 隐私设置

GoToSocial 为贴文提供 Mastodon 风格的隐私设置。从最私密到最不私密的顺序是：

* 私信
* 互关可见
* 私密/仅粉丝可见
* 不列出
* 公开

无论为贴文选择哪种隐私设置，GoToSocial 都会尽力确保你的贴文不会出现在你已屏蔽的实例或你直接屏蔽的用户面前。

与一些其他联邦宇宙服务端实现不同，GoToSocial 对新账户使用 `不列出` 作为默认的贴文设置，而不是使用 `公开`。我们的理念是，将某条贴文公开始终应该是一个明确作出的决定，而不是默认选择。

请注意，尽管 GoToSocial 非常严格地遵循这些隐私设置，但其他服务端实现不一定可靠：联合网络中存在不良行为者。与任何社交媒体一样，你应该仔细考虑你发布的内容及其受众。

### 私信

`私信` 可见性的贴文只会显示给贴文作者和贴文中提到的用户。以下是一个示例：

```text
嘿 @whoever@example.org，这是一条私信！只有咱们能看到！
```

如果这条消息是由 `@someone@server.com` 写的，那么只有 `@whoever@example.org` 和 `@someone@server.com` 能看到。

顾名思义，`私信` 贴文用于你希望与一个或多个特定人交流的情况。

然而，`私信` 贴文并不是替代端到端加密消息（如 [Signal](https://signal.org/) 和 [Matrix](https://matrix.org/)）的合适选择。如果你要直接交流，但不是传递敏感信息，那么 `私信` 贴文是可以胜任的。如果你需要进行敏感且安全的交流，请使用其他工具！

私信贴文可以被点赞，但不能被转发。

私信贴文在你的 GoToSocial 实例上无法通过网页 URL 访问。

### 互关可见

!!! warning "警告"
    目前暂时无法将帖文可见性设为“互关可见”。

`互关可见` 的贴文只会显示给贴文作者和与作者*互相关注*的人。换句话说，只有在满足两个条件时，其他人才能看到：

1. 其他账户关注贴文作者。
2. 贴文作者也关注其他账户。

这在你希望只有朋友能看到某些内容时很有用。

互关可见贴文可以被点赞，但不能被转发。

互关可见贴文在你的 GoToSocial 实例上无法通过网页 URL 访问。

### 私密/仅粉丝可见

`仅粉丝` 贴文只对贴文作者和关注贴文作者的人可见。与 `互关可见` 类似，但只需满足第一个条件；贴文作者不需要回关其他账户。

这在你想向粉丝发布公告，或分享一些比 `互关可见` 稍微不那么私密的内容时很有用。

私密/仅粉丝可见贴文可以被点赞，但不能被转发。

私密/仅粉丝可见贴文在你的 GoToSocial 实例上无法通过网页 URL 访问。

### 不列出

`不列出`（有时称为 `未锁定/悄悄公开`）的贴文是半公开的。它们会被发送给关注你的人，并可以转发到未关注你的人的时间线中，但不会出现在跨站或本站时间线上，也不会出现在你的公开资料页中。

不列出贴文适用于你希望某个贴文传播，但不想被所有人立即看到的情况。也可以用于发布相对公开的贴文，同时不让跨站/本站时间线被占用。

不列出贴文可以被点赞，也可以被转发。

与 Mastodon 不同，不列出贴文在你的 GoToSocial 实例上无法通过网页 URL 访问！

### 公开

`公开` 可见性的贴文是*完全*公开的。它们可以通过网络看到，出现在本地和联合时间线上，并且完全可以被转发。`公开` 是将贴文广泛传播和易于分发的终极设置，适用于你希望某些内容可被广泛访问的情况。

公开贴文可以被点赞，也可以被转发。

**公开贴文可以在你的 GoToSocial 实例上通过网页 URL 访问！**

## 输入类型

GoToSocial 当前接受两种不同类型的贴文（以及用户简介）输入。你可以在[用户设置页面](./settings.md)在这两种类型之间进行选择。它们分别是：

* `plain`
* `Markdown`

纯文本(`plain`)是默认的发帖方式：GtS 接受一些简单的文本，通过解析链接和提及等将其转化为优雅的 HTML。如果你习惯于使用 Mastodon 或 Twitter 或大多数其他社交媒体平台，这种发帖方式会让你一见如故。

Markdown 是一种更复杂的组织文本的方式，在如何解析和格式化文本方面给你更多的控制权。

GoToSocial 支持[基本 Markdown 语法](https://www.markdownguide.org/basic-syntax)，以及部分[扩展 Markdown 语法](https://www.markdownguide.org/extended-syntax/)，包括独立代码块、脚注、删除线、下标、上标和自动 URL 链接。

你还可以在你的 markdown 中包含基本 HTML 片段！

关于 Markdown 的更多信息，请参阅[Markdown 指南](https://www.markdownguide.org/)。

关于 Markdown 语法的速查表，请参阅[Markdown 速查表](https://www.markdownguide.org/cheat-sheet)。

## 媒体附件

GoToSocial 允许你在贴文中附加媒体文件，大多数客户端会在贴文底部以画廊视图呈现这些文件。默认情况下，你可以向贴文附加 6 个媒体文件，但这可能会根据你使用的客户端和实例配置而有所不同。

目前支持以下文件类型：

- image/jpeg
- image/gif
- image/png
- image/webp
- video/mp4（大多数类型）

默认情况下，上传媒体的大小限制为 40MB，但这可能会因实例配置而有所不同。

### 图片描述（alt 文本）

当你在贴文中附加图片或视频等媒体时，大多数客户端会提供选项，让你为图片或视频的内容撰写描述。这个描述将作为所有用户查看媒体时的 alt 文本出现。这对所有人，尤其是对盲人或视力部分受损的人来说有用。如果没有图片描述，对方可能难以理解媒体中包含的内容以及为何你将其附加到特定贴文中。

撰写好的图片描述可能很难，但这样做非常值得！

> 图片描述是一种表示关心的行为，也是无障碍的基本组成部分。没有它们，内容对盲人/视力低下的人来说将完全不可用。通过撰写图片描述，我们展现了对跨残疾团结及运动团结的支持。

-- Alex Chen，[如何撰写图片描述](https://uxdesign.cc/how-to-write-an-image-description-2f30d3bf5546)。

### Exif 数据

当照片或视频拍摄时，大多数传统相机和手机相机会将 [Exif 数据标签](https://en.wikipedia.org/wiki/Exif) 编码为结果媒体的元数据。此 Exif 数据包含如下内容：

- 相机的品牌和型号。
- 图片或视频的颜色和像素信息。
- 图片或视频的尺寸和方向。
- 日期和时间信息。
- 位置信息（如果启用）。

一般来说，这些 Exif 数据点用于摄影师帮助整理他们自己的图片。然而，遗憾的是，它们也带来了[隐私和安全影响](https://en.wikipedia.org/wiki/Exif#Privacy_and_security)，特别是在涉及位置信息时。如果你曾在网上平台（如 Facebook）发布图片，你可能会想知道 Facebook 是如何知道图片的拍摄地点和时间的；这很大程度上归因于 Exif 数据中嵌入的位置信息和时间戳，Facebook 从中读取图片信息，以组装一条“你曾去过的地方”的时间线。

为了避免泄漏你的位置信息，GoToSocial 努力在上传媒体时通过清零 Exif 数据点移除 Exif 信息。

!!! danger "危险"
    为了方便和保护隐私，GoToSocial 在上传图片文件时会自动移除 Exif 标签。然而，**无法自动移除 mp4 视频的 Exif 数据**（参见 [#2577](https://github.com/superseriousbusiness/gotosocial/issues/2577)）。
    
    在你将视频上传至 GoToSocial 之前，建议确保该视频的 Exif 数据标签已经被移除。你可以在线找到多种工具和服务来做到这一点。
    
    为防止 Exif 位置信息在一开始被写入图片或视频中，你还可以关闭设备摄像头应用中的位置标记（通常称为地理标记）。

!!! tip "提示"
    即使你在上传图片或视频之前已完全移除所有 Exif 元数据，恶意用户仍然可以通过媒体本身的内容推断出你的位置信息。
    
    如果你属于在生产中有保密需要的组织，或正在被跟踪或监视，你可能需要考虑不要发布任何可能含有你位置线索的媒体。

## 格式化

当贴文以 `plain`(纯文本) 格式提交时，GoToSocial 会自动进行一些整理和格式化，将其转换为 HTML，如下所述。

### 空格

任何开头或结尾的空格和换行都会从贴文中去除。因此，例如：

```text


这个贴文以换行开头
```

将变为：

```text
这个贴文以换行开头
```

### 包裹

整个贴文将被 `<p></p>` 包裹。

因此以下文本：

```text
你好，这是一条很短的贴文！
```

将变为：

```html
<p>你好，这是一条很短的贴文！</p>
```

### 换行

任何换行符都将被替换为 `<br />`

继续上述示例：

```text
你好，这是一条很短的贴文！

这是另一行。
```

将变为：

```html
<p>你好，这是一条很短的贴文！<br /><br />这是另一行。</p>
```

### 链接

任何可识别的链接将在文本中被缩短并转换为适当的超链接，还会添加一些其他属性。

例如：

```text
这里是某个链接：https://example.org/some/link/address
```

将变为：

```html
这里是某个链接：<a href="https://example.org/some/link/address" rel="nofollow" rel="noreferrer" rel="noopener">example.org/some/link/address</a>
```

呈现为：

> 这里是某个链接：[example.org/some/link/address](https://example.org/some/link/address)

注意这仅对 `http` 和 `https` 链接有效；其他协议不支持。

### 提及

你可以通过以下方式提及其他账户：

> @some_account@example.org

在这个例子中，`some_account` 是你要提及的账户的用户名，`example.org` 是托管他们账户的域名。

被提及的账户将收到你提到他们的通知，并能够看到提及他们的贴文。

提及的格式类似于链接，所以：

```text
嗨 @some_account@example.org 最近怎么样？
```

将变为：

```html
嗨 <span class="h-card"><a href="https://example.org/@some_account" class="u-url mention">@<span>some_account</span></a></span> 最近怎么样？
```

呈现为：

> 嗨 <span class="h-card"><a href="https://example.org/@some_account" class="u-url mention">@<span>some_account</span></a></span> 最近怎么样？

当提及本站账户（即你的实例上的账户）时，提及第二部分是不必要的。如果在你的实例上有一个名为 `local_account_person` 的账户，你可以通过写：

```text
嘿 @local_account_person 你是我的网上邻居
```

变为：

```html
嘿 <span class="h-card"><a href="https://my.instance.org/@local_account_person" class="u-url mention">@<span>local_account_person</span></a></span> 你是我的网上邻居
```

呈现为：

> 嘿 <span class="h-card"><a href="https://my.instance.org/@local_account_person" class="u-url mention">@<span>local_account_person</span></a></span> 你是我的网上邻居

### 话题标签

你可以在贴文中使用一个或多个话题标签来指示贴文主题，并允许贴文与其他使用相同话题标签的贴文被归入同一分组，以帮助你的贴文被他人发现。

大多数 ActivityPub 服务端实现，如 Mastodon 等，只会通过它们使用的话题标签对**公开**贴文进行分组，但这并不是绝对的。一般来说，最好只对那些你希望能比其他情况下更广泛传播的公开可见贴文使用话题标签。这方面的一个好例子是 `#introduction` 话题标签，通常用于新账户想要向联邦宇宙介绍自己时使用！

在贴文中包含话题标签的方式类似于大多数其他社交媒体软件：只需在你想用作话题标签的词前加上 `#` 符号。

一些示例：

* `#introduction`
* `#Mosstodon`
* `#LichenSubscribe`

在 GoToSocial 中，话题标签不区分大小写，因此无论你在书写话题标签时使用大写、小写或两者混合，都会被视为相同的话题标签。例如，`#Introduction` 和 `#introduction` 会被视为完全相同。

出于可访问性原因，在书写话题标签时，考虑使用大驼峰式（即每个单词的首字母大写）是更好的。换句话说：要把 `#thisisahashtag` 替换为 `#ThisIsAHashtag`。这样不仅视觉上更易读，屏幕阅读器也更容易朗读。

你可以在 GoToSocial 贴文中包含任意数量的话题标签，而且每个话题标签的长度限制为 100 个字符。

!!! tip "提示"
    要结束一个话题标签，你只需在话题标签名后输入空格。例如，在文本 `这道 #鸡汤 十分美味` 中，话题标签由空格终止，因此 `#鸡汤` 成为话题标签。但是，你也可以使用管道字符 `|`，或使用 Unicode 字符 `\u200B` （零宽不换行空格）或 `\uFEFF` （零宽空格）,来创建“词语片段”话题标签。例如，在 `这道 #鸡|汤 十分美味` 中，只有 `#鸡` 成为话题标签。同理，对于文本 `这道 #鸡​汤 十分美味` （`鸡` 和 `汤` 之间有一个零宽空格），只有 `#鸡` 成为话题标签。有关零宽空格的更多信息，参见：https://en.wikipedia.org/wiki/Zero-width_space。

## 输入净化

为了不传播脚本、漏洞以及不稳定的 HTML，GoToSocial 执行以下类型的输入净化：

`plain` 输入类型：

* 在解析前，会完全移除贴文正文和内容警告字段中的已有 HTML。
* 在解析后，所有生成的 HTML 都会通过清理器处理以移除有害元素。

`Markdown` 输入类型：

* 在解析前，会完全移除内容警告字段中的已有 HTML。
* 在解析前，贴文正文中的现有 HTML 会通过清理器处理以移除有害元素。
* 在解析后，所有生成的 HTML 都会通过清理器处理以移除有害元素。

GoToSocial 使用 [bluemonday](https://github.com/microcosm-cc/bluemonday) 进行 HTML 清理。
