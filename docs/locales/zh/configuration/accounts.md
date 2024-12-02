# 账户

## 设置

```yaml
###########################
##### 账户配置 #####
###########################

# 服务器上账户创建与维护的配置，以及新账户的默认设置。

# 布尔值。允许人们通过 /signup 表单提交新的注册请求。
#
# 选项: [true, false]
# 默认: false
accounts-registration-open: false

# 布尔值。注册请求是否需要提交请求理由（例如，解释他们为何想加入此实例）？
# 选项: [true, false]
# 默认: true
accounts-reason-required: true

# 布尔值。允许此实例上的账户为其个人资料页面和贴文设置自定义 CSS。
# 启用此设置将允许账户通过 /user 设置页面上传自定义 CSS，
# 然后这些 CSS 将在账户的个人资料和贴文的网页视图中呈现。
#
# 对于允许公开注册的实例，**强烈建议**将此设置保持为 'false'，
# 因为设置为 true 允许恶意账户使其个人资料页面具有误导性、不可用
# 或对访问者甚至危险。换句话说，只有在你信任实例上的用户不会产生有害 CSS 时，
# 才应启用此设置。
#
# 无论此值设置为何，任何上传的 CSS 都不会联合到其他实例，仅在*本*实例上的个人资料和贴文中显示。
#
# 选项: [true, false]
# 默认: false
accounts-allow-custom-css: false

# 整数值。如果 accounts-allow-custom-css 为 true，则为此实例上账户上传的
# CSS 允许的字符长度。如果 accounts-allow-custom-css 为 false，则无效。
#
# 示例: [500, 5000, 9999]
# 默认: 10000
accounts-custom-css-length: 10000
```
