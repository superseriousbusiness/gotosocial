# 帖文及其属性

## 话题标签

GoToSocial 用户可以在贴文中包含话题标签，用于向其他实例表明该用户希望将其贴文与其他使用相同话题标签的贴文加入同一分组，以便于发现。

GoToSocial 默认期望只有公开地址的贴文会按话题标签分组，这与其他 ActivityPub 服务器实现一致。

为了实现话题标签的联合，GoToSocial 在对象的 `tag` 属性中使用被广泛采用的 [ActivityStreams `Hashtag` 类型扩展](https://www.w3.org/wiki/Activity_Streams_extensions#as:Hashtag_type)。

以下是一条外发信息中的 `tag` 属性示例，包含自定义表情和一个话题标签：

```json
"tag": [
  {
    "icon": {
      "mediaType": "image/png",
      "type": "Image",
      "url": "https://example.org/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png"
    },
    "id": "https://example.org/emoji/01F8MH9H8E4VG3KDYJR9EGPXCQ",
    "name": ":rainbow:",
    "type": "Emoji",
    "updated": "2021-09-20T10:40:37Z"
  },
  {
    "href": "https://example.org/tags/welcome",
    "name": "#welcome",
    "type": "Hashtag"
  }
]
```

当仅有一个话题标签时，`tag` 属性将是一个对象而非数组，如下所示：

```json
"tag": {
  "href": "https://example.org/tags/welcome",
  "name": "#welcome",
  "type": "Hashtag"
}
```

### 话题标签的 `href` 属性

GoToSocial 外发话题标签提供的 `href` URL 指向一个提供 `text/html` 的网页。

GoToSocial 对给定 `text/html` 的内容不做任何保证，外站不应该将 URL 解释为规范的 ActivityPub ID/URI 属性。`href` URL 仅作为可能包含该话题标签更多信息的一个端点。

## 表情符号（Emoji）

GoToSocial 使用 `http://joinmastodon.org/ns#Emoji` 类型，以允许用户在贴文中添加自定义表情符号。

例如:

```json
{
  "@context": [
    "https://gotosocial.org/ns",
    "https://www.w3.org/ns/activitystreams",
    {
      "Emoji": "toot:Emoji",
      "sensitive": "as:sensitive",
      "toot": "http://joinmastodon.org/ns#"
    }
  ],
  "type": "Note",
  "content": "<p>这里有个臭烘烘的东西 -> :shocked_pikachu:</p>",
  [...],
  "tag": {
    "icon": {
      "mediaType": "image/png",
      "type": "Image",
      "url": "https://example.org/fileserver/01BPSX2MKCRVMD4YN4D71G9CP5/emoji/original/01AZY1Y5YQD6TREB5W50HGTCSZ.png"
    },
    "id": "https://example.org/emoji/01AZY1Y5YQD6TREB5W50HGTCSZ",
    "name": ":shocked_pikachu:",
    "type": "Emoji",
    "updated": "2022-11-17T11:36:05Z"
  }
  [...]
}
```

上述 `Note` 的 `content` 中的文本 `:shocked_pikachu:` 应当被客户端替换为表情符号图片的小型（内联）版本，在渲染 `Note` 时一并向用户展示。

表情符号的 `updated` 和 `icon.url` 属性可被外站实例用于判断它们对 GoToSocial 表情符号图片的表示是否是最新版本。必要时也可以在其 `id` URI 间接引用 `Emoji`，以便外站对照检查它们缓存的表情符号元数据是否未最新版本。

默认情况下，GoToSocial 对可以上传和发送的表情符号图片的大小设置了 50kb 的限制，并对可以联合并传入的表情符号图片大小设置了 100kb 的限制，但这两项设置都可由用户配置。

GoToSocial 可以发送和接收类型为 `image/png`、`image/jpeg`、`image/gif` 和 `image/webp` 的表情符号图片。

!!! info "附注"
    请注意，`tag` 属性可以是对象的数组，也可以是单个对象。

### `null` / 空 `id` 属性

一些服务端软件，如 Akkoma，将表情符号包含为贴文中的[匿名对象](https://www.w3.org/TR/activitypub/#obj-id)。也就是说，它们将 `id` 属性设置为 `null`，以表明该表情符号不能在任何特定的端点被间接引用。

在接收到这样的表情符号时，GoToSocial 会在数据库中为该表情符号生成一个伪 id，格式为 `https://[host]/dummy_emoji_path?shortcode=[shortcode]`，例如，`https://example.org/dummy_emoji_path?shortcode=shocked_pikachu`。

## 提及

GoToSocial 用户可以在贴文中使用 `@[用户名]@[域名]` 格式提及其他用户。例如，如果一个 GoToSocial 用户想提及实例 `example.org` 上的用户 `someone`，可以在贴文中包含 `@someone@example.org`。

!!! info "提及与活动地址"
    
    提及的表示不仅仅是美观问题，它们也会影响活动的地址。
    
    如果 GoToSocial 用户在 Note 中明确提及另一个用户，该用户的 URI 将始终包含在 Note 的 Create 活动的 `To` 或 `Cc` 属性中。
    
    如果 Note 的可见性为私信（即，不发送给公众或粉丝），每个提及的目标 URI 将位于活动包装的 `To` 属性中。
    
    在所有其他情况下，提及将包含在活动包装的 `Cc` 属性中。

### 外发

当 GoToSocial 用户提及另一个账户时，提及会作为 `tag` 属性中的一个条目包含在外发的联合消息中。

例如，一个 GoToSocial 实例上的用户提及 `@someone@example.org`，外发 Note 的 `tag` 属性可能如下：

```json
"tag": {
  "href": "http://example.org/users/someone",
  "name": "@someone@example.org",
  "type": "Mention"
}
```

如果用户提及同一实例内的本站用户，他们的完整 `name` 仍会被包含。

例如，一个 GoToSocial 用户在域 `some.gotosocial.instance` 上提及同一实例内的用户 `user2`。他们还提及了 `@someone@example.org`。外发 Note 的 `tag` 属性如下：

```json
"tag": [
  {
    "href": "http://example.org/users/someone",
    "name": "@someone@example.org",
    "type": "Mention"
  },
  {
    "href": "http://some.gotosocial.instance/users/user2",
    "name": "@user2@some.gotosocial.instance",
    "type": "Mention"
  }
]
```

为了方便外站，GoToSocial 始终在外发的提及中提供 `href` 和 `name` 属性。GoToSocial 使用的 `href` 属性始终是目标账户的 ActivityPub ID/URI，而不是网页 URL。

### 传入

GoToSocial 尝试以与发送出相同的方式解析传入提及：作为 `tag` 属性中的 `Mention` 类型条目。然而，解析传入提及时，对于必须设置哪些属性的要求会更宽松些。

GoToSocial 更偏好 `href` 属性，它可以是目标的 ActivityPub ID/URI 或网页 URL；如果 `href` 不存在，将使用 `name` 属性作为回退。如果两个属性都不存在，提及将被视为无效并被丢弃。

## 内容、内容映射和语言

与其他 ActivityPub 实现一致，GoToSocial 在 `Objects` 中使用 `content` 和 `contentMap` 字段来推断传入贴文的内容和语言，并在外发贴文中设置内容和语言。

### 外发

如果一个外发 `Object`（通常是 `Note`）有内容，它将以在 `content` 字段中以被字符串化的 HTML 形式给出。

如果 `content` 关联特定用户选择的语言，那么 `Object` 还将设置 `contentMap` 属性为单条目键/值映射，其中键是 BCP47 语言话题标签，值是与 `content` 字段相同的内容。

例如，一篇用英语 (`en`) 写的贴文将如下所示：

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Note",
  "attributedTo": "http://example.org/users/i_p_freely",
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "cc": "http://example.org/users/i_p_freely/followers",
  "id": "http://example.org/users/i_p_freely/statuses/01FF25D5Q0DH7CHD57CTRS6WK0",
  "url": "http://example.org/@i_p_freely/statuses/01FF25D5Q0DH7CHD57CTRS6WK0",
  "published": "2021-11-20T13:32:16Z",
  "content": "<p>This is an example note.</p>",
  "contentMap": {
    "en": "<p>This is an example note.</p>"
  },
  "attachment": [],
  "replies": {...},
  "sensitive": false,
  "summary": "",
  "tag": {...}
}
```

如果贴文有内容，GoToSocial 会始终设置 `content` 字段，但是如果使用的 GoToSocial 版本较旧、用户未设置语言，或设置的语言不是公认的 BCP47 语言标签，则可能不会始终设置 `contentMap` 字段。

### 传入

GoToSocial 在解析传入的 `Object` 时使用 `content` 和 `contentMap` 属性来确定内容并推断该内容的主要语言。它使用以下算法：

#### 仅设置 `content`

仅采用该内容并将语言标记为未知。

#### 同时设置 `content` 和 `contentMap`

在 `contentMap` 中查找键，作为语言标签，要查找的键的值与 `content` 中的 HTML 字符串匹配。

如果找到匹配项，将其用作贴文的语言。

如果未找到匹配项，则保留 `content` 中的内容并将语言标记为未知。

#### 仅设置 `contentMap`

如果 `contentMap` 只有一个条目，则将语言标签和内容(值)作为“主要”语言和内容。

如果 `contentMap` 有多个条目，则无法确定贴文的意图内容和语言，因为映射顺序不可预测。在这种情况下，尝试从 GoToSocial 实例的[配置语言](../configuration/instance.md)中选择与其中一种语言匹配的语言和内容条目。如果无法通过这种方式匹配语言，则从 `contentMap` 中随机选择一个语言和内容条目作为“主要”语言和内容。

!!! note "注意"
    在上述所有情况下，如果推断的语言无法解析为有效的 BCP47 语言话题标签，则语言将回退为未知。

## 互动规则

GoToSocial 在帖文中使用 `interactionPolicy` 属性，以向外站实例描述对于任何给定的帖子，哪些类型的互动在条件允许的情况下可以被原始服务器处理和存储。

有关更多详细信息，请参阅单独的 [互动规则](./interaction_policy.md) 文档。

## 投票

为了联合投票状态，GoToSocial 使用广泛采用的 [ActivityStreams `Question` 类型](https://www.w3.org/TR/activitystreams-vocabulary/#dfn-question)。然而，第一个由 Mastodon 引入和推广的这个类型略微偏离 ActivityStreams 规范。在规范中，Question 类型被标记为 `IntransitiveActivity` 的扩展，此扩展是一个应当不带 `Object` 且所有详细信息应被默认包含的 `Activity` 扩展。但在具体实现中，它作为 `Object` 通过 `Create` 或 `Update` 活动传递。

值得注意的是，虽然 GoToSocial 内部可能将投票视作贴文附件的一种类型，但 ActivityStreams 表示法将贴文和带投票的贴文视为两种不同的 `Object` 类型。贴文以 `Note` 类型联合，投票以 `Question` 类型联合。

GoToSocial 传输（和期望接收）的 `Question` 类型包含所有常见的 `Note` 属性，外加一些附加内容。它们期望以下附加的（伪）JSON：

```json
{
  "@context":[
    {
      // toot:votersCount 扩展，用于添加 votersCount 属性。
      "toot":"http://joinmastodon.org/ns#",
      "votersCount":"toot:votersCount"
    }
  ],

  // oneOf / anyOf 包含投票选项
  // 本身。只有其中之一会被设置，
  // 其中 "oneOf" 表示单选投票，
  // "anyOf" 表示多选投票。
  //
  // 任一属性都包含一个 “Notes” 数组，
  // 特殊的是它们包含一个 “name” 且未设置
  // “content”，其中 “name” 代表实际
  // 投票选项字符串。此外，它们包含
  // 一个 “Collection” 类型的 “replies” 属性，
  // 通过 “totalItems” 表示每个投票选项当前已知的投票数。
  "oneOf": [ // 或 "anyOf"
    {
      "type": "Note",
      "name": "选项 1",
      "replies": {
        "type": "Collection",
        "totalItems": 0
      }
    },
    {
      "type": "Note",
      "name": "选项 2",
      "replies": {
        "type": "Collection",
        "totalItems": 0
      }
    }
  ],

  // endTime 指示此投票将何时结束。
  // 某些服务器实现支持永不结束的投票，
  // 或使用 “closed” 来暗示 “endTime”，因此该项可能不会总是被设置。
  "endTime": "2023-01-01T20:04:45Z",

  // closed 指示此投票结束的时间。
  // 在来到此时间之前，该项将不会被设置。
  "closed": "2023-01-01T20:04:45Z",

  // votersCount 表示参与者的总数，
  // 这在多选投票的情况下很有用。
  "votersCount": 10
}
```

### 外发

你可以期望从 GoToSocial 接收到一个 `Question` 形式的投票，投票在 `Create` 或 `Update` 活动中作为对象属性传递。在 `Update` 活动的情形下，如果投票中除了 `votersCount`、`replies.totalItems` 或 `closed` 之外的任何内容发生了变化，那么就表明包裹的贴文以需要重新创建的方式进行了编辑，因此需要重置。你可以期望在以下时间收到这些活动：

- "Create"：刚刚创建了带有附加投票的贴文

- "Update"：投票/投票人数发生了变化，或者投票刚刚结束

你可以期望的，由 GoToSocial 生成的 `Question` 可以在上面的伪 JSON 中看到。在此 JSON 中，"endTime" 字段将始终被设置（因为我们不支持创建无尽投票），而 "closed" 字段只有在投票结束时才会设置。

### 传入

GoToSocial 期望以与发出投票几乎相同的方式接收投票，在解析 `Question` 对象时采用略显宽容的规则。

- 如果提供 "closed" 而不提供 "endTime"，那么这也将被视为 "endTime" 的值

- 如果既没有提供 "closed" 也没有 "endTime"，则认为投票是永不结束的投票

- 任何情况下，若一个带有 `Question` 的 `Update` 活动提供了一个 `closed` 时间，而之前的活动没有提供，则假定投票刚刚关闭。这将在本站参与投票的用户的客户端触发通知

## 投票行动

为了联合投状态票，GoToSocial 使用特殊格式化 [ActivityStreams "Note" 类型](https://www.w3.org/TR/activitystreams-vocabulary/#dfn-note)。这被 ActivityPub 服务器广泛接受为联合投票的方式，仅将投票作为 "Create" 活动的 "Object" 附加对象。

GoToSocial 传输的 "Note" 类型（以及期望接收到的）包含内容有：
- "name": [确切的投票选项文本]
- "content": [未设置]
- "inReplyTo": [指向 AS Question 的 IRI]
- "attributedTo": [投票作者的 IRI]
- "to": [投票作者的 IRI]

例如：

```json
{
  "type": "Note",
  "name": "选项 1",
  "inReplyTo": "https://example.org/users/bobby_tables/statuses/123456",
  "attributedTo": "https://sample.com/users/willy_nilly",
  "to": "https://example.org/users/bobby_tables"
}
```

### 外发

你可以期望以上面特定描述形式接收到来自 GoToSocial 的投票。投票仅作为附属于 "Create" 活动的对象发送。

特别地，如上节所述，GoToSocial 会在 `name` 字段中提供选项文本，不设置 `content` 字段，在 `inReplyTo` 字段提供一个 IRI，指向你的实例上的带投票贴文。

以下是一个 `Create` 示例，其中用户 `https://sample.com/users/willy_nilly` 在用户 `https://example.org/users/bobby_tables` 创建的多选投票中投票：

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "https://sample.com/users/willy_nilly",
  "id": "https://sample.com/users/willy_nilly/activity#vote/https://example.org/users/bobby_tables/statuses/123456",
  "object": [
    {
      "attributedTo": "https://sample.com/users/willy_nilly",
      "id": "https://sample.com/users/willy_nilly#01HEN2R65468ZG657C4ZPHJ4EX/votes/1",
      "inReplyTo": "https://example.org/users/bobby_tables/statuses/123456",
      "name": "纸巾",
      "to": "https://example.org/users/bobby_tables",
      "type": "Note"
    },
    {
      "attributedTo": "https://sample.com/users/willy_nilly",
      "id": "https://sample.com/users/willy_nilly#01HEN2R65468ZG657C4ZPHJ4EX/votes/2",
      "inReplyTo": "https://example.org/users/bobby_tables/statuses/123456",
      "name": "金融时报",
      "to": "https://example.org/users/bobby_tables",
      "type": "Note"
    }
  ],
  "published": "2021-09-11T11:45:37+02:00",
  "to": "https://example.org/users/bobby_tables",
  "type": "Create"
}
```

### 传入

GoToSocial 期望以与发送投票的几乎相同形式接收投票。即只会期望把投票作为 "Create" 活动的一部分接收。

特别地，GoToSocial 将 votes 识别为不同于其他 "Note" 对象，因为其包含一个 "name" 字段，省略 "content" 字段，且 "inReplyTo" 字段是指向带附有投票的贴文的 URI。 如果满足这些条件，GoToSocial 将把提供的 "Note" 视为格式不正确的贴文对象。

## 贴文删除

GoToSocial 允许用户删除他们创建的贴文。这些删除操作将会向其他实例进行联合，其他实例也应删除其缓存的贴文。

### 外发

当 GoToSocial 用户删除贴文时，服务器会向其他实例发送一个 `Delete` 活动。

`Delete` 活动的 `Object` 条目会设置为该贴文的 ActivityPub URI。

`to` 和 `cc` 将根据原始贴文的可见性以及任何被提及/回复的用户进行设置。

如果原始贴文不是私信，ActivityPub `Public` URI 将在 `to` 中注明。否则，只会涉及被提及和回复的用户。

在以下示例中，'admin' 用户删除了一篇公开贴文，其中提到了 'foss_satan' 用户：

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://example.org/users/admin",
  "cc": [
    "http://example.org/users/admin/followers",
    "http://fossbros-anonymous.io/users/foss_satan"
  ],
  "object": "http://example.org/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0",
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Delete"
}
```

在下一个示例中，'1happyturtle' 用户删除了一条原本发给 'the_mighty_zork' 用户的直接消息。

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://example.org/users/1happyturtle",
  "cc": [],
  "object": "http://example.org/users/1happyturtle/statuses/01FN3VJGFH10KR7S2PB0GFJZYG",
  "to": "http://somewhere.com/users/the_mighty_zork",
  "type": "Delete"
}
```

要处理从 GoToSocial 实例发出的 `Delete` 活动，外站实例应检查其是否根据提供的 URI 存储了 `Object`。如果有，它们应从本站缓存中删除该对象。如果没有，那么无需执行任何操作，因为它们从未存储过现在被删除的贴文。

### 接收

GoToSocial 尽可能彻底地处理来自外站实例的 `Delete` 活动，以尊重其他用户的隐私。

当 GoToSocial 实例收到 `Delete` 时，它会尝试从 `Object` 字段中提取被删除的贴文 URI。如果 `Object` 仅是一个 URI，则使用该 URI。如果 `Object` 是一个 `Note` 或其他常用于表示贴文的类型，则会从中提取 URI。

然后，GoToSocial 将检查其是否存储了具有给定 URI 的贴文。如果有，它将在数据库和所有用户时间线上完全删除。

GoToSocial 仅在确认原帖是被 `Delete` 所属的 `actor` 所拥有的情况下才会删除对应贴文。

## 贴文串

由于去中心化和联合的特性，Fediverse 上的任何一个服务器几乎不可能知道给定贴文串中的每篇贴文。

即便如此，也可以尽力对贴文串进行解引用，从不同的外站实例拉取回复，以更充分地展现整个对话。

GoToSocial 通过在对话贴文串上下迭代，尽可能获取外站贴文，来实现这一点。

假设我们有两个账户：`local_account` 在 `our.server` 上，`remote_1` 在 `remote.1` 上。

在这种情况下，`local_account` 关注了 `remote_1`，所以 `remote_1` 的贴文会出现在 `local_account` 的主页时间线上。

现在，`remote_1` 转发/转贴了来自第三方账户 `remote_2` 的一篇贴文，该账户在服务器 `remote.2` 上。

`local_account` 未关注 `remote_2`，`our.server` 上也没有其他人关注，因此 `our.server` 未曾见过 `remote_2` 的这篇贴文。

![贴文串的示意图，展示了来自 remote_2 的贴文，以及可能的祖先和后代贴文](../public/diagrams/conversation_thread.png)

此时，GoToSocial 会对 `remote_2` 的贴文进行“解引用”，检查其是否属于某个贴文串，以及贴文串的其他部分是否可以获取。

GtS 首先检查贴文的 `inReplyTo` 属性，该属性在贴文回复其他贴文时设置。[见此处](https://www.w3.org/TR/activitystreams-vocabulary/#dfn-inreplyto)。如果设置了 `inReplyTo`，GoToSocial 会解引用被回复的贴文。如果 *这篇* 贴文也设置了 `inReplyTo`，那么 GoToSocial 也会对此进行解引用，如此反复。

一旦获取到贴文的所有 **祖先** 后，GtS 将开始处理贴文的 **后代**。

这种情况下通过检查解引用贴文的 `replies` 属性，依次处理回复及回复的回复。[见此处](https://www.w3.org/TR/activitystreams-vocabulary/#dfn-replies)。

这个贴文串解引用的过程可能需要进行多次 HTTP 请求到不同的服务器，尤其是在贴文串长且复杂的情况下。

解引用的最终结果是，假设由 `remote_2` 转发的贴文属于一个贴文串，那么 `local_account` 在主页时间线上打开贴文时，现在应该能够看到贴文串中的贴文。换句话说，他们将看到来自其他服务器账户的回复（他们可能尚未相遇），以及由 `remote_2` 发布的贴文串之前和之后的贴文。

这为 `local_account` 提供更完整的对话视图，而不仅仅是孤立和断章取义地看到被转发的贴文。此外，这还为 `local_account` 提供了根据对 `remote_2` 的回复发现新账户以进行关注的机会。
