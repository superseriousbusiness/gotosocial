# Storage

## Settings

```yaml
##########################
##### STORAGE CONFIG #####
##########################

# Config pertaining to storage of user-created uploads (videos, images, etc).

# String. Type of storage backend to use.
# Examples: ["local", "s3"]
# Default: "local" (storage on local disk)
storage-backend: "local"

# String. Directory to use as a base path for storing files.
# Make sure whatever user/group gotosocial is running as has permission to access
# this directory, and create new subdirectories and files within it.
# Only required when running with the local storage backend.
# Examples: ["/home/gotosocial/storage", "/opt/gotosocial/datastorage"]
# Default: "/gotosocial/storage"
storage-local-base-path: "/gotosocial/storage"

# String. API endpoint of the S3 compatible service.
# Only required when running with the s3 storage backend.
#
# If your endpoint contains the bucket name, all files will be put into a
# subdirectory with the name of `storage-s3-bucket`
#
# Examples: ["minio:9000", "s3.nl-ams.scw.cloud", "s3.us-west-002.backblazeb2.com"]
# Default: ""
storage-s3-endpoint: ""

# Bool. If data stored in S3 should be proxied through GoToSocial instead of redirecting to a presigned URL.
#
# Default: false
storage-s3-proxy: false

# String. Access key part of the S3 credentials.
# Consider setting this value using environment variables to avoid leaking it via the config file
# Only required when running with the s3 storage backend.
# Examples: ["AKIAJSIE27KKMHXI3BJQ","miniouser"]
# Default: ""
storage-s3-access-key: ""
# String. Secret key part of the S3 credentials.
# Consider setting this value using environment variables to avoid leaking it via the config file
# Only required when running with the s3 storage backend.
# Examples: ["5bEYu26084qjSFyclM/f2pz4gviSfoOg+mFwBH39","miniopassword"]
# Default: ""
storage-s3-secret-key: ""
# String. Name of the storage bucket.
#
# If you have already encoded your bucket name in the storage-s3-endpoint, this
# value will be used as a directory containing your data.
#
# The bucket must exist prior to starting GoToSocial
#
# Only required when running with the s3 storage backend.
# Examples: ["gts","cool-instance"]
# Default: ""
storage-s3-bucket: ""
```

### Migrating between backends

Currently, migration between backends is freely possible. To do so, you only
have to move the directories (and their contents) between the different implementations.

One way to do so, is by utilizing the [MinIO
Client](https://docs.min.io/docs/minio-client-complete-guide.html). The
migration process might look something like this:

```bash
# 1. Change the GoToSocial configuration to the new backend (and restart)
# 2. Register the S3 Backend with the MinIO client
mc alias set scw https://s3.nl-ams.scw.cloud
# 3. Mirror the folder structure to the remote bucket
mc mirror /gotosocial/storage/ scw/example-bucket/
# 4. Aaaand we're done!
```

If you want to migrate back, switch around the arguments of the `mc mirror` command.
