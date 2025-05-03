# 使用 API 进行身份验证

使用客户端 API 需要进行身份验证。本页记录了如何获取身份验证令牌的通用流程，并提供了使用 `curl` 在命令行界面进行操作的示例。

!!! tip "提示"
    如果你不想使用命令行，而是想通过设置面板获取 API 访问令牌，可以参考 [应用文档](https://docs.gotosocial.org/zh-cn/latest/user_guide/settings/#applications)。

## 创建新应用

我们需要注册一个新应用，以便请求 OAuth 令牌。这可以通过向 `/api/v1/apps` 端点发送 `POST` 请求来完成。注意将下面命令中的 `your_app_name` 替换为你想使用的应用名称：

```bash
curl \
  -H 'Content-Type:application/json' \
  -d '{
        "client_name": "your_app_name",
        "redirect_uris": "urn:ietf:wg:oauth:2.0:oob",
        "scopes": "read"
      }' \
  'https://example.org/api/v1/apps'
```

字符串 `urn:ietf:wg:oauth:2.0:oob` 表示一种称为带外身份验证的技术，这是一种用于多因素身份验证的技术，旨在减少恶意行为者干扰身份验证过程的途径。在此情况下，它允许我们查看并手动复制生成的令牌以便继续使用。

!!! tip "权限范围"
    根据应用执行的工作对应用进行最低限度的授权是一个好习惯。例如，如果你的应用不会发布贴文，请使用 `scope=read` 或进一步仅授权子权限。
    
    本着这种精神，上例使用了`read`，这意味着应用将仅限于执行`read`操作。
    
    可用范围列表请参阅[Swagger 文档](https://docs.gotosocial.org/zh-cn/latest/api/swagger/).

!!! warning "警告"
    GoToSocial 0.19.0 之前的版本并不支持范围授权令牌，因此运行低于 0.19.0 的 GoToSocial 的用户通过此流程获得的任何令牌都可以代表用户执行所有操作。如果用户具有管理权限，那么令牌还可以执行管理操作。

成功调用会返回一个带有 `client_id` 和 `client_secret` 的响应，我们将在后续流程中需要使用这些信息。它看起来像这样：

```json
{
  "id": "01J1CYJ4QRNFZD6WHQMZV7248G",
  "name": "your_app_name",
  "redirect_uri": "urn:ietf:wg:oauth:2.0:oob",
  "client_id": "YOUR_CLIENT_ID",
  "client_secret": "YOUR_CLIENT_SECRET"
}
```

!!! tip "提示"
    确保将 `client_id` 和 `client_secret` 的值保存到某个位置，以便在需要时参考。

## 授权你的应用代表你操作

我们已经在 GoToSocial 注册了一个新应用，但它尚未与你的账户连接。现在，我们需要告知 GoToSocial 这个新应用将代表你操作。为此，我们需要通过浏览器进行实例认证，以启动登录和权限授予过程。

创建一个带查询字符串的 URL，如下所示，将 `YOUR_CLIENT_ID` 替换为你在上一步收到的 `client_id`，然后将 URL 粘贴到浏览器中：

```text
https://example.org/oauth/authorize?client_id=YOUR_CLIENT_ID&redirect_uri=urn:ietf:wg:oauth:2.0:oob&response_type=code&scope=read
```

!!! tip "提示"
    如果你在注册应用时使用了不同的范围，在上面的 URL 中将 `scope=read` 替换为你注册时使用的加号分隔的范围列表。例如，如果你注册你的应用时使用了 `scopes` 值 `read write`，那么你应该将上面的 `scope=read` 改为 `scope=read+write`。

将 URL 粘贴到浏览器后，你会被引导到实例的登录表单，提示你输入邮箱地址和密码以将应用连接到你的账户。

提交凭据后，你会到达一个页面，上面写着类似这样的内容：

```
嗨嗨，`your_username`!

应用 `your_app_name` 申请以你的名义执行操作，申请的权限范围是 *`read`*.
如果选择允许，应用将跳转到： urn:ietf:wg:oauth:2.0:oob 继续操作
```

点击 `允许`，你将到达这样一个页面：

```text
Here's your out-of-band token with scope "read", use it wisely:
YOUR_AUTHORIZATION_TOKEN
```

复制带外授权令牌到某个地方，因为你将在下一步中需要它。

## 获取访问令牌

下一步是用刚刚收到的带外授权令牌交换一个可重用的访问令牌，该令牌可以在以后所有的 API 请求中发送。

你可以通过另一个 `POST` 请求来完成这项操作，如下所示：

```bash
curl \
  -H 'Content-Type: application/json' \
  -d '{
        "redirect_uri": "urn:ietf:wg:oauth:2.0:oob",
        "client_id": "YOUR_CLIENT_ID",
        "client_secret": "YOUR_CLIENT_SECRET",
        "grant_type": "authorization_code",
        "code": "YOUR_AUTHORIZATION_TOKEN"
      }' \
  'https://example.org/oauth/token'
```

确保替换：

- `YOUR_CLIENT_ID` 为第一步中收到的客户端 ID。
- `YOUR_CLIENT_SECRET` 为第一步中收到的客户端密钥。
- `YOUR_AUTHORIZATION_TOKEN` 为在第二步中收到的带外授权令牌。

你会收到一个包含访问令牌的响应，看起来像这样：

```json
{
  "access_token": "YOUR_ACCESS_TOKEN",
  "created_at": 1719577950,
  "scope": "read",
  "token_type": "Bearer"
}
```

将你的访问令牌复制并安全保存。

## 验证

为了确保一切正常，尝试查询 `/api/v1/verify_credentials` 端点，在请求头中添加你的访问令牌作为 `Authorization: Bearer YOUR_ACCESS_TOKEN`。

请参考以下示例：

```bash
curl \
  -H 'Authorization: Bearer YOUR_ACCESS_TOKEN' \
  'https://example.org/api/v1/accounts/verify_credentials'
```

如果一切顺利，你应该会得到用户资料的 JSON 响应。

## 最后说明

现在你拥有了访问令牌，可以在每次 API 请求中重复使用该令牌进行授权。你不需要每次都执行整个令牌交换过程！

例如，你可以使用相同的访问令牌向 API 发送另一个 `GET` 请求以获取通知，如下所示：

```bash
curl \
  -H 'Authorization: Bearer YOUR_ACCESS_TOKEN' \
  'https://example.org/api/v1/notifications'
```
