# 站点

## 设置

```yaml
###########################
##### 站点配置 #####
###########################

# 与实例间联合、隐藏/显示页面等相关的配置。

# 字符串数组。BCP47 语言标签，用于指示本站用户的首选语言。
#
# 如果你提供了这些标签，请按照从最优先到最次优先的顺序提供。
# 但请注意，从此数组中省略某种语言并不意味着该语言不能在本站使用，
# 而只是表示不会将其作为为本站的首选语言展示。
#
# 这里可以不提供任何条目；这样的话，你的站点将没有特定的首选语言。
#
# 常用标签参考此处：https://en.wikipedia.org/wiki/IETF_language_tag#List_of_common_primary_language_subtags
# 所有当前标签参考此处：https://www.iana.org/assignments/language-subtag-registry/language-subtag-registry
#
# 示例: ["zh", "zh-Hans", "zh-Hans-CN", "zh-Hant", "zh-Hant-TW", "en"]
# 默认值: []
instance-languages: []

# 字符串。用于本站的联合模式。
#
# "blocklist" -- 默认开放联合。只有被明确屏蔽的站点会被拒绝（除非它们被另外明确设定的允许规则放行）。
#
# "allowlist" -- 默认关闭联合。只有被明确允许的站点才能与本站互动。
#
# 关于屏蔽列表和允许列表模式的更多详细信息，请查阅文档：
# https://docs.gotosocial.org/zh-cn/latest/admin/federation_modes
#
# 选项: ["blocklist", "allowlist"]
# 默认值: "blocklist"
instance-federation-mode: "blocklist"

# 布尔值。启用针对通过联合 API 进入本站的消息的启发式垃圾过滤。无论你在此处设置什么，
# 都仍将执行基本的相关性检查，但如果你被其他站点的垃圾信息骚扰，并希望更严格地过滤掉垃圾信息，
# 可以尝试启用此设置。
#
# 这是一个实验性设置，可能会过滤掉合法信息，或未能过滤掉垃圾信息。建议仅在 
# Fediverse 出现垃圾信息潮时启用此设置，以保持站点的可用性。
#
# 判断消息是否为垃圾信息的决策基于以下启发规则，依次进行，接收者 = 收到消息的账户，发送者 = 发送消息的外站账户。
#
# 首先执行基本相关性检查
#
#  1. 接收者关注了发送者。返回 OK。
#  2. 贴文未提及接收者。返回 NotRelevant。
#
# 如果 instance-federation-spam-filter = false，则现在返回 OK。
# 否则进一步检查：
#
#  3. 接收者已锁定账号并被发送者关注。返回 OK。
#  4. 提及五人或以上。返回 Spam。
#  5. 接收者关注（请求关注）一个被提及账户。返回 OK。
#  6. 贴文有媒体附件。返回 Spam。
#  7. 贴文包含非提及、非话题标签的链接。返回 Spam。
#
# 被识别为垃圾的信息将从本站删除，不会插入数据库或主页时间线或通知中。
#
# 选项: [true, false]
# 默认值: false
instance-federation-spam-filter: false

# 布尔值。允许未经认证的用户查询 /api/v1/instance/peers?filter=open 以查看与本站联合的站点列表。
# 即使设置为 'false'，认证用户（本站成员）仍然可以查询该端点。
# 选项: [true, false]
# 默认值: false
instance-expose-peers: false

# 布尔值。允许未经认证的用户查询 /api/v1/instance/peers?filter=suspended 以查看被本站屏蔽/封禁的站点列表。
# 即使设置为 'false'，认证用户（本站成员）仍然可以查询该端点。
#
# 警告：将此变量设置为 'true' 可能导致你的站点被屏蔽列表爬虫攻击。
# 参考: https://docs.gotosocial.org/zh-cn/latest/admin/domain_blocks/#block-announce-bots
#
# 选项: [true, false]
# 默认值: false
instance-expose-suspended: false

# 布尔值。允许未经认证的用户查看 /about/suspended，
# 显示本站屏蔽/封禁站点的 HTML 渲染列表。
# 选项: [true, false]
# 默认值: false
instance-expose-suspended-web: false

# 布尔值。允许未经认证的用户查询 /api/v1/timelines/public 以查看本站的公共贴文列表。
# 即使设置为 'false'，认证用户（本站成员）仍然可以查询该端点。
# 选项: [true, false]
# 默认值: false
instance-expose-public-timeline: false

# 布尔值。此标志是否将 GoToSocial 的 ActivityPub 消息发送到收件人的共享收件箱（如果可用），
# 而不是将每条消息分别发送给应接收消息的每个主体。
#
# 当发送给共享收件箱的多个收件人时，共享收件箱传递能显著减少网络负载（例如，大型 Mastodon 实例）。
#
# 参考: https://www.w3.org/TR/activitypub/#shared-inbox-delivery
#
# 选项: [true, false]
# 默认值: true
instance-deliver-to-shared-inboxes: true

# 布尔值。此标志将在 /api/v1/instance 中包含的版本字段中注入一个 Mastodon 版本。
# 这个版本通常被 Mastodon 客户端用于 API 功能检测。通过注入一个与 Mastodon 兼容的版本，
# 可以促使那些客户端在 GoToSocial 上正常运行。
#
# 选项: [true, false]
# 默认值: false
instance-inject-mastodon-version: false

# 字符串。hh:mm 格式的 24 小时制时间。
# 示例： ["14:30", "00:00", "04:00"]
# 默认值： "23:00" （晚上11点）。
instance-subscriptions-process-from: "23:00"


# 时间间隔。表示更新订阅的周期。
# 示例：["24h", "72h", "12h"]
# 默认值： "24h"（每天一次）。
instance-subscriptions-process-every: "24h"

# 字符串。允许你自定义是否以及如何在 /api/v1|v2/instance
# 和 /nodeinfo 端点向爬虫提供统计数据。
#
# 请注意，无论你在这里进行何种设置，/api/v1|v2/instance
# 端点都不会被 robots.txt 允许抓取，因为这些是客户端
# API端点。
#
# "" / 空字符串（默认模式）: 在 instance 和 nodeinfo 端点提供准确的统计数据，
# 并在 robots.txt 中禁止爬虫抓取这些端点。这种模式相当于礼貌地
# 要求爬虫不抓取，但不能保证它们会遵从这些规则，
# 因为遗憾的是，许多爬虫甚至不会检查robots.txt。
#
# "zero": 在 instance 和 nodeinfo 端点提供全为零的统计数据，
# 并在 robots.txt 中禁止爬虫抓取这些端点。
# 这种模式阻止行为不端的爬虫收集有关您的实例的统计数据,
# 因为所有收集的值都将为0。这（在统计数据方面）
# 是保护您的实例隐私的最安全方法。
#
# "serve": 在 instance 和 nodeinfo 端点提供准确的统计数据,
# 并允许爬虫抓取这些端点。如果您希望为
# 联邦宇宙统计信息收集项目做贡献，此模式将非常有用。
#
# "baffle": 在 instance 和 nodeinfo 端点提供随机且荒谬的统计数据,
# 并在 robots.txt 中禁止爬虫抓取这些端点。
# 这种模式可以用于使不尊重 robots.txt 的爬虫感到困惑。
# 警告，此做法可能会引起不尊重 robots.txt 的爬虫开发者的怨恨，
# 因此可能会给您的实例带来风险。
#
# 选项: ["", "zero", "serve", "baffle"]
# 默认: ""
instance-stats-mode: ""
```
