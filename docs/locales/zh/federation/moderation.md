# 管理

## 举报 / 标记

与其他微博 ActivityPub 实现类似，GoToSocial 使用 [Flag](https://www.w3.org/TR/activitystreams-vocabulary/#dfn-flag) 活动类型来向其他服务器传达用户举报信息。

### 发送举报

发送的 GoToSocial `Flag` 的 JSON 格式如下:

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://example.org/users/example.org",
  "content": "此用户频繁骚扰其它用户",
  "id": "http://example.org/reports/01GP3AWY4CRDVRNZKW0TEAMB5R",
  "object": [
    "http://fossbros-anonymous.io/users/foss_satan",
    "http://fossbros-anonymous.io/users/foss_satan/statuses/01FVW7JHQFSFK166WWKR8CBA6M"
  ],
  "type": "Flag"
}
```

`Flag` 的 `actor` 始终是创建该 `Flag` 的 GoToSocial 实例的实例行为体。这样做是为了保护创建举报的用户，实现部分的匿名性，以防止他们成为骚扰目标。

`Flag` 的 `content` 是由创建 `Flag` 的用户提交的一段文本，该文本应该向外站管理员提供创建举报的理由。如果用户没有提交理由，`content` 可能是空字符串，也可能不存在于 JSON 中。

`Flag` 的 `object` 字段的值可以是一个字符串（被举报用户的 ActivityPub `id`），或者是一个字符串数组，其中数组的第一个条目是被举报用户的 `id`，后续条目是一个或多个被举报的 `Note` / 贴文的 `id`。

`Flag` 活动会原样发送到被举报用户的 `inbox`（或共享收件箱）。它不会被包装在 `Create` 活动中。

### 接收举报

GoToSocial 假设接收到的举报会作为 `Flag` 活动发送到被举报用户的 `inbox`。它将按照创建发送 `Flag` 的相同方式解析接收到的 `Flag`，但有一个不同之处：它会尝试从 `object` 字段和 Misskey/Calckey 格式的 `content` 值中解析贴文的 URL，这些值包含内联的贴文 URL。

GoToSocial 不会假设接收到的 `Flag` 活动中会设置 `to` 字段。相反，它假定外站使用 `bto` 将 `Flag` 发送给其接收者。

接收到的有效 `Flag` 活动将作为举报提供给接收到举报的 GoToSocial 实例管理员，以便他们对被举报用户采取必要的管理措施。

除非 GtS 管理员选择通过其他渠道与被举报用户分享此信息，被举报用户本人不会看到举报信息，也不会收到被举报的通知。
