# Backup and Restore

In certain conditions, it may be desirable to be able to back up a GoToSocial instance, and then to restore it again later, or just save the backup somewhere.

Some potential scenarios:

* You want to close down your instance but you might create it again later and you don't want to break federation.
* You need to migrate to a different database for some reason (Postgres => SQLite or vice versa).
* You want to keep regular backups of your data just in case something happens.
* You want to migrate from GoToSocial to a different Fediverse server, or from a different Fediverse server to GoToSocial.
* You're about to hack around on your instance and you want to make a quick backup so you don't lose everything if you mess up.

There are a few different ways of doing this, most of which require some technical knowledge.

## Image your disk

If you're running GoToSocial on a VPS (a remote machine in the cloud), the easiest way to preserve all of your database entries and media is to ..........................................

