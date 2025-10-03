# Go `memdb` SQLite VFS

This package implements the [`"memdb"`](https://sqlite.org/src/doc/tip/src/memdb.c)
SQLite VFS in pure Go.

It has some benefits over the C version:
- the memory backing the database needs not be contiguous,
- the database can grow/shrink incrementally without copying,
- reader-writer concurrency is slightly improved.

[`memdb.TestDB`](https://pkg.go.dev/github.com/ncruces/go-sqlite3/vfs/memdb#TestDB)
is the preferred way to setup an in-memory database for testing.