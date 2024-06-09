# Replicating SQLite

Next to your regular [backup methods](../admin/backup_and_restore.md), you might want to set up replication for disaster recovery to another path or external host.

For this to work properly, SQLite needs the journal mode to be configured in `WAL` mode, with synchronous mode set to `NORMAL`. This is the default configuration for GoToSocial.

You can check your settings in the configuration file. The journal mode is set in `db-sqlite-journal-mode` and the synchronous mode in `db-sqlite-synchronous`.

## Litestream on Linux or MacOS

A relatively light, and fast way to set up replication with SQLite is by using [Litestream](https://litestream.io). It can be configured very easily and supports different backends like file based replication, S3 compatible storage and many other setups.

You can then install the prebuilt package by either using Homebrew on MacOS or the deb file on Linux.

Using Homebrew on MacOS:

```bash
brew install benbjohnson/litestream/litestream
```

Using a .deb package on Linux:

Navigate to the [releases page](https://github.com/benbjohnson/litestream/releases/latest), and download the latest release (make sure to select the appropiate platform for the wget command below).


```bash
wget https://github.com/benbjohnson/litestream/releases/download/v0.3.13/litestream-v0.3.13-linux-amd64.deb
sudo dpkg -i litestream-*.deb
```

## Configuring Litestream

Configuration is done by editing the configuration file. It's located in /etc/litestream.yml.

### Configuring file based replication

```yaml
dbs:
    - path: /gotosocial/sqlite.db
      - path: /backup/sqlite.db
```

### Configuring S3 based replication

Set up a bucket for replication, and make sure to set it to be private.
Make sure to replace the example `access-key-id` and `secret-access-key` with the proper values from your dashboard.

```yaml
access-key-id: AKIAJSIE27KKMHXI3BJQ
secret-access-key: 5bEYu26084qjSFyclM/f2pz4gviSfoOg+mFwBH39

dbs:
    - path: /gotosocial/sqlite.db
      - url: s3://my.bucket.com/db

```

When using a S3 compatible storage provider you will need to set an endpoint.
For example for minio this can be done with the following configuration.

```yaml
access-key-id: miniouser
secret-access-key: miniopassword

dbs:
    - path: /gotosocial/sqlite.db
      - type: s3
	    bucket: mybucket
		path: sqlite.db
		endpoint: minio:9000
```

## Enabling replication

You can enable replication on Linux by enabling the Litestream service.

```bash
sudo systemctl enable litestream
sudo systemctl start litestream
```

Check if it's running properly using `sudo journalctl -u litestream -f`.

If you need to change the configuration file, restart Litestream:

```bash
sudo systemctl restart litestream
```

### Recovering from the configured backend

You can pull down a recovery file from the stored backend with the following simple command.

```bash
sudo litestream restore
```

If you have configured multiple files to be backupped, or have multiple replicas, specify what you want to do.

For filebased replication:

```bash
sudo litestream restore -o /gotosocial/sqlite.db /backup/sqlite.db
```

For s3 based replication:

```bash
sudo litestream restore -o /gotosocial/sqlite.db s3://bucketname/db
```