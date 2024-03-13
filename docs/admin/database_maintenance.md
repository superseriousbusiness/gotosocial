# Database Maintenance

Regardless of whether you choose to run GoToSocial with SQLite or Postgres, you may need to occasionally take maintenance steps to keep your database running well.

## SQLite

### Analyze / Optimize

Following [SQLite best practice](https://sqlite.org/lang_analyze.html#recommended_usage_pattern), GoToSocial runs the `optimize` SQLite pragma with `analysis_limit=1000` on closing database connections to keep index information up to date.

After each database migration, GoToSocial will also run `ANALYZE` with `analysis_limit=10000` to ensure that any indexes added or removed by migrations are taken into account.

As such, in normal circumstances, you should not need to run manual `ANALYZE` commands against your SQLite database file.

However, if you interrupted a previous `ANALYZE` command, it could be the case that query optimizer data stored in SQLite's internal tables has been removed, and you notice that queries are running remarkably slowly.

If this is the case, you can try manually running an `ANALYZE` command in the SQLite CLI tool, by entering: `PRAGMA analysis_limit=10000; ANALYZE;`.

It is not necessary to run a full analyze, an approximate analyze will do. [See here](https://sqlite.org/lang_analyze.html#approximate_analyze_for_large_databases) for more info.

### Vacuum

GoToSocial does not currently enable auto-vacuum for SQLite. As such, you may want to periodically (eg., every few months) run a `VACUUM` command on your SQLite database, to repack the database file to an optimal size.

You can see lots of information about the `VACUUM` command [here](https://sqlite.org/lang_vacuum.html).

The basic steps are:

1. Stop GoToSocial.
2. Run `VACUUM` using the SQLite CLI tool (this may take quite a few minutes depending on the size of your database file).
3. Start GoToSocial.

## Postgres

TODO: Maintenance recommendations for Postgres. 
