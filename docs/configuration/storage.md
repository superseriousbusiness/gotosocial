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
# this directly, and create new subdirectories and files with in.
# Only required when running with the local storage backend
# Examples: ["/home/gotosocial/storage", "/opt/gotosocial/datastorage"]
# Default: "/gotosocial/storage"
storage-local-base-path: "/gotosocial/storage"


# When using the s3 storage backend you need to set
# * storage-s3-endpoint: S3 Endpoint URL (e.g 'minio.example.org:9000')
# * storage-s3-access-key: S3 Access Key
# * storage-s3-secret-key: S3 Access Key
# * storage-s3-bucket: Name of the s3 bucket. Must exist
#
# You should also consider setting the access and secret key settings using
# environment variables to avoid leaking them via the config file
```
