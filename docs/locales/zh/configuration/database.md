# 数据库

GoToSocial 将贴文、账号等存储在数据库中。可以选择使用 [SQLite](https://sqlite.org/index.html) 或 [Postgres](https://www.postgresql.org/)。

默认情况下，GoToSocial 使用 Postgres，但可以轻松更改。

## SQLite

顾名思义，SQLite 是 GoToSocial 可用的最轻量级的数据库类型。它以简单的文件格式存储条目，通常与 GoToSocial 二进制文件位于同一目录。SQLite 非常适合小规模实例和单板计算机，使用专用数据库对它们来说过于复杂。

要配置 GoToSocial 使用 SQLite，将 `db-type` 更改为 `sqlite`。此时 `address` 设置将是一个文件名而不是地址，所以你需要将其更改为 `sqlite.db` 或类似名称。

注意，`:memory:` 设置将使用 *内存数据库*，当你的 GoToSocial 实例停止运行时，内存将被清除。这仅用于测试，绝不适用于运行正式实例，因此*不要这样做*。

## Postgres

Postgres 是较重的数据库格式，适用于需要扩展性能的大型实例，或者需要在 GoToSocial 实例之外的专用计算机上运行数据库的情况（或运行数据库集群等复杂应用）。

你可以使用 Unix 套接字连接或 TCP 连接到 Postgres，这取决于你设置的 `db-address` 值。

GoToSocial 还支持使用 SSL/TLS 通过 TCP 连接到 Postgres。如果你在不同的计算机上运行 Postgres，并通过 IP 地址或主机名连接它（而不是仅在本地运行），那么 SSL/TLS **至关重要**，以防止数据泄露！

使用 Postgres 时，GoToSocial 期望你已经在数据库中创建了 `db-user` 并拥有 `db-database` 的所有权。

例如，如果你设置了：

```text
db:
  [...]
  user: "gotosocial"
  password: "some_really_good_password"
  database: "gotosocial"  
```

那么你应该已经在 Postgres 中创建了数据库 `gotosocial`，并将其所有权授予 `gotosocial` 用户。

执行这些操作的 psql 命令如下：

```psql
create database gotosocial with locale 'C.UTF-8' template template0;
create user gotosocial with password 'some_really_good_password';
grant all privileges on database gotosocial to gotosocial;
```

如果开始使用时你使用的是 Postgres 14 或更高版本，或者遇到 `error executing command: error creating dbservice: db migration error: ERROR: permission denied for schema public`，你应当授予你的 db 用户 `CREATE` 权限 *(这 **必须** 在连接到 gotosocial 数据库的 postgres shell 中执行)*:

```psql
GRANT CREATE ON SCHEMA public TO gotosocial;
SELECT has_schema_privilege('gotosocial', 'public', 'CREATE'); -- should return t
```

GoToSocial 使用 ULIDs（全局唯一且按字典顺序可排序的标识符），这在非英文排序环境中不起作用。因此，创建数据库时使用 `C.UTF-8` 地区设置很重要。在已经使用非 C 地区初始化的系统上，必须使用 `template0` 原始数据库模板才能进行。

如果你希望使用特定选项连接到 Postgres，可以使用 `db-postgres-connection-string` 定义连接字符串。如果 `db-postgres-connection-string` 已定义，则所有其他与数据库相关的配置字段将被忽略。例如，可以使用 `db-postgres-connection-string` 连接到 `mySchema`，用户名为 `myUser`，密码为 `myPass`，在 `localhost` 上，数据库名称为 `db`：

```yaml
db-postgres-connection-string: 'postgres://myUser:myPass@localhost/db?search_path=mySchema'
```

## 设置

```yaml
############################
##### 数据库配置 ######
############################

# GoToSocial 数据库连接的相关配置

# 字符串。数据库类型。
# 选项: ["postgres","sqlite"]
# 默认: "postgres"
db-type: "postgres"

# 字符串。数据库地址或参数。
#
# 对于 Postgres，这应该是数据库可以访问的地址或套接字。
#
# 对于 Sqlite，这应该是你的 sqlite 数据库文件的路径。比如，/opt/gotosocial/sqlite.db。
# 如果在指定路径不存在该文件，会自动创建。
# 如果只提供了文件名（没有目录），那么数据库将创建在 GoToSocial 二进制文件的同一目录中。
# 如果 `address` 设置为 :memory:，将使用内存数据库（没有文件）。
# 警告: :memory: 应该仅用于测试目的，不应在其他情况下使用。
#
# 示例: ["localhost","my.db.host","127.0.0.1","192.111.39.110",":memory:", "sqlite.db"]
# 默认: ""
db-address: ""

# 整数。数据库连接的端口。
# 示例: [5432, 1234, 6969]
# 默认: 5432
db-port: 5432

# 字符串。数据库连接的用户名。
# 示例: ["mydbuser","postgres","gotosocial"]
# 默认: ""
db-user: ""

# 字符串。数据库连接使用的密码
# 示例: ["password123","verysafepassword","postgres"]
# 默认: ""
db-password: ""

# 字符串。要在提供的数据库类型中使用的数据库名称。
# 示例: ["mydb","postgres","gotosocial"]
# 默认: "gotosocial"
db-database: "gotosocial"

# 字符串。禁用、启用或要求数据库的 SSL/TLS 连接。
# 如果为 "disable"，则不会尝试 TLS 连接。
# 如果为 "enable"，则会尝试 TLS，但不会检查数据库证书（适用于自签名证书）。
# 如果为 "require"，则需要 TLS 进行连接，并且必须提供有效证书。
# 选项: ["disable", "enable", "require"]
# 默认: "disable"
db-tls-mode: "disable"

# 字符串。用于数据库证书验证的主机的 CA 证书路径。
# 如果留空，仅使用主机证书。
# 如果填写，则会加载证书并添加到主机证书中。
# 示例: ["/path/to/some/cert.crt"]
# 默认: ""
db-tls-ca-cert: ""

# 整数。乘以 CPU 数量以设置允许总数的打开数据库连接（使用和空闲）。
# 你可以使用此设置来调整你的数据库连接行为，但大多数管理员不需要更改它。
#
# 乘数 8 的示例值：
#
# 1 cpu = 08 打开的连接
# 2 cpu = 16 打开的连接
# 4 cpu = 32 打开的连接
#
# 乘数 4 的示例值：
#
# 1 cpu = 04 打开的连接
# 2 cpu = 08 打开的连接
# 4 cpu = 16 打开的连接
#
# 乘数 8 是一个合理的默认值，但你可能希望为在非常高性能硬件上运行的实例增加此值，或为使用非常慢的 CPU 的实例减少此值。
#
# 请注意！！：此设置目前仅适用于 Postgres。SQLite 将始终使用 1 个连接，无论此处设置为何。这种行为将在实现更好的 SQLITE_BUSY 处理时更改。
# 更多详情请参见 https://codeberg.org/superseriousbusiness/gotosocial/issues/1407。
#
# 示例: [16, 8, 10, 2]
# 默认: 8
db-max-open-conns-multiplier: 8

# 字符串。SQLite 日志模式。
# 仅适用于 SQLite -- 否则不使用。
# 如果设置为空字符串，则使用 sqlite 默认值。
# 参见: https://www.sqlite.org/pragma.html#pragma_journal_mode
# 示例: ["DELETE", "TRUNCATE", "PERSIST", "MEMORY", "WAL", "OFF"]
# 默认: "WAL"
db-sqlite-journal-mode: "WAL"

# 字符串。SQLite 同步模式。
# 仅适用于 SQLite -- 否则不使用。
# 如果设置为空字符串，则使用 sqlite 默认值。
# 参见: https://www.sqlite.org/pragma.html#pragma_synchronous
# 示例: ["OFF", "NORMAL", "FULL", "EXTRA"]
# 默认: "NORMAL"
db-sqlite-synchronous: "NORMAL"

# 字节大小。SQlite 缓存大小。
# 仅适用于 SQLite -- 否则不使用。
# 如果设置为空字符串或零，则使用 sqlite 默认值（2MiB）。
# 参见: https://www.sqlite.org/pragma.html#pragma_cache_size
#
# 缓存并非越大越好。它们需要针对工作负载进行调整。默认设置对于大多数实例应该已足够，不应该更改。
# 如果你确实更改它，请确保在 GoToSocial 帮助频道求助时提到这一点。
#
# 示例: ["0", "2MiB", "8MiB", "64MiB"]
# 默认: "8MiB"
db-sqlite-cache-size: "8MiB"

# 持续时间。SQlite 忙等待时间。
# 仅适用于 SQLite -- 否则不使用。
# 如果设置为空字符串或零，则使用 sqlite 默认值。
# 参见: https://www.sqlite.org/pragma.html#pragma_busy_timeout
# 示例: ["0s", "1s", "30s", "1m", "5m"]
# 默认: "30m"
db-sqlite-busy-timeout: "30m"

# 字符串。完整的数据库连接字符串
#
# 此连接字符串仅适用于 Postgres。当定义此字段时，所有其他与数据库相关的配置字段将被忽略。
# 此字段允许你微调与 Postgres 的连接。
# 
# 示例: ["postgres://user:pass@localhost/db?search_path=gotosocial", "postgres://user:pass@localhost:9999/db"]
# 默认: ""
db-postgres-connection-string: ""

cache:
  # cache.memory-target 设置一个目标限制，
  # 应用程序将尝试将其缓存保持在此限制内。
  # 这是基于内存对象的估计大小，因此绝对不精确。
  # 示例: ["100MiB", "200MiB", "500MiB", "1GiB"]
  # 默认: "100MiB"
  memory-target: "100MiB"
```