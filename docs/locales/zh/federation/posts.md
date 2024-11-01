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

!!! Note
    在上述所有情况下，如果推断的语言无法解析为有效的 BCP47 语言话题标签，则语言将回退为未知。

## 互动规则

GoToSocial 使用 `interactionPolicy` 属性告知外站给定帖文允许的互动类型（有前提）。

!!! danger
    
    互动规则旨在限制用户贴文上用户不希望的回复和其他互动的有害影响（例如，“回复家(reply guys)” —— 不请自来地发表冒失回复的人）。
    
    然而，这远远不能满足这一目的，因为仍然有许多“额外”方式可以分发或回复贴文，进而超出用户的初衷或意图。
    
    例如，用户可能创建一个附有非常严格互动规则的贴文，却发现其他服务器软件不尊重该规则，而其他实例上的用户从他们的实例视角进行讨论并回复该贴文。原始发布者的实例将自动从视图中删除这些用户不想要的互动，但外站实例可能仍会显示它们。
    
    另一个例子：有人可能会看到一个指定没人可以回复的贴文，但截屏该贴文，将截屏作为新帖文发布，并将提及原作者或只是附上原贴文的 URL。在这种情况下，他们成功通过新创建的贴文串来达到“回复”该贴文的目的。
    
    无论好坏，GoToSocial 只能为这一部分问题提供尽最大努力的部分技术解决方案，这更多的是一个社会行为和边界的问题。

### 概述

`interactionPolicy` 是贴文类 `Object`（如 `Note`、`Article`、`Question` 等）附带的一个属性，其格式如下：

```json
{
  [...],
  "interactionPolicy": {
    "canLike": {
      "always": [ "始终可进行此操作的零个或多个 URI" ],
      "approvalRequired": [ "需要批准才能进行此操作的零个或多个 URI" ]
    },
    "canReply": {
      "always": [ "始终可进行此操作的零个或多个 URI" ],
      "approvalRequired": [ "需要批准才能进行此操作的零个或多个 URI" ]
    },
    "canAnnounce": {
      "always": [ "始终可进行此操作的零个或多个 URI" ],
      "approvalRequired": [ "需要批准才能进行此操作的零个或多个 URI" ]
    }
  },
  [...]
}
```

在此对象中：

- `canLike` 指定可创建 `Like` 并将帖文 URI 作为 `Like` 的 `Object` 的人。
- `canReply` 指定可创建 `inReplyTo` 设置为帖文 URI 的帖文的人。
- `canAnnounce` 指定可创建 `Announce` 并将帖文 URI 作为 `Announce` 的 `Object` 的人。 

并且：

- `always` 是一个 ActivityPub URI/ID 的 `Actor` 或 `Actor` 的 `Collection`，无需 `Accept` 即可进行互动分发到其粉丝。
- `approvalRequired` 是一个 ActivityPub URI/ID 的 `Actor` 或 `Actor` 的 `Collection`，可以互动，但在将互动分发给其粉丝之前需要获得 `Accept`。

`always` 和 `approvalRequired` 的有效 URI 条目包括 magic ActivityStreams 公共 URI `https://www.w3.org/ns/activitystreams#Public`，贴文创建者的 `Following` 和/或 `Followers` 集合的 URI，以及个人请求体的 URI。例如：

```json
[
    "https://www.w3.org/ns/activitystreams#Public",
    "https://example.org/users/someone/followers",
    "https://example.org/users/someone/following",
    "https://example.org/users/someone_else",
    "https://somewhere.else.example.org/users/someone_on_a_different_instance"
]
```

### 指定无人能进行的操作

!!! note
    即使规则指定无人可互动，GoToSocial 仍做出默认假设。参见[默认假设](#默认假设)。

空数组或缺少/空的键表示无人能进行此互动。

例如，以下 `canLike` 指定无人能 `Like` 该贴文：

```json
"canLike": {
  "always": [],
  "approvalRequired": []
},
```

类似的，`canLike` 值为 `null` 也表示无人能 `Like` 该帖：

```json
"canLike": null
```

或

```json
"canLike": {
  "always": null,
  "approvalRequired": null
}
```

缺失的 `canLike` 值同样表达了相同的意思：

```json
{
  [...],
  "interactionPolicy": {
    "canReply": {
      "always": [ "始终可进行此操作的零个或多个 URI" ],
      "approvalRequired": [ "需要批准才能进行此操作的零个或多个 URI" ]
    },
    "canAnnounce": {
      "always": [ "始终可进行此操作的零个或多个 URI" ],
      "approvalRequired": [ "需要批准才能进行此操作的零个或多个 URI" ]
    }
  },
  [...]
}
```

### 冲突/重复值

在用户位于集合 URI 中, 且也通过 URI 被显式指定的情况下，**更具体的值**优先。

例如：

```json
[...],
"canReply": {
  "always": [
    "https://example.org/users/someone"
  ],
  "approvalRequired": [
    "https://www.w3.org/ns/activitystreams#Public"
  ]
},
[...]
```

在此情形下，`@someone@example.org` 位于 `always` 数组中，并且也存在于 `approvalRequired` 数组中的 magic ActivityStreams 公共集合中。在这种情况下，他们可以始终回复，因为 `always` 值更为明确。

另一个例子：

```json
[...],
"canReply": {
  "always": [
    "https://www.w3.org/ns/activitystreams#Public"
  ],
  "approvalRequired": [
    "https://example.org/users/someone"
  ]
},
[...]
```

在此，`@someone@example.org` 位于 `approvalRequired` 数组中，但也隐含地存在于 `always` 数组中的 magic ActivityStreams 公共集合中。在这种情况下，除了 `@someone@example.org` 需要批准外，其他人都可以无需批准进行回复。

在相同 URI 存在于 `always` 和 `approvalRequired` 两者中时，**最高级别的权限**优先（即 `always` 中的 URI 优先于 `approvalRequired` 中的相同 URI）。

### 默认假设

GoToSocial 对 `interactionPolicy` 做出若干默认假设。

**首先**，无论贴文的可见性和 `interactionPolicy` 如何，被[提及](#提及)或回复的用户应**始终**能够回复该贴而无需批准，**除非**提及或回复他们的贴文本身正在等待批准。

这是为了防止潜在的骚扰者在辱骂贴文中提及某人，并不给被提及的用户任何回复的机会。

因此，当发送互动规则时，GoToSocial 始终将提及用户的 URI 添加到 `canReply.always` 数组中，除非它们已被 magic ActivityStreams 公共 URI 覆盖。

同样，在执行接收到的互动规则时，即使用户 URI 不存在于 `canReply.always` 数组中，GoToSocial 也将被提及用户的 URI 视作已存在。

**其次**，用户应**始终**能够回复自己的贴文，点赞自己的贴文，并转发自己的贴文而无需批准，**除非**该贴文本身正在等待批准。

因此，当发送互动规则时，GoToSocial 始终将贴文作者的 URI 添加到 `canLike.always`、`canReply.always` 和 `canAnnounce.always` 数组中，除非它们已被 magic ActivityStreams 公共 URI 覆盖。

同样，在执行接收到的互动规则时，即使贴文作者 URI 不存在于这些 `always` 数组中，GoToSocial 也始终将贴文作者 URI 视为已存在。

### 默认值

当贴文上没有 `interactionPolicy` 属性时，GoToSocial 会根据贴文可见级别和发帖作者为该帖假定默认的 `interactionPolicy`。

对于 `@someone@example.org` 的**公开**或**未列出**贴文，默认 `interactionPolicy` 为：

```json
{
  [...],
  "interactionPolicy": {
    "canLike": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": []
    },
    "canReply": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": []
    },
    "canAnnounce": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": []
    }
  },
  [...]
}
```

对于 `@someone@example.org` 的**仅限粉丝**贴文，假定的 `interactionPolicy` 为：

```json
{
  [...],
  "interactionPolicy": {
    "canLike": {
      "always": [
        "https://example.org/users/someone",
        "https://example.org/users/someone/followers",
        [...提及的用户的 URI...]
      ],
      "approvalRequired": []
    },
    "canReply": {
      "always": [
        "https://example.org/users/someone",
        "https://example.org/users/someone/followers",
        [...提及的用户的 URI...]
      ],
      "approvalRequired": []
    },
    "canAnnounce": {
      "always": [
        "https://example.org/users/someone"
      ],
      "approvalRequired": []
    }
  },
  [...]
}
```

对于 `@someone@example.org` 的**私信**贴文，假定的 `interactionPolicy` 为：

```json
{
  [...],
  "interactionPolicy": {
    "canLike": {
      "always": [
        "https://example.org/users/someone",
        [...提及的用户的 URI...]
      ],
      "approvalRequired": []
    },
    "canReply": {
      "always": [
        "https://example.org/users/someone",
        [...提及的用户的 URI...]
      ],
      "approvalRequired": []
    },
    "canAnnounce": {
      "always": [
        "https://example.org/users/someone"
      ],
      "approvalRequired": []
    }
  },
  [...]
}
```

### 示例 1 - 限制对话范围

在此示例中，用户 `@the_mighty_zork` 想开始与用户 `@booblover6969` 和 `@hodor` 之间的对话。

为了避免讨论被他人打断，他们希望来自三名参与者以外的用户的回复仅在获得 `@the_mighty_zork` 批准后才被允许。

此外，他们希望将贴文转发/`Announce` 的权利限制为仅限于他们自己的粉丝和三个对话参与者。

然而，任何人都可以 `Like` `@the_mighty_zork` 的贴文。

这可以通过以下 `interactionPolicy` 来实现，它附加在可见性为公开的帖文上：

```json
{
  [...],
  "interactionPolicy": {
    "canLike": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": []
    },
    "canReply": {
      "always": [
        "https://example.org/users/the_mighty_zork",
        "https://example.org/users/booblover6969",
        "https://example.org/users/hodor"
      ],
      "approvalRequired": [
        "https://www.w3.org/ns/activitystreams#Public"
      ]
    },
    "canAnnounce": {
      "always": [
        "https://example.org/users/the_mighty_zork",
        "https://example.org/users/the_mighty_zork/followers",
        "https://example.org/users/booblover6969",
        "https://example.org/users/hodor"
      ],
      "approvalRequired": []
    }
  },
  [...]
}
```

### 示例 2 - 长独白贴文串

在此示例中，用户 `@the_mighty_zork` 想写一个长篇独白。

他们不介意别人转发和点赞贴文，但不想收到任何回复，因为他们没有精力去管理讨论；他们只是想通过发泄自己的想法去表达自我。

这可以通过在贴文串中的每个贴文上设置以下 `interactionPolicy` 来实现：

```json
{
  [...],
  "interactionPolicy": {
    "canLike": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": []
    },
    "canReply": {
      "always": [
        "https://example.org/users/the_mighty_zork"
      ],
      "approvalRequired": []
    },
    "canAnnounce": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": []
    }
  },
  [...]
}
```

在这里，任何人都可以点赞或转发，但无人能够回复（除了 `@the_mighty_zork` 自己）。

### 示例 3 - 完全开放

在此示例中，`@the_mighty_zork` 想写一篇完全开放的贴文，任何能看到此帖的人都可以进行回复、转发或点赞（即解锁和公开贴文默认行为）：

```json
{
  [...],
  "interactionPolicy": {
    "canLike": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": []
    },
    "canReply": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": []
    },
    "canAnnounce": {
      "always": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "approvalRequired": []
    }
  },
  [...]
}
```

### 请求、获得和验证批准

当用户的 URI 在需要获得批准的互动的 `approvalRequired` 数组中时，如果他们希望获得批准以分发互动，应该执行以下步骤：

1. 像往常一样撰写互动 `Activity`（即 `Like`、`Create` （回复）或 `Announce`）。
2. 像往常一样将 `Activity` 的 `to` 和 `cc` 地址设为预期的收件人。
3. 将 `Activity` **仅**发送到要互动帖的作者的 `Inbox`（或 `sharedInbox`）。
4. **此时不要进一步分发 `Activity`**。

此时，互动可视为等待批准，并不应该显示在被互动的贴文的回复或点赞集合等中。

可以向发送互动的用户显示“互动待批准”状态，但理想情况下不应该向与该用户共享实例的其他用户显示。

从这里开始，可能会出现以下三种情况之一：

#### 拒绝

在这种情况下，互动目标贴文的作者发回一个 `Reject` `Activity`，该活动的 `Object` 属性带有待拒绝互动活动的 URI/ID。

例如，以下 JSON 对象 `Reject` 了 `@someone@somewhere.else.example.org` 回复 `@post_author@example.org` 贴文的尝试：

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "https://example.org/users/post_author",
  "to": "https://somewhere.else.example.org/users/someone",
  "id": "https://example.org/users/post_author/activities/reject/01J0K2YXP9QCT5BE1JWQSAM3B6",
  "object": "https://somewhere.else.example.org/users/someone/statuses/01J17XY2VXGMNNPH1XR7BG2524",
  "type": "Reject"
}
```

如果发生这种情况，`@someone@somewhere.else.example.org`（及其实例）应视交互为被拒绝。该实例应从其内部存储（即数据库）中删除该活动，或以其他方式表明它已被拒绝，并且不应进一步分发该 `Activity` 或重试该交互。

#### 无响应

在这种情况下，正在互动的贴文的作者从不发送 `Reject` 或 `Accept` `Activity`。在这种情况下，交互被视为永久“待处理”。实例可能希望实现某种清理功能，达到一定时间期限的已发送且待处理交互应被视为过期，然后按照上述方式被处理为 `Rejected` 并删除。

#### 接受

在这种情况下，正在互动的贴文的作者发回一个`Accept` `Activity`，该活动的 `Object` 属性带有待批准互动活动的 URI/ID。

例如，以下 JSON 对象 `Accept` 了 `@someone@somewhere.else.example.org` 回复 `@post_author@example.org` 贴文的尝试：

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "https://example.org/users/post_author",
  "cc": [
    "https://www.w3.org/ns/activitystreams#Public",
    "https://example.org/users/post_author/followers"
  ],
  "to": "https://somewhere.else.example.org/users/someone",
  "id": "https://example.org/users/post_author/activities/accept/01J0K2YXP9QCT5BE1JWQSAM3B6",
  "object": "https://somewhere.else.example.org/users/someone/statuses/01J17XY2VXGMNNPH1XR7BG2524",
  "type": "Accept"
}
```

如果发生这种情况，`@someone@somewhere.else.example.org`（及其实例）应视为交互已被批准/接受。然后，该实例可以自由地将此交互 `Activity` 分发给所有由 `to`、`cc` 等目标的接收者，并附加属性 `approvedBy`。

### 验证在粉丝/关注中是否存在

如果一个 `Actor` 在其互动规则的 `always` 字段中因为存在于 `Followers` 或 `Following` 集合中而被允许进行交互（例如 `Like`、`inReplyTo` 或 `Announce`），则其服务器仍应等待来自目标帐户服务器的 `Accept`，然后才更广泛地分发交互，并将 `approvedBy` 属性设置为 `Accept` 的 URI。

这样可以防止第三方服务器被迫以某种方式验证互动的 `Actor` 是否存在于接收互动的用户的 `Followers` 或 `Following` 集合中。让目标服务器进行验证，并采信其 `Accept` ，将其视为交互 `Actor` 存在于相关集合中的证明，更为简单。

同样，当接收到一个具有匹配 `Following` 或 `Followers` 集合的 `Actor` 的互动请求时，接收互动的 `Actor` 的服务器应确保尽快发送出 `Accept`，以便交互 `Actor` 服务器可以带着适当的接受证明发送出 `Activity`。

这个过程应绕过通常的“待批准”阶段，因此没有必要通知 `Actor` 待批准的交互，因为他们已明确同意。在 GoToSocial 代码库中，这被称为“预批准”。

### `approvedBy`

`approvedBy` 是一个附加属性，添加到 `Like` 和 `Announce` 活动以及任何被视为“贴文”的 `Object`（如 `Note`、`Article`）中。

`approvedBy` 的存在表明贴文的作者接受了由 `Activity` 作为目标或由 `Object` 所回复的互动，并现在可以分发给其预期观众。

`approvedBy` 的值应为创建 `Accept` `Activity` 的接收交互贴文作者的 URI。

例如，以下 `Announce` `Activity` 的 `approvedBy` 表示它已被 `@post_author@example.org` `Accept`：

```json
{
  "actor": "https://somewhere.else.example.org/users/someone",
  "to": [
    "https://somewhere.else.example.org/users/someone/followers"
  ],
  "cc": [
    "https://example.org/users/post_author"
  ],
  "id": "https://somewhere.else.example.org/users/someone/activities/announce/01J0K2YXP9QCT5BE1JWQSAM3B6",
  "object": "https://example.org/users/post_author/statuses/01J17ZZFK6W82K9MJ9SYQ33Y3D",
  "approvedBy": "https://example.org/users/post_author/activities/accept/01J18043HGECBDZQPT09CP6F2X",
  "type": "Announce"
}
```

接收到一个带有 `approvedBy` 值的 `Activity` 时，外站实例应解引用字段的 URI 值以获取 `Accept` `Activity`。

然后，他们应验证 `Accept` `Activity` 的 `object` 值是否等于交互 `Activity` 或 `Object` 的 `id`，并验证 `actor` 值是否等于接收交互的贴文的作者。

此外，他们应确保解引用的 `Accept` 的 URL 域名等于接收交互贴文的作者的 URL 域名。

如果无法解引用 `Accept` 或未通过有效性检查，则交互应被视为无效并丢弃。

由于这种验证机制，实例应确保他们对涉及 `interactionPolicy` 的 `Accept` URI 的解引用响应提供一个有效的 ActivityPub 对象。如果不这样做，他们会无意中限制外站实例分发其贴文的能力。

### 后续回复/范围扩展

对话中的每个后续回复将有其自己的互动规则，由创建回复的用户选择。换句话说，整个*对话*或*贴文串*并不由一个 `interactionPolicy` 控制，而是贴文串中的每个后续贴文可以由贴文作者设置不同的规则。

不幸的是，这意味着即使有 `interactionPolicy`，贴文串的范围也可能不小心超出第一个贴文作者的意图。

例如，在上述[示例 1](#示例-1---限制对话范围)中，`@the_mighty_zork` 在第一个贴文中指定了 `canReply.always` 值为

```json
[
  "https://example.org/users/the_mighty_zork",
  "https://example.org/users/booblover6969",
  "https://example.org/users/hodor"
]
```

在后续回复中，`@booblover6969` 无意或有意地将 `canReply.always` 值设为：

```json
[
  "https://www.w3.org/ns/activitystreams#Public"
]
```

这扩大了对话的范围，因为现在任何人都可以回复 `@booblover6969` 的贴文，并可能也在该回复中标记 `@the_mighty_zork`。

为了避免这个问题，建议外站实例防止用户能够扩大范围（具体机制待定）。

同时，实例应将任何与仍处于待批准状态的贴文或贴文类似的 `Object` 的交互视作待批准。

换句话说，只要某条贴文处于待批准状态，实例应将该贴文下的所有互动标记为待批准，无论此贴文的互动规则通常允许什么。

这可避免有用户回复贴文，且在回复尚未得到批准的情况下继续回复*他们自己的回复*并将其标记为允许（作为贴文回复的作者，他们默认拥有对贴文回复的[回复权限](#默认假设)）。

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
