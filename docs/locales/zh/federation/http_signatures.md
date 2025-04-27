# HTTP 签名

GoToSocial 要求所有发送到 ActivityPub 服务器的 `GET` 和 `POST` 请求都必须附带有效的 HTTP 签名。

GoToSocial 也会为其向其他服务器发送的所有 `GET` 和 `POST` 请求签名。

这种行为与 Mastodon 的 [AUTHORIZED_FETCH / "安全模式"](https://docs.joinmastodon.org/admin/config/#authorized_fetch) 等效。

GoToSocial 使用 [superseriousbusiness/httpsig](https://codeberg.org/superseriousbusiness/httpsig) 库（从 go-fed 派生）来为发出的请求签名，并解析和验证传入请求的签名。该库严格遵循 [Cavage HTTP Signature RFC](https://datatracker.ietf.org/doc/html/draft-cavage-http-signatures-12)，这是其他实现（如 Mastodon、Pixelfed、Akkoma/Pleroma 等）使用的同一份 RFC。（此 RFC 后来被 [httpbis HTTP Signature RFC](https://datatracker.ietf.org/doc/html/draft-ietf-httpbis-message-signatures) 取代，但尚未广泛实施。）

## 查询参数

关于是否应该在用于生成和验证签名的 URL 中包含查询参数，HTTP 签名规范并无明确规定。

在历史上，GoToSocial 在签名中包含了查询参数，而大多数其他实现则没有。这导致在对 Collection 端点进行签名 GET 请求或验证签名的 GET 请求时（通常使用查询参数进行分页），出现了兼容性问题。

从 0.14 开始，GoToSocial 尝试同时签署和验证携带和不携带查询参数请求，以确保与其他实现更好的兼容性。

发送请求时，GtS 将首先尝试包含查询参数的情况。当收到外站服务器的 `401` 响应时，它将尝试在不包含查询参数的情况下重新发送请求。

接收请求时，GtS 将首先尝试验证包含查询参数的签名。如果签名验证失败，它将尝试在不包含查询参数的情况下重新验证签名。

详细信息请参见 [#894](https://codeberg.org/superseriousbusiness/gotosocial/issues/894)。

## 传入请求

GoToSocial 的请求签名验证在 [internal/federation](https://codeberg.org/superseriousbusiness/gotosocial/src/branch/main/internal/federation/authenticate.go) 中实现。

GoToSocial 将尝试按以下算法顺序解析签名，成功后将停止：

```text
RSA_SHA256
RSA_SHA512
ED25519
```

## 发出请求

GoToSocial 的请求签名在 [internal/transport](https://codeberg.org/superseriousbusiness/gotosocial/src/branch/main/internal/transport/signing.go) 中实现。

一旦解决了 https://codeberg.org/superseriousbusiness/gotosocial/issues/2991 ，GoToSocial 将使用 `(created)` 伪标头代替 `date`。

然而，目前在组装签名时：

- 发出的 `GET` 请求使用 `(request-target) host date`
- 发出的 `POST` 请求使用 `(request-target) host date digest` 

GoToSocial 在签名中将 `algorithm` 字段设置为 `hs2019`，这一般意味着“从与 keyId 关联的元数据中导出算法”。生成签名时使用的实际算法是 `RSA_SHA256`，这与其他 ActivityPub 实现一致。在验证 GoToSocial 的 HTTP 签名时，外站服务器可以安全地假设该签名是使用 `sha256` 生成的。

## 特点

GoToSocial 在 `Signature` 标头中使用的 `keyId` 格式如下所示：

```text
https://example.org/users/example_user/main-key
```

这不同于大多数其他实现，它们通常在 `keyId` URI 中使用片段 (`#`)。例如，在 Mastodon 上，用户的密钥会这样表示：

```text
https://example.org/users/example_user#main-key
```

对于 Mastodon，用户的公钥作为该用户的 Actor 表示的一部分提供。GoToSocial 在提供用户的公钥时模仿了这种行为，但并不在 `main-key` 端点返回完整的 Actor（这可能包含敏感字段），而是仅返回 Actor 的部分存根。它如下所示：

```json
{
  "@context": [
    "https://w3id.org/security/v1",
    "https://www.w3.org/ns/activitystreams"
  ],
  "id": "https://example.org/users/example_user",
  "preferredUsername": "example_user",
  "publicKey": {
    "id": "https://example.org/users/example_user/main-key",
    "owner": "https://example.org/users/example_user",
    "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzGB3yDvMl+8p+ViutVRG\nVDl9FO7ZURYXnwB3TedSfG13jyskoiMDNvsbLoUQM9ajZPB0zxJPZUlB/W3BWHRC\nNFQglE5DkB30GjTClNZoOrx64vLRT5wAEwIOjklKVNk9GJi1hFFxrgj931WtxyML\nBvo+TdEblBcoru6MKAov8IU4JjQj5KUmjnW12Rox8dj/rfGtdaH8uJ14vLgvlrAb\neQbN5Ghaxh9DGTo1337O9a9qOsir8YQqazl8ahzS2gvYleV+ou09RDhS75q9hdF2\nLI+1IvFEQ2ZO2tLk3umUP1ioa+5CWKsWD0GAXbQu9uunAV0VoExP4+/9WYOuP0ei\nKwIDAQAB\n-----END PUBLIC KEY-----\n"
  },
  "type": "Person"
}
```

与 GoToSocial 联合的外站服务器应从 `publicKey` 字段提取公钥。然后，它们应该使用公钥的 `owner` 字段签名 `GET` 请求，进一步解引用 Actor 的完整版本。

这种行为是为了避免外站服务器对完整 Actor 端点进行未签名的 `GET` 请求引入的。然而，由于不合规且引发问题，此行为可能会在未来发生变化。在 [此问题](https://codeberg.org/superseriousbusiness/gotosocial/issues/1186) 中进行跟踪。
