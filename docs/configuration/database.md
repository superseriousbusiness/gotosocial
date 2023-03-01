# Database

GoToSocial stores statuses, accounts, etc, in a database. This can be either [SQLite](https://sqlite.org/index.html) or [Postgres](https://www.postgresql.org/).

By default, GoToSocial will use Postgres, but this is easy to change.

## SQLite

SQLite, as the name implies, is the lightest database type that GoToSocial can use. It stores entries in a simple file format, usually in the same directory as the GoToSocial binary itself. SQLite is great for small instances and lower-powered machines like Raspberry Pi, where a dedicated database would be overkill.

To configure GoToSocial to use SQLite, change `db-type` to `sqlite`. The `address` setting will then be a filename instead of an address, so you will want to change it to `sqlite.db` or something similar.

Note that the `:memory:` setting will use an *in-memory database* which will be wiped when your GoToSocial instance stops running. This is for testing only and is absolutely not suitable for running a proper instance, so *don't do this*.

## Postgres

Postgres is a heavier database format, which is useful for larger instances where you need to scale performance, or where you need to run your database on a dedicated machine separate from your GoToSocial instance (or do funky stuff like run a database cluster).

You can connect to Postgres using either a Unix socket connection, or via TCP, depending on what you've set as your `db-address` value.

GoToSocial also supports connecting to Postgres using SSL/TLS over TCP. If you're running Postgres on a different machine from GoToSocial, and connecting to it via an IP address or hostname (as opposed to just running on localhost), then SSL/TLS is **CRUCIAL** to avoid leaking data all over the place!

When you're using Postgres, GoToSocial expects whatever you've set for `db-user` to already be created in the database, and to have ownership of whatever you've set for `db-database`.

For example, if you set:

```text
db:
  [...]
  user: "gotosocial"
  password: "some_really_good_password"
  database: "gotosocial"  
```

Then you should have already created database `gotosocial` in Postgres, and given ownership of it to the `gotosocial` user.

The psql commands to do this will look something like:

```psql
create database gotosocial with locale C.UTF-8 template template0;
create user gotosocial with password 'some_really_good_password';
grant all privileges on database gotosocial to gotosocial;
```

GoToSocial makes use of ULIDs (Universally Unique Lexicographically Sortable Identifiers) which will not work in non-English collate environments. For this reason it is important to create the database with `C.UTF-8` locale. To do that on systems which were already initialized with non-C locale, `template0` pristine database template must be used.

## Settings

```yaml
############################
##### DATABASE CONFIG ######
############################

# Config pertaining to the Gotosocial database connection

# String. Database type.
# Options: ["postgres","sqlite"]
# Default: "postgres"
db-type: "postgres"

# String. Database address or parameters.
#
# For Postgres, this should be the address or socket at which the database can be reached.
#
# For Sqlite, this should be the path to your sqlite database file. Eg., /opt/gotosocial/sqlite.db.
# If the file doesn't exist at the specified path, it will be created.
# If just a filename is provided (no directory) then the database will be created in the same directory
# as the GoToSocial binary.
# If address is set to :memory: then an in-memory database will be used (no file).
# WARNING: :memory: should NOT BE USED except for testing purposes.
#
# Examples: ["localhost","my.db.host","127.0.0.1","192.111.39.110",":memory:", "sqlite.db"]
# Default: ""
db-address: ""

# Int. Port for database connection.
# Examples: [5432, 1234, 6969]
# Default: 5432
db-port: 5432

# String. Username for the database connection.
# Examples: ["mydbuser","postgres","gotosocial"]
# Default: ""
db-user: ""

# String. Password to use for the database connection
# Examples: ["password123","verysafepassword","postgres"]
# Default: ""
db-password: ""

# String. Name of the database to use within the provided database type.
# Examples: ["mydb","postgres","gotosocial"]
# Default: "gotosocial"
db-database: "gotosocial"

# String. Disable, enable, or require SSL/TLS connection to the database.
# If "disable" then no TLS connection will be attempted.
# If "enable" then TLS will be tried, but the database certificate won't be checked (for self-signed certs).
# If "require" then TLS will be required to make a connection, and a valid certificate must be presented.
# Options: ["disable", "enable", "require"]
# Default: "disable"
db-tls-mode: "disable"

# String. Path to a CA certificate on the host machine for db certificate validation.
# If this is left empty, just the host certificates will be used.
# If filled in, the certificate will be loaded and added to host certificates.
# Examples: ["/path/to/some/cert.crt"]
# Default: ""
db-tls-ca-cert: ""

# Int. Number to multiply by CPU count to set permitted total of open database connections (in-use and idle).
# You can use this setting to tune your database connection behavior, though most admins won't need to touch it.
#
# Example values for multiplier 8:
#
# 1 cpu = 08 open connections
# 2 cpu = 16 open connections
# 4 cpu = 32 open connections
#
# Example values for multiplier 4:
#
# 1 cpu = 04 open connections
# 2 cpu = 08 open connections
# 4 cpu = 16 open connections
#
# A multiplier of 8 is a sensible default, but you may wish to increase this for instances
# running on very performant hardware, or decrease it for instances using v. slow CPUs.
#
# If you set the multiplier to less than 1, only one open connection will be used regardless of cpu count.
#
# PLEASE NOTE!!: This setting currently only applies for Postgres. SQLite will always use 1 connection regardless
# of what is set here. This behavior will change in future when we implement better SQLITE_BUSY handling.
# See https://github.com/superseriousbusiness/gotosocial/issues/1407 for more details.
#
# Examples: [16, 8, 10, 2]
# Default: 8
db-max-open-conns-multiplier: 8

# String. SQLite journaling mode.
# SQLite only -- unused otherwise.
# If set to empty string, the sqlite default will be used.
# See: https://www.sqlite.org/pragma.html#pragma_journal_mode
# Examples: ["DELETE", "TRUNCATE", "PERSIST", "MEMORY", "WAL", "OFF"]
# Default: "WAL"
db-sqlite-journal-mode: "WAL"

# String. SQLite synchronous mode.
# SQLite only -- unused otherwise.
# If set to empty string, the sqlite default will be used.
# See: https://www.sqlite.org/pragma.html#pragma_synchronous
# Examples: ["OFF", "NORMAL", "FULL", "EXTRA"]
# Default: "NORMAL"
db-sqlite-synchronous: "NORMAL"

# Byte size. SQlite cache size.
# SQLite only -- unused otherwise.
# If set to empty string or zero, the sqlite default (2MiB) will be used.
# See: https://www.sqlite.org/pragma.html#pragma_cache_size
# Examples: ["0", "2MiB", "8MiB", "64MiB"]
# Default: "8MiB"
db-sqlite-cache-size: "8MiB"

# Duration. SQlite busy timeout.
# SQLite only -- unused otherwise.
# If set to empty string or zero, the sqlite default will be used.
# See: https://www.sqlite.org/pragma.html#pragma_busy_timeout
# Examples: ["0s", "1s", "30s", "1m", "5m"]
# Default: "5s"
db-sqlite-busy-timeout: "5m"
```
