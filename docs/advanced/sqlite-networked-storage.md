# SQLite on networked storage

SQLite's operating model assumes the database and the processes or applications using it are colocated on the same host. When running the database in WAL-mode, which is GoToSocial's default, it relies on shared memory between processes to ensure the integrity of your database.

!!! quote
    All processes using a database must be on the same host computer; WAL does not work over a network filesystem. This is because WAL requires all processes to share a small amount of memory and processes on separate host machines obviously cannot share memory with each other. 

    — SQLite.org [Write-Ahead Logging](https://www.sqlite.org/wal.html)

This also means that any other processes accessing the database need to run in the same namespace or container context.

It is in theory possible to run SQLite over Samba, NFS, iSCSI or other forms of filesystems accessed over the network. But it is neither recommended nor supported by the SQLite maintainers, irrespective of whether you're running with write-ahead logging or not. Doing so puts you at risk of database corruption. There is a long history of networked storage having synchronisation issues in their locking primitives, or implementing them with weaker guarantees than what a local filesystem can provide.

Your cloud provider's external volumes, like Hetzner Cloud Volumes, AWS EBS, GCP Persistent Disk etc. may also cause problems, and add variable latency. This has a tendency to severely degrade SQLite's performance.

If you're going to access your database over the network, it's better to use a database with a client-server architecture. GoToSocial supports Postgres for such use-cases.

For the purpose of having a copy of the SQLite database on durable long-term storage, refer to [SQLite streaming replication](replicating-sqlite.md) instead. Remember that neither replication nor using a networked filesystem are a substitute [for having backups](../admin/backup_and_restore.md).

## Settings

!!! danger "Corrupted database"
    We do not support running GoToSocial with SQLite on a networked filesystem and we will not be able to help you if you damage your database this way.

Should you really want to take this risk, you'll need to adjust the SQLite [synchronous][sqlite-sync] mode and [journal][sqlite-journal] mode to match the limitations of the filesystem.

[sqlite-sync]: https://www.sqlite.org/pragma.html#pragma_synchronous
[sqlite-journal]: https://www.sqlite.org/pragma.html#pragma_journal_mode

You'll need to update the following settings:

* `db-sqlite-journal-mode`
* `db-sqlite-synchronous`

We don't provide any recommendations as this will vary based on the solution you're using. See [this issue](https://github.com/superseriousbusiness/gotosocial/issues/3360#issuecomment-2380332027) for what you could potentially set those values to.
