# 密码管理

## 更改你的密码

你可以使用[用户设置面板](./settings.md)来更改密码。只需登录用户设置面板，滑动到页面底部，输入你的旧密码和想要的新密码即可。

如果你提供的新密码不够长或不够复杂，设置面板会报错并提示你尝试其他密码。

如果你的实例使用 OIDC（即通过 Google 或其他外部提供商登录），你需要通过 OIDC 提供商，而不是通过用户设置面板更改密码。

## 密码存储

GoToSocial 使用安全的 [bcrypt](https://en.wikipedia.org/wiki/Bcrypt) 函数，通过 [Go 标准库](https://pkg.go.dev/golang.org/x/crypto/bcrypt)在数据库中存储用户密码的哈希值。

这意味着，即使你的 GoToSocial 实例数据库遭到破坏，你的密码明文也是安全的。这也意味着你的实例管理员无法访问你的密码。

为了在接受前检查密码是否足够安全，GoToSocial 使用[这个库](https://github.com/wagslane/go-password-validator)，其熵值设置为 60。这意味着像 `password` 这样的密码会被拒绝，但像 `verylongandsecurepasswordhahaha` 这样的密码会被接受，即使没有特殊字符/大写和小写字母等。

我们建议遵循 EFF 关于[创建强密码](https://ssd.eff.org/en/module/creating-strong-passwords)的指南。
