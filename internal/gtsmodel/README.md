# gtsmodel

This package contains types used *internally* by GoToSocial and added/removed/selected from the database. As such, they contain sensitive fields which should **never** be serialized or reach the API level. Use the [mastotypes](../../pkg/mastotypes) package for that.

The annotation used on these structs is for handling them via the go-pg ORM. See [here](https://pg.uptrace.dev/models/).
