# Database Maintenance

Regardless of whether you choose to run GoToSocial with SQLite or Postgres, you may need to occasionally take maintenance steps to keep your database running well.

!!! tip
    
    Though the maintenance tips provided here are intended to be non-destructive, you should backup your database before manually performing maintenance. That way if you mistype something or accidentally run a bad command, you can restore your backup and try again.

!!! danger
    
    Manually creating, deleting, or updating entries in your GoToSocial database is **heavily discouraged**, and such commands are not provided here. Even if you think you know what you are doing, running `DELETE` statements etc. may introduce issues that are very difficult to debug. The maintenance tips below are designed to help with the smooth running of your instance; they will not save your ass if you have manually gone into your database and hacked at entries, tables, and indexes.

## SQLite

To do manual SQLite maintenance, you should first install the SQLite command line tool `sqlite3` on the same machine that your GoToSocial sqlite.db file is stored on. See [here](https://sqlite.org/cli.html) for details about `sqlite3`.

### Analyze / Optimize

Following [SQLite best practice](https://sqlite.org/lang_analyze.html#recommended_usage_pattern), GoToSocial runs the `optimize` SQLite pragma with `analysis_limit=1000` on closing database connections to keep index information up to date.

After each set of database migrations (eg., when starting a newer version of GoToSocial), GoToSocial will run `ANALYZE` to ensure that any indexes added or removed by migrations are taken into account correctly by the query planner.

The `ANALYZE` command may take ~10 minutes depending on your hardware and the size of your database file.

Because of the above automated steps, in normal circumstances you should not need to run manual `ANALYZE` commands against your SQLite database file.

However, if you interrupted a previous `ANALYZE` command, and you notice that queries are running remarkably slowly, it could be the case that the index metadata stored in SQLite's internal tables has been removed or undesirably altered.

If this is the case, you can try manually running a full `ANALYZE` command, by doing the following:

1. Stop GoToSocial.
2. While connected to your GoToSocial database file in the `sqlite3` shell, run `PRAGMA analysis_limit=0; ANALYZE;` (this may take quite a few minutes).
3. Start GoToSocial.

[See here](https://sqlite.org/lang_analyze.html#approximate_analyze_for_large_databases) for more info.

### Vacuum

GoToSocial does not currently enable auto-vacuum for SQLite. To repack the database file to an optimal size you may want to run a `VACUUM` command on your SQLite database periodically (eg., every few months).

You can see lots of information about the `VACUUM` command [here](https://sqlite.org/lang_vacuum.html).

The basic steps are:

1. Stop GoToSocial.
2. While connected to your GoToSocial database file in the `sqlite3` shell, run `VACUUM;` (this may take quite a few minutes).
3. Start GoToSocial.

### Replication

It's a common practice to set up safeguards for your database like replication. SQLite can be replicated using external software. The basic steps are described on the [Replicating SQLite](../advanced/replicating-sqlite.md) page.

## Postgres

TODO: Maintenance recommendations for Postgres. 
