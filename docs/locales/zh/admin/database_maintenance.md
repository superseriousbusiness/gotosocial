# 数据库维护

无论你选择使用 SQLite 还是 Postgres 来运行 GoToSocial，可能都需要偶尔执行一些维护工作，以保持数据库的良好运作。

!!! tip

    尽管此处提供的维护建议旨在不破坏现有数据，你还是应该在手动执行维护操作之前备份数据库。这样，如果输入错误或意外运行了不当命令，可以恢复备份并重试。

!!! danger
    
    **强烈不建议**手动创建、删除或更新 GoToSocial 数据库中的条目，这里不会提供相关命令。即使你认为自己知道在做什么，运行 `DELETE` 等语句可能会引入非常难以排查的问题。以下维护建议旨在帮助你的实例平稳运行；如果你手动进入数据库并对条目、表和索引进行修改，它们不会拯救你的数据。

## SQLite

要进行手动 SQLite 维护，你首先应该在存储 GoToSocial sqlite.db 文件的机器上安装 SQLite 命令行工具 `sqlite3`。有关 `sqlite3` 的详细信息，请参见[此处](https://sqlite.org/cli.html)。

### 分析/优化

按照 [SQLite 最佳实践](https://sqlite.org/lang_analyze.html#recommended_usage_pattern)，GoToSocial 在关闭数据库连接时运行 `optimize` SQLite pragma，`analysis_limit=1000`，以保持索引信息的更新。

在每次数据库迁移后（例如，启动新版本的 GoToSocial 时），GoToSocial 将运行 `ANALYZE`，以确保查询计划器正确考虑迁移新增或删除的索引。

`ANALYZE` 命令可能需要大约 10 分钟，具体时间取决于硬件和数据库文件的大小。

由于上述自动化步骤，正常情况下你不需要针对 SQLite 数据库文件手动运行 `ANALYZE` 命令。

然而，如果你中断了之前的 `ANALYZE` 命令，并发现查询运行缓慢，可能是因为 SQLite 内部表中存储的索引元数据已被删除或不当修改。

如果是这种情况，可以尝试手动运行完整的 `ANALYZE` 命令，步骤如下：

1. 停止 GoToSocial。
2. 在 `sqlite3` shell 中连接到你的 GoToSocial 数据库文件，运行 `PRAGMA analysis_limit=0; ANALYZE;`（这可能需要几分钟）。
3. 启动 GoToSocial。

[查看更多信息](https://sqlite.org/lang_analyze.html#approximate_analyze_for_large_databases).

### 清理（Vacuum）

GoToSocial 当前未启用 SQLite 的自动清理（auto-vacuum）。要将数据库文件重新打包到最佳大小，你可能需要定期（例如每几个月）在 SQLite 数据库上运行 `VACUUM` 命令。

可以在[此处](https://sqlite.org/lang_vacuum.html)查看有关 `VACUUM` 命令的详细信息。

基本步骤如下：

1. 停止 GoToSocial。
2. 在 `sqlite3` shell 中连接到你的 GoToSocial 数据库文件，运行 `VACUUM;`（这可能需要几分钟）。
3. 启动 GoToSocial。

### 副本

为数据库设置副本等保护措施是常见做法。SQLite 可以使用外部软件进行副本创建。基本步骤描述在 [配置 SQLite 副本](../advanced/replicating-sqlite.md) 页面。

## Postgres

待完成：Postgres 的维护建议。
