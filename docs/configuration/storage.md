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
# Examples: ["minio:9000", "s3.nl-ams.scw.cloud", "s3.us-west-002.backblazeb2.com"]
# GoToSocial uses "DNS-style" when accessing buckets.
# If you are using Scaleways object storage, please remove the "bucket name" from the endpoint address
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

## AWS S3 Configuration

### Creating a bucket

GoToSocial by default creates signed URL's which means we don't need to change anything major on the policies of the bucket.

1. Login to AWS -> select S3 as service.
2. click Create Bucket
3. Provide a unique name and avoid adding "." in the name
4. Do not change the public access settings (Let them be on "block public access" mode)

### IAM Configuration

1. Create a [new user](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html) with programatic API access
2. Add an inline policy on this user, replacing <bucketname> with your bucket name

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
3. Create an [access key](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html) for this user
4. Provide the values in config above
  
  * `storage-s3-endpoint` -> S3 API endpoint for your region, for example: `s3.ap-southeast-1.amazonaws.com`
  * `storage-s3-access-key` -> Access key ID you obtained for the user created above
  * `storage-s3-secret-key` -> Secret key you obtained for the user created above
  * `storage-s3-bucket` -> The <bucketname> that you created just now

## Storage migration

Migration between backends is freely possible. To do so, you only have to move the directories (and their contents) between the different implementations.

When moving from one backend to another, the database will still contain references to headers and avatars from remote accounts pointing to the old storage backend which may result in them not loading correctly in clients. This will resolve itself over time, but you can force GoToSocial to refetch the avatar and header the next time you interact with a remote account by running the following SQL query on your database:

```sql
UPDATE accounts SET (avatar_media_attachment_id, avatar_remote_url, header_media_attachment_id, header_remote_url, fetched_at) = (null, null, null, null, null) WHERE domain IS NOT null;
```

### From local to AWS S3

There are multiple tools available that can help you copy the data from your filesystem to an AWS S3 bucket.

#### AWS CLI

With the official [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide)

```sh
aws s3 sync <storage-local-base-path> s3://<bucket name>
```

#### s3cmd

With [s3cmd](https://github.com/s3tools/s3cmd), you can use the following command:

```sh
s3cmd sync --add-header="Cache-Control:public, max-age=315576000, immutable" <storage-local-base-path> s3://<bucket name>
```

### From local to S3-compatible

This works for any S3-compatible store, including AWS S3 itself.

#### Minio CLI

You can use the [MinIO Client](https://docs.min.io/docs/minio-client-complete-guide.html). To perform the migration, you need to register your S3 compatible backend with the client and then ask it to copy the files:

```sh
mc alias set scw https://s3.nl-ams.scw.cloud
mc mirror <storage-local-base-path> scw/example-bucket/
```

If you want to migrate back, switch around the arguments of the `mc mirror` command.
