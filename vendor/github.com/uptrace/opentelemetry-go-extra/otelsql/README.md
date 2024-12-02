[![PkgGoDev](https://pkg.go.dev/badge/github.com/uptrace/opentelemetry-go-extra/otelsql)](https://pkg.go.dev/github.com/uptrace/opentelemetry-go-extra/otelsql)

# database/sql instrumentation for OpenTelemetry Go

[OpenTelemetry database/sql](https://uptrace.dev/get/instrument/opentelemetry-database-sql.html)
instrumentation records database queries (including `Tx` and `Stmt` queries) and reports `DBStats`
metrics.

## Installation

```shell
go get github.com/uptrace/opentelemetry-go-extra/otelsql
```

## Usage

To instrument database/sql, you need to connect to a database using the API provided by otelsql:

| sql                         | otelsql                         |
| --------------------------- | ------------------------------- |
| `sql.Open(driverName, dsn)` | `otelsql.Open(driverName, dsn)` |
| `sql.OpenDB(connector)`     | `otelsql.OpenDB(connector)`     |

```go
import (
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

db, err := otelsql.Open("sqlite", "file::memory:?cache=shared",
	otelsql.WithAttributes(semconv.DBSystemSqlite),
	otelsql.WithDBName("mydb"))
if err != nil {
	panic(err)
}

// db is *sql.DB
```

And then use context-aware API to propagate the active span via
[context](https://uptrace.dev/opentelemetry/go-tracing.html#context):

```go
var num int
if err := db.QueryRowContext(ctx, "SELECT 42").Scan(&num); err != nil {
	panic(err)
}
```

See [example](/example/) for details.

## Options

Both [otelsql.Open](https://pkg.go.dev/github.com/uptrace/opentelemetry-go-extra/otelsql#Open) and
[otelsql.OpenDB](https://pkg.go.dev/github.com/uptrace/opentelemetry-go-extra/otelsql#OpenDB) accept
the same [options](https://pkg.go.dev/github.com/uptrace/opentelemetry-go-extra/otelsql#Option):

- [WithAttributes](https://pkg.go.dev/github.com/uptrace/opentelemetry-go-extra/otelsql#WithAttributes)
  configures attributes that are used to create a span.
- [WithDBName](https://pkg.go.dev/github.com/uptrace/opentelemetry-go-extra/otelsql#WithDBName)
  configures a `db.name` attribute.
- [WithDBSystem](https://pkg.go.dev/github.com/uptrace/opentelemetry-go-extra/otelsql#WithDBSystem)
  configures a `db.system` attribute. When possible, you should prefer using WithAttributes and
  [semconv](https://pkg.go.dev/go.opentelemetry.io/otel/semconv/v1.10.0), for example,
  `otelsql.WithAttributes(semconv.DBSystemSqlite)`.

## sqlboiler

You can use otelsql to instrument [sqlboiler](https://github.com/volatiletech/sqlboiler) ORM:

```go
import (
    "github.com/uptrace/opentelemetry-go-extra/otelsql"
    semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

db, err := otelsql.Open("postgres", "dbname=fun user=abc",
    otelsql.WithAttributes(semconv.DBSystemPostgreSQL))
if err != nil {
  return err
}

boil.SetDB(db)
```

## GORM 1

You can use otelsql to instrument [GORM 1](https://v1.gorm.io/):

```go
import (
    "github.com/jinzhu/gorm"
    "github.com/uptrace/opentelemetry-go-extra/otelsql"
    semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

// gormOpen is like gorm.Open, but it uses otelsql to instrument the database.
func gormOpen(driverName, dataSourceName string, opts ...otelsql.Option) (*gorm.DB, error) {
	db, err := otelsql.Open(driverName, dataSourceName, opts...)
	if err != nil {
		return nil, err
	}
	return gorm.Open(driverName, db)
}

db, err := gormOpen("mysql", "user:password@/dbname",
    otelsql.WithAttributes(semconv.DBSystemMySQL))
if err != nil {
    panic(err)
}
```

To instrument GORM 2, use
[otelgorm](https://github.com/uptrace/opentelemetry-go-extra/tree/main/otelgorm).

## Alternatives

- https://github.com/XSAM/otelsql - different driver registration and no metrics.
- https://github.com/j2gg0s/otsql - like XSAM/otelsql but with Prometheus metrics.
