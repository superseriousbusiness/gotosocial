# 存储

## 设置

```yaml
##########################
##### 存储配置指南 #####
##########################

# 用户创建上传内容（如视频、图片等）的存储配置。

# 字符串。要使用的存储后端类型。
# 示例: ["local", "s3"]
# 默认: "local"（存储在本地磁盘上）
storage-backend: "local"

# 字符串。用于存储文件的根目录。
# 确保运行 GoToSocial 的用户/组有权限访问此目录，并能在其中创建新的子目录和文件。
# 仅在使用本地存储后端时需要。
# 示例: ["/home/gotosocial/storage", "/opt/gotosocial/datastorage"]
# 默认: "/gotosocial/storage"
storage-local-base-path: "/gotosocial/storage"

# 字符串。S3 兼容服务的 API 端点。
# 仅在使用 s3 存储后端时需要。
# 示例: ["minio:9000", "s3.nl-ams.scw.cloud", "s3.us-west-002.backblazeb2.com"]
# GoToSocial 使用“DNS 风格”访问桶。
# 如果你使用 Scaleways 对象存储，请移除端点地址中的“桶名称”
# 默认: ""
storage-s3-endpoint: ""

# 布尔值。如果 S3 中存储的数据应通过 GoToSocial 代理而不是转发到预签名 URL，请将其设置为 true。
#
# 在大多数情况下，你无需更改此设置，但如果你的桶提供商无法生成预签名 URL，
# 或者你的桶无法暴露给更广泛的互联网，这可能有用。
#
# 默认: false
storage-s3-proxy: false

# 字符串。用于重定向传入媒体请求的基本 URL。
#
# 必须以“http://”或“https://”开头，并以无尾斜杠结尾。
#
# 除非有正当理由，否则不要设置此值！对“常规”s3 使用没有必要，大多数管理员可以忽略此设置。
#
# 如果设置，那么对实例的媒体文件服务器请求将被重定向到此 URL 而不是你的桶 URL，保留相关路径部分。
#
# 如果你在 S3 桶前使用 CDN 代理，并希望通过 CDN 提供媒体，而不是直接从 S3 桶提供媒体，这会很有用。
#
# 例如，如果你的 storage-s3-endpoint 值设置为“s3.my-storage.example.org”，
# 并且你设置了一个 CDN 以代理你的桶，从“cdn.some-fancy-host.org”提供服务，
# 那么你应该将 storage-s3-redirect-url 设置为“https://cdn.some-fancy-host.org”。
#
# 这将允许你的 GoToSocial 实例*上传*数据到“s3.my-storage.example.org”，
# 但引导调用者从“https://cdn.some-fancy-host.org” *下载* 这些数据。
#
# 如果 storage-backend 不是 s3，或者 storage-s3-proxy 为 true，则忽略此值。
#
# 示例: ["https://cdn.some-fancy-host.org"]
# 默认: ""
storage-s3-redirect-url: ""

# 布尔值。使用 SSL 进行 S3 连接。
#
# 仅在本地测试时将此设置为 'false'。
#
# 默认: true
storage-s3-use-ssl: true

# 字符串。S3 凭证的访问密钥部分。
# 请考虑使用环境变量设置此值，以避免通过配置文件泄露
# 仅在使用 s3 存储后端时需要。
# 示例: ["AKIAJSIE27KKMHXI3BJQ","miniouser"]
# 默认: ""
storage-s3-access-key: ""

# 字符串。S3 凭证的秘密密钥部分。
# 请考虑使用环境变量设置此值，以避免通过配置文件泄露
# 仅在使用 s3 存储后端时需要。
# 示例: ["5bEYu26084qjSFyclM/f2pz4gviSfoOg+mFwBH39","miniopassword"]
# 默认: ""
storage-s3-secret-key: ""

# 字符串。存储桶的名称。
#
# 如果你已经在 storage-s3-endpoint 中包含了你的桶名称，
# 此值将用作包含你数据的目录。
#
# 存储桶必须在启动 GoToSocial 之前就已存在
#
# 仅在使用 s3 存储后端时需要。
# 示例: ["gts","cool-instance"]
# 默认: ""
storage-s3-bucket: ""
```

## AWS S3 配置

### 创建一个桶

GoToSocial 默认创建签名 URL，这意味着我们不需要在桶的策略上做重大更改。

1. 登录 AWS -> 选择 S3 作为服务。
2. 点击创建桶
3. 提供一个唯一名称，避免在名称中添加`.`
4. 不要更改公开访问设置（保持“屏蔽公开访问”模式）

### IAM 配置

1. 创建一个具有程序化 API 访问权限的[新用户](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html)
2. 在此用户上添加在线政策，将 `<bucketname>` 替换为你的桶名称
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
3. 为此用户创建一个[访问密钥](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html)
4. 在上方配置中提供值
    * `storage-s3-endpoint` -> 你所在区域的 S3 API 端点，例如: `s3.ap-southeast-1.amazonaws.com`
    * `storage-s3-access-key` -> 你为上面创建的用户获取的访问密钥 ID
    * `storage-s3-secret-key` -> 你为上面创建的用户获取的秘密密钥
    * `storage-s3-bucket` -> 你刚刚创建的 `<bucketname>`

### `storage-s3-redirect-url`

如果你在 S3 桶前使用 CDN，并希望通过 CDN 提供媒体，而不是直接从 S3 桶提供媒体，你应将 `storage-s3-redirect-url` 设置为 CDN URL。

例如，如果你的 `storage-s3-endpoint` 值设置为 `s3.my-storage.example.org`，并且你设置了一个 CDN 来代理你的桶，从 `cdn.some-fancy-host.org` 提供服务，那么你应该将 `storage-s3-redirect-url` 设置为 `https://cdn.some-fancy-host.org`。

这将允许你的 GoToSocial 实例*上传*数据到 `s3.my-storage.example.org`，但引导调用者从 `https://cdn.some-fancy-host.org` *下载* 那些数据。

## 存储迁移

可以自由地在后端之间迁移。要做到这一点，你只需在不同的实现之间移动目录（及其内容）。

从一个后端迁移到另一个后端时，数据库中的外站账户的头像和资料卡横幅背景引用仍然指向旧存储后端，这可能导致它们在客户端中无法正确加载。这将在一段时间后自行解决，但你可以在下一次与外站账户交互时强制 GoToSocial 重新获取头像和封面。当 GoToSocial 不运行时，你可以在数据库上执行以下指令（执行后重启实例也可）。这将确保缓存被清除。

```sql
UPDATE accounts SET (avatar_media_attachment_id, avatar_remote_url, header_media_attachment_id, header_remote_url, fetched_at) = (null, null, null, null, null) WHERE domain IS NOT null;
```

### 从本地到 AWS S3

有多种工具可帮助你将数据从文件系统复制到 AWS S3 桶。

#### AWS CLI

使用官方 [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide)

```sh
aws s3 sync <storage-local-base-path> s3://<bucket name>
```

#### s3cmd

使用 [s3cmd](https://github.com/s3tools/s3cmd)，可以使用以下命令：

```sh
s3cmd sync --add-header="Cache-Control:public, max-age=315576000, immutable" <storage-local-base-path> s3://<bucket name>
```

### 从本地到 S3 兼容存储

这适用于任何 S3 兼容的存储，包括 AWS S3 本身。

#### Minio CLI

你可以使用 [MinIO 客户端](https://docs.min.io/docs/minio-client-complete-guide.html)。要执行迁移，你需要使用客户端注册你的 S3 兼容后端，然后让它复制文件：

```sh
mc alias set scw https://s3.nl-ams.scw.cloud
mc mirror <storage-local-base-path> scw/example-bucket/
```

如果你想迁移回来，请交换 `mc mirror` 命令的参数顺序。
