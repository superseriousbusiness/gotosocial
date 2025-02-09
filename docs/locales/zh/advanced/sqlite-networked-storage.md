# 网络存储上的 SQLite

SQLite 的运行模式假定数据库和使用它的进程或应用程序位于同一主机上。在运行 WAL 模式（GoToSocial 的默认模式）时，它依赖于进程之间的共享内存来确保数据库完整性。

!!! quote "参考"
    所有使用数据库的进程必须在同一台主机计算机上；WAL 不能在网络文件系统上工作。这是因为 WAL 需要所有进程共享少量内存，而在不同主机上的进程显然不能相互共享内存。

    — SQLite.org [写前日志](https://www.sqlite.org/wal.html)

这也意味着访问数据库的任何其他进程需要在相同的命名空间或容器上下文中运行。

理论上，可以通过 Samba、NFS、iSCSI 或其他形式的网络访问文件系统运行 SQLite。但无论是否使用写前日志模式，SQLite 维护者都不推荐或不支持这样做。这样做会使你的数据库面临损坏的风险。长期以来，网络存储在其锁定原语中存在同步问题，实现的保证也比本地存储更弱。

你的云供应商的外部卷，如 Hetzner 云存储卷、AWS EBS、GCP 持久磁盘等，也可能导致问题，并增加不确定的延迟。这往往会严重降低 SQLite 的性能。

如果你打算通过网络访问数据库，最好使用具有客户端-服务器架构的数据库。GoToSocial 支持这种用例的 Postgres。

如果想要在耐久的长期存储上保留 SQLite 数据库的副本，请参阅 [SQLite 流式副本](replicating-sqlite.md)。请记住，无论是还是副本使用网络文件系统都不能替代[备份](../admin/backup_and_restore.md)。

## 设置

!!! danger "数据库损坏"
    我们不支持在网络文件系统上使用 SQLite 运行 GoToSocial，如果你因此损坏了数据库，我们将无法帮助你。

如果你确实想冒这个风险，你需要调整 SQLite 的 [synchronous][sqlite-sync] 模式和 [journal][sqlite-journal] 模式以适应文件系统的限制。

[sqlite-sync]: https://www.sqlite.org/pragma.html#pragma_synchronous
[sqlite-journal]: https://www.sqlite.org/pragma.html#pragma_journal_mode

你需要更新以下设置：

* `db-sqlite-journal-mode`
* `db-sqlite-synchronous`

我们不提供任何建议，因为这将根据你使用的解决方案而有所不同。请参阅 [此问题](https://github.com/superseriousbusiness/gotosocial/issues/3360#issuecomment-2380332027)以了解你可能设置的值。
