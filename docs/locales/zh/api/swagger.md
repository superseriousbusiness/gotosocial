# 路由和方法

GoToSocial 使用 [go-swagger](https://github.com/go-swagger/go-swagger) 从代码注释生成一个 V2 [OpenAPI 规范](https://swagger.io/specification/v2/)文档。

生成的 API 文档如下所示。请注意，本文档仅供参考。你将无法使用以下小部件内置的授权功能实际连接到实例或进行 API 调用。相反，你应该使用像 curl、Postman 等工具。

大多数 GoToSocial API 端点需要用户级别的 OAuth 令牌。有关如何使用 OAuth 令牌进行 API 认证的指南，请参阅[认证文档](./authentication.md)。

!!! tip "提示"
    如果你想更多地使用该规范，还可以直接查看 [swagger.yaml](./swagger.yaml)，然后将其粘贴到 [Swagger Editor](https://editor.swagger.io/) 等工具中。这样你可以尝试自动生成不同语言的 GoToSocial API 客户端（不支持，但可以尝试），或者将文档转换为 JSON 或 OpenAPI v3 规范等。更多信息请参见[这里](https://swagger.io/tools/open-source/getting-started/)。

!!! info "注意事项：上传文件"
    当使用涉及提交表单上传文件的 API 端点时（例如，媒体附件端点或表情符号上传端点等），请注意，在表单字段中 `filename` 是必需的，这是由于 GoToSocial 用于解析表单的依赖关系以及 Go 的某些特性导致的。
    
    有关更多背景信息，请参见以下问题：
    
    - [#1958](https://codeberg.org/superseriousbusiness/gotosocial/issues/1958)
    - [#1944](https://codeberg.org/superseriousbusiness/gotosocial/issues/1944)
    - [#2641](https://codeberg.org/superseriousbusiness/gotosocial/issues/2641)

<swagger-ui src="swagger.yaml"/>
