# 行为体与行为体属性

## 收件箱

GoToSocial 按照 [ActivityPub 规范](https://www.w3.org/TR/activitypub/#inbox)，为行为体实现了收件箱。

如规范所述，[外站](https://www.w3.org/TR/activitypub/#delivery) 应通过向活动目标受众的每个收件箱发送 HTTP POST 请求，将活动传送到 GoToSocial 服务器。

GoToSocial 帐号目前没有实现 [共享收件箱](https://www.w3.org/TR/activitypub/#shared-inbox-delivery) 端点，但这可能会有所改变。当 GoToSocial 服务器上有多个行为体是活动受众时，对已传送活动的去重由 GoToSocial 处理。

对 GoToSocial 行为体收件箱的 POST 请求必须由发起活动的行为体进行正确地 [HTTP 签名](#http-signatures)。

可被接受的收件箱 POST `Content-Type` 头为：

- `application/activity+json`
- `application/activity+json; charset=utf-8`
- `application/ld+json; profile="https://www.w3.org/ns/activitystreams"`

未使用上述 `Content-Type` 头之一的收件箱 POST 请求将被拒绝，并返回 HTTP 状态码 [406 - Not Acceptable](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/406)。

有关可接受内容类型的更多信息，请参阅 ActivityPub 协议的 [服务器间交互](https://www.w3.org/TR/activitypub/#server-to-server-interactions) 部分。

对格式正确且已签名的收件箱 POST 请求，GoToSocial 将返回 HTTP 状态码 [202 - Accepted](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/202)。

对格式错误的收件箱 POST 请求，将返回 HTTP 状态码 [400 - Bad Request](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400)。响应正文可能包含有关 GoToSocial 服务器为何认为请求内容格式错误的更多信息。对于代码 `400` 的回应，其他服务器不应重试交付活动。

即使 GoToSocial 返回 `202` 状态码，也可能不继续处理已传送的活动，具体取决于活动的发起者、目标和活动类型。ActivityPub 是一个广泛的协议，GoToSocial 并未涵盖每种活动和对象的组合。

## 发件箱

GoToSocial 按照 [ActivityPub 规范](https://www.w3.org/TR/activitypub/#outbox)，为行为体（即实例账户）实现了发件箱。

要获取某行为体最近发布的活动 [OrderedCollection](https://www.w3.org/TR/activitystreams-vocabulary/#dfn-orderedcollection)，外站可以对用户的发件箱进行 `GET` 请求。其地址类似于 `https://example.org/users/whatever/outbox`。

服务器将返回以下结构的 OrderedCollection：

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "https://example.org/users/whatever/outbox",
  "type": "OrderedCollection",
  "first": "https://example.org/users/whatever/outbox?page=true"
}
```

请注意，`OrderedCollection` 本身不包含项目。调用者必须解引用 `first` 页面以开始获取项目。例如，对 `https://example.org/users/whatever/outbox?page=true` 的 `GET` 请求将生成如下内容：

```json
{
  "id": "https://example.org/users/whatever/outbox?page=true",
  "type": "OrderedCollectionPage",
  "next": "https://example.org/users/whatever/outbox?max_id=01FJC1Q0E3SSQR59TD2M1KP4V8&page=true",
  "prev": "https://example.org/users/whatever/outbox?min_id=01FJC1Q0E3SSQR59TD2M1KP4V8&page=true",
  "partOf": "https://example.org/users/whatever/outbox",
  "orderedItems": [
    {
      "id": "https://example.org/users/whatever/statuses/01FJC1MKPVX2VMWP2ST93Q90K7/activity",
      "type": "Create",
      "actor": "https://example.org/users/whatever",
      "published": "2021-10-18T20:06:18Z",
      "to": [
        "https://www.w3.org/ns/activitystreams#Public"
      ],
      "cc": [
        "https://example.org/users/whatever/followers"
      ],
      "object": "https://example.org/users/whatever/statuses/01FJC1MKPVX2VMWP2ST93Q90K7"
    }
  ]
}
```

`orderedItems` 数组最多包含 30 个条目。要获取超过此数量的条目，调用者可以使用响应中提供的 `next` 链接。

请注意，在返回的 `orderedItems` 中，所有活动类型都将是 `Create`。在每个活动中，`object` 字段将是由拥有发件箱的行为体创建的原始公共贴文的 AP URI（即 `Note`，`to` 字段中包含 `https://www.w3.org/ns/activitystreams#Public`，且不是对另一个贴文的回复）。调用者可以使用返回的 AP URI 来解引用这些 `Note` 的内容。

## 粉丝与关注集合

GoToSocial 将粉丝和关注的集合实现为 `OrderedCollection`。例如，对行为体的关注集合进行正确签名的 `GET` 请求将返回如下内容：

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "first": "https://example.org/users/someone/following?limit=40",
  "id": "https://example.org/users/someone/following",
  "totalItems": 397,
  "type": "OrderedCollection"
}
```

从这里开始，你可以使用 `first` 页面开始获取项目。例如，对 `https://example.org/users/someone/following?limit=40` 的 `GET` 请求将产生如下内容：

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "https://example.org/users/someone/following?limit=40",
  "next": "https://example.org/users/someone/following?limit=40&max_id=01V1AY4ZJT4JK1NT271SH2WMGH",
  "orderedItems": [
    "https://example.org/users/someone_else",
    "https://somewhere.else.example.org/users/another_account",
    [... 38 more entries here ...]
  ],
  "partOf": "https://example.org/users/someone/following",
  "prev": "https://example.org/users/someone/following?limit=40&since_id=021HKBY346X7BPFYANPPJN493P",
  "totalItems": 397,
  "type": "OrderedCollectionPage"
}
```

然后，你可以使用 `next` 和 `prev` 端点在 OrderedCollection 中向下和向上翻页。

## 个人资料字段

与 Mastodon 和其他联邦宇宙软件类似，GoToSocial 允许用户在其个人资料上设置键/值对；这对于传达简短的信息如链接、代词、年龄等很有用。

为了与其他实现兼容，GoToSocial 使用与 Mastodon 相同的 schema.org PropertyValue 扩展，作为设置字段的行为体上的 `attachment` 数组值。例如，以下 JSON 显示了两个 PropertyValue 字段的账户：

```json
{
  "@context": [
    "http://joinmastodon.org/ns",
    "https://w3id.org/security/v1",
    "https://www.w3.org/ns/activitystreams",
    "http://schema.org"
  ],
  "attachment": [
    {
      "name": "接受关注",
      "type": "PropertyValue",
      "value": "纯看个人心情"
    },
    {
      "name": "年龄",
      "type": "PropertyValue",
      "value": "120"
    }
  ],
  "discoverable": false,
  "featured": "http://example.org/users/flyingsloth/collections/featured",
  "followers": "http://example.org/users/flyingsloth/followers",
  "following": "http://example.org/users/flyingsloth/following",
  "id": "http://example.org/users/flyingsloth",
  "inbox": "http://example.org/users/flyingsloth/inbox",
  "manuallyApprovesFollowers": true,
  "name": "飞翔的树懒 :3",
  "outbox": "http://example.org/users/flyingsloth/outbox",
  "preferredUsername": "flyingsloth",
  "publicKey": {
    "id": "http://example.org/users/flyingsloth#main-key",
    "owner": "http://example.org/users/flyingsloth",
    "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtTc6Jpg6LrRPhVQG4KLz\n2+YqEUUtZPd4YR+TKXuCnwEG9ZNGhgP046xa9h3EWzrZXaOhXvkUQgJuRqPrAcfN\nvc8jBHV2xrUeD8pu/MWKEabAsA/tgCv3nUC47HQ3/c12aHfYoPz3ufWsGGnrkhci\nv8PaveJ3LohO5vjCn1yZ00v6osMJMViEZvZQaazyE9A8FwraIexXabDpoy7tkHRg\nA1fvSkg4FeSG1XMcIz2NN7xyUuFACD+XkuOk7UqzRd4cjPUPLxiDwIsTlcgGOd3E\nUFMWVlPxSGjY2hIKa3lEHytaYK9IMYdSuyCsJshd3/yYC9LqxZY2KdlKJ80VOVyh\nyQIDAQAB\n-----END PUBLIC KEY-----\n"
  },
  "summary": "\u003cp\u003e只是一只普通树懒\u003c/p\u003e",
  "tag": [],
  "type": "Person",
  "url": "http://example.org/@flyingsloth"
}
```

对于没有 `PropertyValue` 字段的行为体，`attachment` 属性将不设置。即，`attachment` 键值不会在行为体中出现（即使是空数组或 null 值也不会）。

尽管 `attachment` 在规范上不是一个有序集合，GoToSocial（还是为了与其他实现保持一致）仍会按应显示的顺序呈现 `attachment` 的 `PropertyValue` 字段。

GoToSocial 还将解析 GoToSocial 实例发现的外站行为体的 PropertyValue 字段，以便可以把它们展示给 GoToSocial 实例上的用户。

GoToSocial 默认允许设置最多 6 个 `PropertyValue` 字段，而 Mastodon 的默认值为 4 个。

## 置顶/特色贴文

GoToSocial 允许用户在他们的个人资料上置顶贴文。

在 ActivityPub 术语中，GoToSocial 在行为体的 [featured](https://docs.joinmastodon.org/spec/activitypub/#featured) 字段中指定的端点提供这些置顶贴文，格式为 [OrderedCollection](https://www.w3.org/TR/activitystreams-vocabulary/#dfn-orderedcollection) 。该字段的值将被设置为类似 `https://example.org/users/some_user/collections/featured`。

通过向此端点发送经过签名的 GET 请求，外站实例可以解引用特色贴文集合，这将返回带有 `orderedItems` 字段，其中包含贴文 URI 列表的 `OrderedCollection`。

置顶多条 `Note` 的用户的 featured collection 示例：

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "https://example.org/users/some_user/collections/featured",
  "orderedItems": [
    "https://example.org/users/some_user/statuses/01GS7VTYH0S77NNXTP6W4G9EAG",
    "https://example.org/users/some_user/statuses/01GSFY2SZK9TPCJFQ1WCCPGDRT",
    "https://example.org/users/some_user/statuses/01GSCXY70MZCBFMH5EKJW9ENC8"
  ],
  "totalItems": 3,
  "type": "OrderedCollection"
}
```

置顶单条 `Note` 的用户的 featured collection 示例：

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "https://example.org/users/some_user/collections/featured",
  "orderedItems": [
    "https://example.org/users/some_user/statuses/01GS7VTYH0S77NNXTP6W4G9EAG"
  ],
  "totalItems": 1,
  "type": "OrderedCollection"
}
```

没有置顶 `Note` 的示例：

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "https://example.org/users/some_user/collections/featured",
  "orderedItems": [],
  "totalItems": 0,
  "type": "OrderedCollection"
}
```

与 Mastodon 和一些其他实现不同，GoToSocial 不会将在 `orderedItems` 的值中提供完整的 `Note` 表示。相反，它仅提供每个 `Note` 的 URI，外站服务器可以自行进一步解引用（如果它们已经在本地缓存了该 `Note` 则可以不执行此操作）。

作为集合一部分提供的一些 URI 可能指向仅限粉丝可见性的贴文，请求的 `Actor` 不一定有权限查看。外站服务器应确保进行适当的过滤（与其他任何类型的贴文一样），以确保这些贴文仅显示给有权查看的用户。

GoToSocial 和其他服务器实现之间的另一个区别是，当用户置顶或取消置顶贴文时，GoToSocial 不会向外站服务器发送更新。Mastodon 会通过发送 [Add](https://www.w3.org/TR/activitypub/#add-activity-inbox) 和 [Remove](https://www.w3.org/TR/activitypub/#remove-activity-inbox) 活动类型来进行，`object` 是被置顶或取消置顶的贴文，`target` 是发送 `Actor` 的 `featured` 集合。尽管在概念上这很合理，但这与 ActivityPub 协议建议不一致，因为活动的 `target`“不属于接收服务器，因此他们不能更新它”。

相反，建议外站仅定期轮询 GoToSocial 行为体的 `featured` 集合，并根据需要在其缓存表示中添加/删除贴文，以构建和更新 GoToSocial 用户置顶贴文的视图。

## 行为体迁移 / 别名

GoToSocial 支持通过 `Move` 活动以及行为体对象属性 `alsoKnownAs` 和 `movedTo` 从一个实例/服务器迁移到另一个。

### `alsoKnownAs`

GoToSocial 支持使用 `alsoKnownAs` 行为体属性进行帐户别名，这是一个 [公认的 ActivityPub 扩展](https://www.w3.org/wiki/Activity_Streams_extensions#as:alsoKnownAs_property)。

#### 传入

在传入的 AP 消息中，GoToSocial 将查找行为体上的 `alsoKnownAs` 属性，该属性是行为体也可以通过的其他活动 IDs/URIs 构成的数组。

例如：

```json
{
  "@context": [
    "http://joinmastodon.org/ns",
    "https://w3id.org/security/v1",
    "https://www.w3.org/ns/activitystreams",
    "http://schema.org"
  ],
  "featured": "http://example.org/users/flyingsloth/collections/featured",
  "followers": "http://example.org/users/flyingsloth/followers",
  "following": "http://example.org/users/flyingsloth/following",
  "id": "http://example.org/users/flyingsloth",
  "inbox": "http://example.org/users/flyingsloth/inbox",
  "manuallyApprovesFollowers": true,
  "name": "飞翔的树懒 :3",
  "outbox": "http://example.org/users/flyingsloth/outbox",
  "preferredUsername": "flyingsloth",
  "publicKey": {...},
  "summary": "\u003cp\u003e只是一只普通树懒\u003c/p\u003e",
  "type": "Person",
  "url": "http://example.org/@flyingsloth",
  "alsoKnownAs": [
    "https://another-server.com/users/flyingsloth",
    "https://somewhere-else.org/users/originalsloth"
  ]
}
```

在上述 AP JSON 中，行为体 `http://example.org/users/flyingsloth` 已设置别名为其他行为体 `https://another-server.com/users/flyingsloth` 和 `https://somewhere-else.org/users/originalsloth`。

GoToSocial 将传入的 `alsoKnownAs` URI 存储在数据库中，但（当前）不会使用它们，除非用于验证 `Move` 活动（见下文）。

#### 传出

GoToSocial 用户可以通过 GoToSocial 客户端 API 在其账户上设置多个 `alsoKnownAs` URI。GoToSocial 会在存入数据库并在传出 AP 消息序列化之前验证这些 `alsoKnownAs` 别名是否为有效的行为体 URI。

然而，GoToSocial 并不验证用户在设置别名之前对那些 `alsoKnownAs` URI 的*所有权*；它期望外站自行进行验证，然后再采信任何传入的 `alsoKnownAs` 值。

例如，GoToSocial 实例中的用户 `http://example.org/users/flyingsloth` 可能会在他们的账户上设置 `alsoKnownAs: [ "https://unrelated-server.com/users/someone_else" ]`，GoToSocial 会如实传输此别名到其他服务器。

在这种情况下，`https://unrelated-server.com/users/someone_else` 或许不是 `flyingsloth`。`flyingsloth` 可能无意或恶意地设置了此别名。为了正确验证 `someone_else` 的所有权，外站应检查行为体 `https://unrelated-server.com/users/someone_else` 的 `alsoKnownAs` 属性是否包含 `http://example.org/users/flyingsloth` 条目。

换句话说，外站不应默认信任 `alsoKnownAs` 别名，而应确保在将别名视为有效之前，行为体之间存在**双向别名**。

### `movedTo`

GoToSocial 使用 `movedTo` 属性标记账户已迁移。与 `alsoKnownAs` 不同，这不是一个被接受的 ActivityPub 扩展，但它已被 Mastodon 广泛推广，也在 `Move` 活动中使用。[参见 Mastodon 文档获取更多信息](https://documentation.sig.gy/spec/activitypub/#namespaces)。

#### 传入

对于传入的 AP 消息，GoToSocial 查找行为体上的 `movedTo` 属性，该属性设置为单个 ActivityPub 行为体 URI/ID。

例如：

```json
{
  "@context": [
    "http://joinmastodon.org/ns",
    "https://w3id.org/security/v1",
    "https://www.w3.org/ns/activitystreams",
    "http://schema.org"
  ],
  "featured": "http://example.org/users/flyingsloth/collections/featured",
  "followers": "http://example.org/users/flyingsloth/followers",
  "following": "http://example.org/users/flyingsloth/following",
  "id": "http://example.org/users/flyingsloth",
  "inbox": "http://example.org/users/flyingsloth/inbox",
  "manuallyApprovesFollowers": true,
  "name": "飞翔的树懒 :3",
  "outbox": "http://example.org/users/flyingsloth/outbox",
  "preferredUsername": "flyingsloth",
  "publicKey": {...},
  "summary": "\u003cp\u003e只是一只普通树懒\u003c/p\u003e",
  "type": "Person",
  "url": "http://example.org/@flyingsloth",
  "alsoKnownAs": [
    "https://another-server.com/users/flyingsloth"
  ],
  "movedTo": "https://another-server.com/users/flyingsloth"
}
```

在上述 JSON 中，行为体 `http://example.org/users/flyingsloth` 已设置别名为行为体 `https://another-server.com/users/flyingsloth` 并已迁移/转向该账户。

GoToSocial 将传入的 `movedTo` 值存储在数据库中，但除非行为体在进行移动之前发送了一个 `Move` 活动，否则不会认为帐户迁移已处理（见下文）。

### `Move` 活动

为了实际触发账户迁移，GoToSocial 使用 `Move` 活动，并将行为体 URI 作为对象和目标，例如：

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "https://example.org/users/flyingsloth/moves/01HR9FDFCAGM7JYPMWNTFRDQE9",
  "actor": "https://example.org/users/flyingsloth",
  "type": "Move",
  "object": "https://example.org/users/flyingsloth",
  "target": "https://another-server.com/users/my_new_account_hurray",
  "to": "https://example.org/users/flyingsloth/followers"
}
```

在上述 `Move` 中，行为体 `https://example.org/users/flyingsloth` 指示其账户正在迁移到 URI `https://another-server.com/users/my_new_account_hurray`。

### 迁入

在收到行为体收件箱中的 `Move` 活动时，GoToSocial 将首先通过以下检查验证 `Move`：

1. 请求由 `actor` 签名。
2. `actor` 和 `object` 字段相同（你不能迁移其他人的账户）。
3. `actor` 尚未迁移到其他地方。
4. `target` 是有效的行为体 URI：可检索、未封禁、未迁移，且不在接收到此 `Move` 的 GoToSocial 实例屏蔽的实例上。
5. `target` 将 `alsoKnownAs` 设置为发送 `Move` 的 `actor`。在此示例中，`https://another-server.com/users/my_new_account_hurray` 必须具有 `alsoKnownAs` 值，其中包括 `https://example.org/users/flyingsloth`。

如果检查通过，则 GoToSocial 将通过将粉丝重定向到新账户来处理 `Move`：

1. 选择此 GoToSocial 实例上执行 `Move` 的 `actor` 的所有粉丝。
2. 对于以这种方式选择的每个本站粉丝，从该粉丝的账户发送关注请求到 `Move` 的 `target`。
3. 删除针对“旧” `actor` 的所有关注。

这样做的最终结果是，在接收实例上 `https://example.org/users/flyingsloth` 的所有粉丝现在将关注 `https://another-server.com/users/my_new_account_hurray`。

GoToSocial 还会删除由执行 `Move` 的 `actor` 拥有的所有关注和待关注请求；由 `target` 帐户再次发送关注请求。

为了防止潜在的 DoS 向量，GoToSocial 对 `Move` 强制进行 7 天冷却期。一旦帐户成功转移，GoToSocial 将在上次迁移后的 7 天内不处理来自新帐户的进一步迁移活动。

#### 迁出

发送帐户迁移时，GoToSocial 以类似方式使用 `Move` 活动。当 GoToSocial 实例上的行为体想要执行 `Move` 时，GoToSocial 将首先检查和验证 `Move` 目标，并确保它具有等于执行 `Move` 的行为体的 `alsoKnownAs` 条目。在成功验证后，将向所有发起迁移的行为体的粉丝发送 `Move` 活动，为其指示 `Move` 的 `target`。GoToSocial 期望外站将相应的追随者迁移到 `target` 名下。
