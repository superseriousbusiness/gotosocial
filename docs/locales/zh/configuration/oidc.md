# OpenID Connect (OIDC)

GoToSocial 支持 [OpenID Connect](https://openid.net/connect/)，这是一种基于 [OAuth 2.0](https://oauth.net/2/) 构建的身份验证协议，OAuth 2.0 是授权协议的行业标准协议之一。

这意味着你可以将 GoToSocial 连接到外部 OIDC 提供商，如 [Gitlab](https://docs.gitlab.com/ee/integration/openid_connect_provider.html)、[Google](https://cloud.google.com/identity-platform/docs/web/oidc)、[Keycloak](https://www.keycloak.org/) 或 [Dex](https://dexidp.io/)，并允许用户使用其凭据登录 GoToSocial。

在以下情况下，这非常方便：

- 你在一个平台上运行多个服务（Matrix、Peertube、GoToSocial），并希望用户可以使用相同的登录页面登录所有服务。
- 你希望将用户、账户、密码等的管理委托给外部服务，以简化管理。
- 你已经在外部系统中有很多用户，不想在 GoToSocial 中手动重新创建他们。

!!! tip "提示"
    如果用户尚不存在，且你的 IdP 没有返回非空的 `email` 作为 claims 的一部分，登录将会失败。这个 email 需要在此实例中是唯一的。尽管我们使用 `sub` claim 将外部身份与 GtS 用户关联，但创建用户时需要一个与之关联的 email。

## 设置

GoToSocial 为 OIDC 提供以下配置设置，以下显示的是其默认值。

```yaml
#######################
##### OIDC CONFIG #####
#######################

# 配置与外部 OIDC 提供商（如 Dex、Google、Auth0 等）的身份验证。

# 布尔值。启用与外部 OIDC 提供商的身份验证。如果设置为 true，则其他 OIDC 选项也必须设置。
# 如果设置为 false，则使用标准的内部 OAuth 流程，用户使用用户名/密码登录 GtS。
# 选项: [true, false]
# 默认值: false
oidc-enabled: false

# 字符串。oidc idp（身份提供商）的名称。这将在用户登录时显示。
# 示例: ["Google", "Dex", "Auth0"]
# 默认值: ""
oidc-idp-name: ""

# 布尔值。跳过对从 OIDC 提供商返回的令牌的正常验证流程，即不检查过期或签名。
# 这应仅用于调试或测试，绝对不要在生产环境中使用，因为这极其不安全！
# 选项: [true, false]
# 默认值: false
oidc-skip-verification: false

# 字符串。OIDC 提供商 URI。这是 GtS 将用户重定向到的登录地址。
# 通常这看起来像是一个标准的网页 URL。
# 示例: ["https://auth.example.org", "https://example.org/auth"]
# 默认值: ""
oidc-issuer: ""

# 字符串。在 OIDC 提供商处注册的此客户端的 ID。
# 示例: ["some-client-id", "fda3772a-ad35-41c9-9a59-f1943ad18f54"]
# 默认值: ""
oidc-client-id: ""

# 字符串。在 OIDC 提供商处注册的此客户端的密钥。
# 示例: ["super-secret-business", "79379cf5-8057-426d-bb83-af504d98a7b0"]
# 默认值: ""
oidc-client-secret: ""

# 字符串数组。向 OIDC 提供商请求的范围。返回的值将用于填充在 GtS 中创建的用户。
# 'openid' 和 'email' 是必需的。
# 'profile' 用于提取新创建用户的用户名。
# 'groups' 是可选的，可以用于根据 oidc-admin-groups 确定用户是否为管理员。
# 示例: 见 eg., https://auth0.com/docs/scopes/openid-connect-scopes
# 默认值: ["openid", "email", "profile", "groups"]
oidc-scopes:
  - "openid"
  - "email"
  - "profile"
  - "groups"

# 布尔值。将通过 OIDC 进行身份验证的用户与现有用户基于其电子邮件地址进行关联。
# 这主要用于迁移目的，即从使用不稳定 `email` claim 进行唯一用户标识的旧版 GtS 迁移。对于大多数用例，应设置为 false。
# 选项: [true, false]
# 默认值: false
oidc-link-existing: false

# 字符串数组。如果返回的 ID 令牌包含与 oidc-allowed-groups 中的某个组匹配的 'groups' claim，则该用户将在 GtS 实例上被授予访问权限。
# 如果数组为空，则授予所有组权限。
# 默认值: []
oidc-allowed-groups: []

# 字符串数组。如果返回的 ID 令牌包含与 oidc-admin-groups 中的某个组匹配的 'groups' claim，则该用户将在 GtS 实例上被授予管理员权限。
# 默认值: []
oidc-admin-groups: []
```

## 行为

在 GoToSocial 上启用 OIDC 后，默认登录页面会自动重定向到 OIDC 提供商的登录页面。

这意味着 OIDC 本质上 *替代* 了正常的 GtS 邮箱/密码登录流程。

由于 ActivityPub 标准的工作方式，你 _不能_ 在设置用户名后更改它。这与 OIDC 规范冲突，该规范不保证 `preferred_username` 字段是稳定的。

为了解决这个问题，我们要求用户在首次登录尝试时提供用户名。此字段预先填入 `preferred_username` claim 的值。

在认证后，GtS 存储由 OIDC 提供商提供的 `sub` claim。在后续的身份验证尝试中，这个 claim 被用作唯一的用户查找方式。

这使你可以在提供商级别更改用户名而不丢失对 GtS 账户的访问。

### 群组成员身份

大多数 OIDC 提供商允许在返回的 claims 中包含群组和群组成员身份的概念。GoToSocial 可以使用群组成员身份来确定从 OIDC 流中返回的用户是否应创建为管理员账户。

如果返回的 OIDC 群组信息中包含配置在 `oidc-admin-groups` 中的群组成员身份，则该用户将被创建/登录为管理员。

## 从旧版本迁移

如果你从使用不稳定 `email` claim 进行唯一用户标识的旧版 GtS 迁移过来，可以将 `oidc-link-existing` 配置设置为 `true`。如果无法为提供商返回的 ID 找到用户，则会根据 `email` claim 进行查找。如果成功，稳定 ID 将被添加到匹配的用户数据库中。

你应仅在有限时间内使用此功能，以避免恶意账户盗取。

## 提供商示例

### Dex

[Dex](https://dexidp.io/) 是一个可以自行托管的开源 OIDC 提供商。安装 Dex 的过程不在本文档范围内，你可以在 [这里](https://dexidp.io/docs/) 查看 Dex 文档。

Dex 的优势在于它也用 Go 编写，像 GoToSocial 一样，这意味着它体积小、运行快，在低功耗系统上也能很好地运行。

要配置 Dex 和 GoToSocial 一起工作，在 Dex 配置的 `staticClients` 部分创建以下客户端：

```yaml
staticClients:
  - id: 'gotosocial_client'
    redirectURIs:
      - 'https://gotosocial.example.org/auth/callback'
    name: 'GoToSocial'
    secret: 'some-client-secret'
```

确保将 `gotosocial_client` 替换为你想要的客户端 ID，并将 `secret` 替换为一个合理长且安全的密钥（例如 UUID）。你还应该将 `gotosocial.example.org` 替换为 GtS 实例的 `host`，但保留 `/auth/callback` 不变。

然后，编辑 GoToSocial config.yaml 中的 `oidc` 部分如下：

```yaml
oidc:
  enabled: true
  idpName: "Dex"
  skipVerification: false
  issuer: "https://auth.example.org"
  clientID: "gotosocial_client"
  clientSecret: "some-client-secret"
  scopes:
    - "openid"
    - "email"
    - "profile"
    - "groups"
```

确保将 `issuer` 变量替换为你的 Dex 提供商设置。这应该是你的 Dex 实例的可访问到的确切 URI。

现在，重启 GoToSocial 和 Dex，以便新设置生效。

当你下次登录 GtS 时，你应该会被重定向到 Dex 的登录页面。登录成功后，你将返回到 GoToSocial。
