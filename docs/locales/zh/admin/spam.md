# 骚扰信息过滤

为了让管理员在应对来自开放注册实例的骚扰信息时稍微轻松一些，GoToSocial 提供了一个实验性的骚扰信息过滤选项。

如果你或你的用户受到骚扰信息的轰炸，可以尝试在 `config.yaml` 中将选项 `instance-federation-spam-filter` 设置为 true。你可以在[实例配置页面](../configuration/instance.md)了解有关使用的启发算法的更多信息。

被认为是骚扰信息的消息将不会存储在你的本站实例上，也不会生成通知。

!!! warning "警告"
    骚扰信息过滤器必然是不完美的工具，因为它们可能会误判一些合法的信息为垃圾，或者确实未能抓住一些*确实*是垃圾的信息。

    启用 `instance-federation-spam-filter` 应被视为当联合网络遭遇骚扰信息攻击时的一种“加固”选项。在正常情况下，你可能希望将其关闭，以避免意外过滤掉合法信息。

!!! tip "提示"
    如果你想检查骚扰信息过滤器捕获了哪些内容（如果有的话），可以在日志中搜索 `looked like spam`。

    如果你[将 GoToSocial 作为 systemd 服务运行](../getting_started/installation/metal.md#optional-enable-the-systemd-service)，可以使用以下命令：

    ```bash
    journalctl -u gotosocial --no-pager | grep 'looked like spam'
    ```

    如果没有输出，说明过滤器中没有捕获到骚扰信息。否则，你将看到一行或多行日志，其中包含已被过滤并丢弃的贴文链接。
