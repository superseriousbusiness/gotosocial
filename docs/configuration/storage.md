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
# Bool. Use SSL for S3 connections.
#
# Only set this to 'false' when testing locally.
#
# Default: true
storage-s3-use-ssl: true

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

### AWS S3 Bucket Configuration

#### Bucket Created
GoToSocial by default creates signed URL's which means we dont need to change anything major on the policies of the bucket.
Here are the steps to follow for bucket creation

1. Login to AWS -> select S3 as service.
2. click Create Bucket
3. Provide a unique name and avoid adding "." in the name
4. Do not change the public access settings (Let them be on "block public access" mode)

#### AWS ACCESS KEY Configuration

1. In AWS Console -> IAM (under Security, Identity, & Compliance)
2. Add a user with programatic api's access
3. We recommend setting up below listed policy, replace <bucketname> with your buckets name

```json
{
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "s3:ListAllMyBuckets",
            "Resource": "arn:aws:s3:::*"
        },
        {
            "Effect": "Allow",
            "Action": "s3:*",
            "Resource": [
                "arn:aws:s3:::<bucket_name>",
                "arn:aws:s3:::<bucket_name>/*"
            ]
        }
    ]
}
```

4. Provide the values in config above
  
  * storage-s3-endpoint -> should be your bucket location say `s3.ap-southeast-1.amazonaws.com`
  * storage-s3-access-key -> Access key you obtained for the user created above
  * storage-s3-secret-key -> Secret key you obtained for the user created above
  * storage-s3-bucket -> Keep this as the <bucketname> that you created just now.



#### Migrating data from local storage to AWS s3 bucket

This step is only needed if you have a running instance. Ignore this if you are setting up a fresh instance. 
We have provided [s3cmd](https://github.com/s3tools/s3cmd) command for the copy operation.

```bash
s3cmd sync --add-header="Cache-Control:public, max-age=315576000, immutable" ./ s3://<bucket name>
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
