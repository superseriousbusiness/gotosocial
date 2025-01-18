# 访问控制

GoToSocial 使用访问控制限制来保护用户和资源免受外站账户和实例的不必要互动。

如[HTTP 签名](#http-signatures)部分所示，GoToSocial 要求所有来自外站服务器的传入 `GET` 和 `POST` 请求必须签名。未签名的请求将被拒绝，并返回 http 代码 `401 Unauthorized`。

访问控制限制通过检查签名的 `keyId` （即谁拥有发出请求的公钥/私钥对）来实现。

首先，会将 `keyId` URI 的主机值与 GoToSocial 实例的已屏蔽（取消联合）的域列表进行检查。如果域名被检测到位于屏蔽列表，则 http 请求将立即以 http 代码 `403 Forbidden` 中止。

接下来，GoToSocial 将检查发出 http 请求的公钥所有者与请求目标资源的所有者之间是否存在（任一方向的）屏蔽。如果 GoToSocial 用户屏蔽了发出请求的外站账户，则请求将以 http 代码 `403 Forbidden` 中止。
