# 术语表

本文档描述了有关联合的一些常用术语。

## `ActivityPub`

一种基于 ActivityStreams 数据格式的去中心化社交网络协议。参见 [这里](https://www.w3.org/TR/activitypub/)。

GoToSocial 在 与其它 GtS 服务器和其它联邦宇宙服务器（如 Mastodon, Pixelfed 等）通信时使用 ActivityPub 协议。

## `ActivityStreams`

使用 JSON 表示潜在活动和已完成活动的模型/数据格式。参见 [这里](https://www.w3.org/TR/activitystreams-core/)。

GoToSocial 使用 ActivityStreams 数据模型与其他服务器进行 ActivityPub 通信。

## `Actor` (行为体)

Actor 是一个可以执行某些活动（例如关注、点赞、创建贴文、转发等）的 ActivityStreams 对象。参见 [这里](https://www.w3.org/TR/activitypub/#actors)。

在 GoToSocial 中，每个账号/用户都是一个 actor。

## `Dereference` (解引用)

“解引用”一个贴文或账户意味着向托管该贴文或账户的服务器发出 HTTP 请求，以获取其 ActivityStreams 表示。

GoToSocial 对外站服务器解引用贴文和账户，以将它们转换为 GoToSocial 可以理解和处理的模型。

以下是一些示例的详细说明：

假设有人在 ActivityPub 服务器上搜索用户名 `@tobi@goblin.technology`。

他们的服务器会在 `goblin.technology` 上通过以下 URL 进行 webfinger 查询：

```text
https://goblin.technology/.well-known/webfinger?resource=acct:tobi@goblin.technology
```

`goblin.technology` 服务器会返回一些 JSON 作为响应，类似于：

```json
{
  "subject": "acct:tobi@goblin.technology",
  "aliases": [
    "https://goblin.technology/users/tobi",
    "https://goblin.technology/@tobi"
  ],
  "links": [
    {
      "rel": "http://webfinger.net/rel/profile-page",
      "type": "text/html",
      "href": "https://goblin.technology/@tobi"
    },
    {
      "rel": "self",
      "type": "application/activity+json",
      "href": "https://goblin.technology/users/tobi"
    }
  ]
}
```

在链接部分，请求服务器会寻找类型为 `application/activity+json` 的链接，这表示用户的 ActivityStreams 表示。在此情况下，URL 是：

```text
https://goblin.technology/users/tobi
```

上述 URL 是 `tobi` 在 `goblin.technology` 实例上的 用户/行为体 的 activitypub 表示的*引用*。它之所以被称为引用，是因为它不包含关于该用户的所有信息，只是信息所在位置的参考点。

现在，请求服务器将向该 URL 发送请求，以获得 `@tobi@goblin.technology` 的更完整表示，以符合 ActivityPub 规范。换句话说，服务器现在通过一个*引用*来获取它所引用的内容。这使得它*不再是一个引用*，因此称为*解引用*。

作为类比，考虑在书的目录中查找某些内容时的情况：首先你获得你感兴趣的材料所在的页码，这是一个引用。然后你翻到引用的页面查看内容，这就是解引用。
