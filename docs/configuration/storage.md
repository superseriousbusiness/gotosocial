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
# Examples: ["/home/gotosocial/storage", "/opt/gotosocial/datastorage"]
# Default: "/gotosocial/storage"
storage-local-base-path: "/gotosocial/storage"
```
