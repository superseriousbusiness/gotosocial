# 从以前的实例导入帖文

从 v0.18.0 版本开始，GoToSocial 可以导入你在之前实例中的帖文存档。

!!! tip "注意"
    此过程会为你之前的帖文创建一个 *副本*。截至 2025 年初，Fediverse 中部署的 ActivityPub 还没有提供 *迁移* 帖文的方法，也还不能使现有帖文在新位置上可被检索。如果你之前的实例停止运营，则原始帖文也会消失——但你依然会保留此处被导入的副本。

## 你可以导入的内容

- 你的独立帖文
- 你对自己帖文的回复
- 这些帖文的原始创建日期
- 随附在这些帖文中的媒体（照片、音频、视频）
- 帖文中使用的表情符号（前提是这些表情符号已经在你的实例上提供）

导入帖文的过程是有意减少打扰的：导入的帖文不会被推送到你在外站实例中的粉丝，不会插入到你实例的时间线上，也不会为订阅你的新帖文的用户生成通知。这意味着你可以导入大量帖文而不会打扰到粉丝。不过，一旦帖文导入后，你依然可以像对待其他帖文那样进行转发或分享其的链接。

## 你无法导入的内容

- 对其他账户的回复
- 对其他账户的提及
- 转发（无论是对你自己的帖文的转发还是对其他账户的帖文的转发）
- 其他账户对你帖文进行点赞或转发的记录
- 带有投票的帖文
- 日期在 Unix 纪元（1970 年 1 月 1 日）或之前的帖文

这些对其他账户互动的限制是出于礼貌的设计考虑：允许在导入的旧帖文中回复他人可能会引发混淆（例如“嘿，我记得这段对话，但不是和这个账户进行的”）。此外，如果你导入了大量对某人连续对话的回复，然后再转发它们，使得这些内容在对方的实例上可见，可能会导致该用户在原对话结束多年后收到大量提及或待处理提及通知。出于类似的原因，重放点赞或转发记录将不可避免地带来信息骚扰的问题。

## 如何导入你的帖文

当前，该过程需要借助利用 GTS API 的第三方工具。未来，我们可能会将此功能整合到 GoToSocial 内部：请关注 [issue #2](https://github.com/superseriousbusiness/gotosocial/issues/2) 以获取最新动态。

[`slurp`](https://github.com/VyrCossont/slurp)（由 GTS 开发者 Vyr Cossont 开发）可以导入来自 [Mastodon 归档](https://github.com/VyrCossont/slurp?tab=readme-ov-file#importing-a-mastodon-archive) 和 [Pixelfed 归档](https://github.com/VyrCossont/slurp?tab=readme-ov-file#importing-a-pixelfed-archive) 的帖文。请查阅 `slurp` 的文档、[Mastodon 关于导出数据的说明](https://docs.joinmastodon.org/user/moving/#export) 以及 Jeff Sikes 的文章 [“使用 Slurp 将 Pixelfed 帖文导入 GoToSocial”](https://box464.com/posts/gotosocial-slurp/) 以了解更多细节。你需要熟悉命令行基础，并提前安装 Git 和 Go 编译器。

!!! warning "警告"
    如果你从 Pixelfed 导入，请注意 Pixelfed 归档中不包含你的照片，因此在导入时，你原来的实例和账户必须仍然可用。

## 开发者

你可以通过向 POST /v1/statuses/create 发送带有 scheduled_at 参数的请求来使用 GoToSocial 的日期回溯功能。如果设置了该参数且该日期位于 *过去*，该帖文将被视为对旧帖文的导入，scheduled_at 参数的日期将用于设置帖文的创建日期和 ID。（GoToSocial 使用 [ULID](https://github.com/ulid/spec) 作为 ID，且这些 ID 可以通过字典序排序来按时间排序。）此外，该帖文不会被推送给粉丝、插入时间线，也不会触发通知。创建既往帖文的导入时，返回类型与正常发布帖文时的返回类型相同，均为一个 status 对象。

由于此过程使用的是 GTS API，原始帖文不必为 ActivityPub 活动，其来源可以是博客、 Cohost、Bluesky、Usenet 等平台。
