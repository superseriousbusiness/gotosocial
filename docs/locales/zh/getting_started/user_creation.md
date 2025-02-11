# 创建用户

无论使用哪种安装方法，你都需要创建一些用户。GoToSocial 目前还没有通过网页 UI 创建用户或让人们通过网页 UI 注册的功能。

在此期间，你可以使用 CLI 创建用户：

```sh
./gotosocial --config-path /path/to/config.yaml \
    admin account create \
    --username some_username \
    --email some_email@whatever.org \
    --password 'SOME_PASSWORD'
```

在上述命令中，将 `some_username` 替换为你想要的用户名，将 `some_email@whatever.org` 替换为你想关联到用户的电子邮件地址，将 `SOME_PASSWORD` 替换为一个安全的密码。

如果你想让用户拥有管理员权限，可以使用类似的命令提升他们：

```sh
./gotosocial --config-path /path/to/config.yaml \
    admin account promote --username some_username
```

将 `some_username` 替换为你刚创建的账户的用户名。

!!! warning "提权需要重启服务器"
    
    由于 GoToSocial 的缓存机制，某些管理员 CLI 命令在执行后需要重启服务器才能使更改生效。
    
    例如，将用户提升为管理员后，你需要重启 GoToSocial 服务器，以便从数据库加载新值。

!!! tip "提示"
    
    要查看其他可用的 CLI 命令，请点击[这里](../admin/cli.md)。

## 容器

当从容器运行 GoToSocial 时，你需要在容器中执行上述命令。如何操作取决于你的容器运行时，但对于 Docker 来说，应该像这样：

```sh
docker exec -it CONTAINER_NAME_OR_ID \
    /gotosocial/gotosocial \
    admin account create \
    --username some_username \
    --email someone@example.org \
    --password 'some_very_good_password'
```

如果你遵循我们的 Docker 指南，容器名应该为 `gotosocial`。你可以通过 `docker ps` 获取名称或 ID。
