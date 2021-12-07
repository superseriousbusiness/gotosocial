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
storage-base-path: "/gotosocial/storage"

# String. Protocol to use for serving stored files.
# It's very unlikely that you'll need to change this ever, but there might be edge cases.
# Examples: ["http", "https"]
storage-serve-protocol: "https"

# String. Host for serving stored files.
# If you're using local storage, this should be THE SAME as the value you've set for Host, above.
# It should only be a different value if you're serving stored files from a host
# other than the one your instance is running on.
# Examples: ["localhost", "example.org"]
# Default: "localhost" -- you should absolutely change this.
storage-serve-host: "localhost"

# String. Base path for serving stored files. This will be added to serveHost and serveProtocol
# to form the prefix url of your stored files. Eg., https://example.org/fileserver/.....
# It's unlikely that you will need to change this.
# Examples: ["/fileserver", "/media"]
# Default: "/fileserver"
storage-serve-base-path: "/fileserver"
```
