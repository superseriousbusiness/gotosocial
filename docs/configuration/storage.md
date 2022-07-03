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
# NOTE: s3 storage is not yet supported!
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
# Examples: ["minio:9000", "s3.nl-ams.scw.cloud", "s3.us-west-002.backblazeb2.com"]
# Default: ""
storage-s3-endpoint: ""

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
