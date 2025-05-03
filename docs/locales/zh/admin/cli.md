# GtS CLI 工具

GoToSocial 编译为二进制可执行文件。

使用此二进制文件的标准方法是通过运行 `gotosocial server start` 命令启动服务器。

不过，此二进制文件也可以作为管理工具和调试工具使用。

以下是 `gotosocial --help` 的完整输出，不包括全局配置选项的大列表。

```text
GoToSocial - 一个联合式社交媒体服务器

帮助文档参见：https://docs.gotosocial.org/zh-cn/。

代码仓库：https://codeberg.org/superseriousbusiness/gotosocial

用法：
  gotosocial [command]

可用命令：
  admin       gotosocial 管理相关任务
  debug       gotosocial 调试相关任务
  help        获取任何命令的帮助
  server      gotosocial 服务器相关任务
```

在 `可用命令` 下，可以看到标准的 `server` 命令。但是也有处理管理和调试的命令，这些将在本文档中进行解释。

!!! info "将全局配置传递给 CLI"
    
    对于所有这些命令，你仍然需要正确设置全局选项，以便 CLI 工具知道如何连接到你的数据库，以及使用哪个数据库、哪个主机和账户域等。
    
    你可以使用环境变量设置这些选项，通过 CLI 标志传递它们（例如，`gotosocial [commands] --host example.org`），或者只需将 CLI 工具指向你的配置文件（例如，`gotosocial --config-path ./config.yaml [commands]`）。

!!! info "附注"
    
    运行 CLI 命令时，你将会看到如下输出：
    
    ```text
    time=XXXX level=info msg=connected to SQLITE database
    time=XXXX level=info msg=there are no new migrations to run func=doMigration
    time=XXXX level=info msg=closing db connection
    ```
    
    这是正常的，表示命令已按预期运行。

!!! warning "运行管理命令后重启 GtS"
    
    由于 GoToSocial 的内部缓存机制，你可能需要在运行某些命令后重启 GoToSocial，以使命令的效果“生效”。我们仍在寻找一种无需重启的方法。在此期间，需要在运行命令后重启的命令将在下文中突出显示。

## gotosocial admin

包含 `account`、`export`、`import` 和 `media` 子命令。

### gotosocial admin account create

此命令可用于在你的实例上创建新账户。

!!! Warning "警告"
    执行此命令前，你必须至少运行过一次服务端，以初始化数据库中的必要条目，然后才能运行此命令。

`gotosocial admin account create --help`:

```text
创建一个新的本站账户

用法：
  gotosocial admin account create [flags]

标志：
      --email string      该账户的电子邮件地址
  -h, --help              获取创建命令的帮助
      --password string   为该账户设置的密码
      --username string   要创建/删除等的用户名
```

示例：

```bash
gotosocial admin account create \
   --username some_username \
   --email someuser@example.org \
   --password 'somelongandcomplicatedpassword' \
   --config-path config.yaml
```

### gotosocial admin account confirm

此命令可用于确认你的实例上的用户+账户，允许他们登录并使用账户。

!!! info "附注"
    
    如果账户是使用 `admin account create` 创建的，则不必在账户上运行 `confirm`，它将已被确认。

`gotosocial admin account confirm --help`:

```text
手动确认现有本站账户，从而跳过电子邮件确认

用法：
  gotosocial admin account confirm [flags]

标志：
  -h, --help              获取确认命令的帮助
      --username string   要创建/删除等的用户名
```

示例：

```bash
gotosocial admin account confirm --username some_username --config-path config.yaml
```

### gotosocial admin account promote

此命令可用于将用户提升为管理员。

!!! warning "需要重启服务器"
    
    为使更改生效，此命令需要在运行命令后重启 GoToSocial。

`gotosocial admin account promote --help`:

```text
将本站账户提升为管理员

用法：
  gotosocial admin account promote [flags]

标志：
  -h, --help              获取提升命令的帮助
      --username string   要创建/删除等的用户名
```

示例：

```bash
gotosocial admin account promote --username some_username --config-path config.yaml
```

### gotosocial admin account demote

此命令可用于将用户从管理员降级为普通用户。

!!! warning "需要重启服务器"
    
    为使更改生效，此命令需要在运行命令后重启 GoToSocial。

`gotosocial admin account demote --help`:

```text
将本站账户从管理员降级为普通用户

用法：
  gotosocial admin account demote [flags]

标志：
  -h, --help              获取降级命令的帮助
      --username string   要创建/删除等的用户名
```

示例：

```bash
gotosocial admin account demote --username some_username --config-path config.yaml
```

### gotosocial admin account disable

此命令可用于在你的实例上禁用一个账户：禁止其登录或执行任何操作，但不删除数据。

!!! warning "需要重启服务器"
    
    为使更改生效，此命令需要在运行命令后重启 GoToSocial。

`gotosocial admin account disable --help`:

```text
将本站账户的 `disabled` 设置为 true，以防止其登录或发布等，但不删除任何内容

用法：
  gotosocial admin account disable [flags]

标志：
  -h, --help              获取禁用命令的帮助
      --username string   要创建/删除等的用户名
```

示例：

```bash
gotosocial admin account disable --username some_username --config-path config.yaml
```

### gotosocial admin account enable

此命令可用于重新启用你实例上的账户，撤销之前的 `disable` 命令。

!!! warning "需要重启服务器"
    
    为使更改生效，此命令需要在运行命令后重启 GoToSocial。

`gotosocial admin account enable --help`:

```text
通过将本站账户的 `disabled` 设置为 false，撤销之前的禁用命令

用法：
  gotosocial admin account enable [flags]

标志：
  -h, --help              获取启用命令的帮助
      --username string   要创建/删除等的用户名
```

示例：

```bash
gotosocial admin account enable --username some_username --config-path config.yaml
```

### gotosocial admin account password

此命令可用于为指定的本站账户设置新密码。

!!! warning "需要重启服务器"
    
    为使更改生效，此命令需要在运行命令后重启 GoToSocial。

`gotosocial admin account password --help`:

```text
为指定的本站账户设置新密码

用法：
  gotosocial admin account password [flags]

标志：
  -h, --help              获取密码命令的帮助
      --password string   为该账户设置的密码
      --username string   要创建/删除等的用户名
```

示例：

```bash
gotosocial admin account password --username some_username --password some_really_good_password --config-path config.yaml
```

### gotosocial admin export

此命令可用于将你的 GoToSocial 实例中的数据导出到文件，以便备份/存储。

文件格式将是一系列以换行符分隔的 JSON 对象。

`gotosocial admin export --help`:

```text
将数据从数据库导出到指定路径的文件

用法：
  gotosocial admin export [flags]

标志：
  -h, --help          获取导出命令的帮助
      --path string   导入/导出文件的路径
```

Example:

```bash
gotosocial admin export --path example.json --config-path config.yaml
```

`example.json`:

```json
{"type":"account","id":"01F8MH5NBDF2MV7CTC4Q5128HF","createdAt":"2021-08-31T12:00:53.985645Z","username":"1happyturtle","locked":true,"language":"en","uri":"http://localhost:8080/users/1happyturtle","url":"http://localhost:8080/@1happyturtle","inboxURI":"http://localhost:8080/users/1happyturtle/inbox","outboxURI":"http://localhost:8080/users/1happyturtle/outbox","followingUri":"http://localhost:8080/users/1happyturtle/following","followersUri":"http://localhost:8080/users/1happyturtle/followers","featuredCollectionUri":"http://localhost:8080/users/1happyturtle/collections/featured","actorType":"Person","privateKey":"-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAzLP7oyyR+BU9ejn0CN9K+WpX3L37pxUcCgZAGH5lf3cGPZjz\nausfsFME94OjVyzw3K5M2beDkZ4E+Fak46NLtakLB1yovy9jKtj4Y4txHoMvRJLz\neUPxdfeXtpx2d3FDj++Uq4DEE0BhbePXhTGJWaNdC9MQmWKghJnCS5mrnFkdpEFx\njUz9l0UHl2Z4wppxPdpt7FyevcdfKqzGsAA3BxTM0dg47ZJWjtcvfCiSYpAKFNJY\nfKhKn9T3ezZgrLsF+o0IpD23KxWe1X4d5lgJRU9T4FmLmbvyJKUnfgYXbSEvLUcq\n79WbhgRcWwxWubjmWXgPGwzULVhhpYlwd2Cv3wIDAQABAoIBAGF+MxHjD15VV2NY\nKKb1GjMx98i1Xx6TijgoA+zmfha4LGu35e79Lql+0LXFp0zEpa6lAQsMQQhgd0OD\nmKKmSk+pxAvskJ4FxrhIf/yBFA4RMrj5OCaAOocRtdsOJ8n5UtFBrNAF0tzMY9q/\nkgzoq97aVF1mV9iFxaeBx6zT8ozSdqBq1PK/3w1dVg89S5tfKYc7Q0lQ00SfsTnd\niTDClKyqurebo9Pt6M7gXavgg3tvBlmwwr6XHs34Leng3oiN9mW8DVzaBMPzn+rE\nxF2eqs3v9vVpj8es88OwCh5P+ff8vJYvhu7Fcr/bJ8BItBQwfb8QBDATg/MXU2BI\n2ssW6AECgYEA4wmIyYGeu9+hzDa/J3Vh8GnlVNUCohHcChQdOsWsFXUgpVlUIHrX\neKHn42vD4Rzy52/YzJts4NkZTM9sL+kEXIEcpMG/S9xIIud7U0m/hMSAlmnJK/9j\niEXws3o4jo0E77jnRcBdIjpG4K5Eekm0DSR3SFhtZfEdN2DWPvu7K98CgYEA5tER\n/qJwFMc51AobMU87ZjXON7hI2U1WY/pVF62jSl0IcSsnj2riEKWLrs+GRG+HUg+U\naFSqAHcxaVHA0h0AYR8RopAhDdVKh0kvB8biLo+IEzNjPv2vyn0yRN5YSfXdGzyJ\nUjVU6kWdQOwmzy86nHgFaqEx7eofHIaGZzJK/AECgYEAu2VNQHX63TuzQuoVUa5z\nzoq5vhGsALYZF0CO98ndRkDNV22qIL0ESQ/qZS64GYFZhWouWoQXlGfdmCbFN65v\n6SKwz9UT3rvN1vGWO6Ltr9q6AG0EnYpJT1vbV2kUcaU4Y94NFue2d9/+TMnKv91B\n/m8Q/efvNGuWH/WQIaCKV6UCgYBz89WhYMMDfS4M2mLcu5vwddk53qciGxrqMMjs\nkzsz0Va7W12NS7lzeWaZlAE0gf6t98urOdUJVNeKvBoss4sMP0phqxwf0eWV3ur0\ncjIQB+TpGGikLVdRVuGY/UXHKe9AjoHBva8B3aTpB3lbnbNJBXZbIc1uYq3sa5w7\nXWWUAQKBgH3yW73RRpQNcc9hTUssomUsnQQgHxpfWx5tNxqod36Ytd9EKBh3NqUZ\nvPcH6gdh7mcnNaVNTtQOHLHsbPfBK/pqvb3MAsdlokJcQz8MQJ9SGBBPY6PaGw8z\nq/ambaQykER6dwlXTIlU20uXY0bttOL/iYjKmgo3vA66qfzS6nsg\n-----END RSA PRIVATE KEY-----\n","publicKey":"-----BEGIN RSA PUBLIC KEY-----\nMIIBCgKCAQEAzLP7oyyR+BU9ejn0CN9K+WpX3L37pxUcCgZAGH5lf3cGPZjzausf\nsFME94OjVyzw3K5M2beDkZ4E+Fak46NLtakLB1yovy9jKtj4Y4txHoMvRJLzeUPx\ndfeXtpx2d3FDj++Uq4DEE0BhbePXhTGJWaNdC9MQmWKghJnCS5mrnFkdpEFxjUz9\nl0UHl2Z4wppxPdpt7FyevcdfKqzGsAA3BxTM0dg47ZJWjtcvfCiSYpAKFNJYfKhK\nn9T3ezZgrLsF+o0IpD23KxWe1X4d5lgJRU9T4FmLmbvyJKUnfgYXbSEvLUcq79Wb\nhgRcWwxWubjmWXgPGwzULVhhpYlwd2Cv3wIDAQAB\n-----END RSA PUBLIC KEY-----\n","publicKeyUri":"http://localhost:8080/users/1happyturtle#main-key"}
{"type":"account","id":"01F8MH0BBE4FHXPH513MBVFHB0","createdAt":"2021-09-08T10:00:53.985634Z","username":"weed_lord420","locked":true,"language":"en","uri":"http://localhost:8080/users/weed_lord420","url":"http://localhost:8080/@weed_lord420","inboxURI":"http://localhost:8080/users/weed_lord420/inbox","outboxURI":"http://localhost:8080/users/weed_lord420/outbox","followingUri":"http://localhost:8080/users/weed_lord420/following","followersUri":"http://localhost:8080/users/weed_lord420/followers","featuredCollectionUri":"http://localhost:8080/users/weed_lord420/collections/featured","actorType":"Person","privateKey":"-----BEGIN RSA PRIVATE KEY-----\nMIIEogIBAAKCAQEAzsCcTHzwIgpWKVvut0Q/t1bFwnbj9hO6Ic6k0KXCXbf6qi0b\nMIyLRZr8DS61mD+SPSO2QKEL647xxyW2D8YGtwN6Cc6MpWETsWJkNtS8t7tDL//P\nceYpo5LiqKgn0TXj0Pq8Lvb7rqpH8QJ2EVm14SK+elhKZW/Bi5ZOEwfL8pw6EHI4\nus6VxCNQ099dksu++kbdD7zxqEKnk/4zOttYt0whlVrxzkibTjlKdlSlTYpIstU+\nfNyYVE0xWvrn+yF7jVlEwZYOFGfZbpELadrdOr2k1hvAk7upkrpKmLqYfwqD/xPc\nqwtx0iS6AEnmkSiTcAvju5vLkoLFRU7Of4AZ2wIDAQABAoIBAEAA4GHNS4k+Ke4j\nx4J0XkUjV5UbuPY0pSpSDjOJHOJmUfLcg85Ds9mYYO6zxwOaqmrC42ieclI5rh84\nTWQUqX9+VAk1J9UKeE4xZ1SSBtnZ3rK9PjrERZ+dmQ0dATaCuEO5Wwgu7Trk++Bg\nIqy8WNGZL94v9tfwALp1jTXW9AvmQoNdCFBP62vcmYW4YLjnggxLCFTA8YKfdePa\nTuxxY6uLkeBbxzWpbRU2+bmlxd5OnCkiRSMHIX+6JdtCu2JdWpUTCnWrFi2n1TZz\nZQx9z5rvowK1O785jGMFum5vBWpjIU8sJcXmPjGMU25zzmrhzfmkJsTXER3CXoUo\nSqSPqgECgYEA78OR7bY5KKQQ7Lyz6dru4Fct5P/OXTQoOg5aS7TKb95LVWj+TANn\n5djwIbLmAUV30z0Id9VgiZOL0Hny8+3VV9eU088Z408pAy5WQrL3dB8tZLUJSq5c\n5k6X15/VjWOOZKppDxShzoV3mcohrnwVwkv4fhPFQQOJJBYz6xurWs0CgYEA3MDE\nsDMd9ahzO0dl62ynojkkA8ZTcn2UdyvLpGj9UxT5j9vWF3CfqitXgcpNiVSIbxqQ\nbo/pBch7c/2Xakv5zkdcrJj5/6gyr+m1/tK2o7+CjDaSE4SYwufXx+qkl03Zpyzt\nKdOi7Hz/b2tdjump7ECEDE45mG2ea8oSnPgXl0cCgYBkGGFzu/9g2B24t47ksmHH\nhp3CXIjqoDurARLxSCi7SzJoFc0ULtfRPSAC8YzUOwwrQ++lF4+V3+MexcqHy2Kl\nqXqYcn18SC/3BAE/Fzf3Yoyw3mNiqihefbEmc7PTsxxfKkVx5ksmzNGBgsFM9sCe\nvNigyeAvpCo8xogmPwbqgQKBgE34mIBTzcUzFmBdu5YH7r3RyPK8XkUWLhZZlbgg\njTmHMw6o61mkIgENBf+F4RUckoQLsfAbTIcKZPB3JcAZzcYaVpVwAv1V/3E671lu\nO6xivE2iCL50GzDcis7GBhSbHsF5kNsxMV6uV9qW5ZjQ13/m2b0u9BDuxwHzgdeH\nmW2JAoGAIUOYniuEwdygxWVnYatpr3NPjT3BOKoV5i9zkeJRu1hFpwQM6vQ4Ds5p\nGC5vbMKAv9Cwuw62e2HvqTun3+U2Y5Uived3XCpgM/50BFrFHCfuqXEnu1bEzk5z\n9mIhp8uXPxzC5N7tRQfb3/eU1IUcb6T6ksbr2P81z0j03J55erg=\n-----END RSA PRIVATE KEY-----\n","publicKey":"-----BEGIN RSA PUBLIC KEY-----\nMIIBCgKCAQEAzsCcTHzwIgpWKVvut0Q/t1bFwnbj9hO6Ic6k0KXCXbf6qi0bMIyL\nRZr8DS61mD+SPSO2QKEL647xxyW2D8YGtwN6Cc6MpWETsWJkNtS8t7tDL//PceYp\no5LiqKgn0TXj0Pq8Lvb7rqpH8QJ2EVm14SK+elhKZW/Bi5ZOEwfL8pw6EHI4us6V\nxCNQ099dksu++kbdD7zxqEKnk/4zOttYt0whlVrxzkibTjlKdlSlTYpIstU+fNyY\nVE0xWvrn+yF7jVlEwZYOFGfZbpELadrdOr2k1hvAk7upkrpKmLqYfwqD/xPcqwtx\n0iS6AEnmkSiTcAvju5vLkoLFRU7Of4AZ2wIDAQAB\n-----END RSA PUBLIC KEY-----\n","publicKeyUri":"http://localhost:8080/users/weed_lord420#main-key"}
{"type":"account","id":"01F8MH17FWEB39HZJ76B6VXSKF","createdAt":"2021-09-05T10:00:53.985641Z","username":"admin","locked":true,"language":"en","uri":"http://localhost:8080/users/admin","url":"http://localhost:8080/@admin","inboxURI":"http://localhost:8080/users/admin/inbox","outboxURI":"http://localhost:8080/users/admin/outbox","followingUri":"http://localhost:8080/users/admin/following","followersUri":"http://localhost:8080/users/admin/followers","featuredCollectionUri":"http://localhost:8080/users/admin/collections/featured","actorType":"Person","privateKey":"-----BEGIN RSA PRIVATE KEY-----\nMIIEogIBAAKCAQEAxr2e1pqfLwwUCwHUdx56Mxnq5Kzc2EBwqN6jIPjiqVaG5eVq\nhujDhdqwMq0hnpBSPzLnvjiOtEh7Bwhx0MjuC/GRPTM9oNWPYD4PcjX5ofrubyLR\nBI97qD0SbyzUWzeyBi6R5tpW8LK1MJXNbnYlz5WouEiC4mY77ulri0EN2hCq80wg\nfvtEjEvELcKBqIytKH3rutIzfAyqXD7LSQ8UDoNh9GHyIfq8Zj32gWVk2MiPI3+G\n8kQJDmD8CKEasnrGVdSJBQUg3xDAtOibPXLP+07AIsKYMon35hVNvQNQPS7ru/Bk\nRhhGp2R44zqj6L9mxYbSrhFAaKDedu8oVe1aLQIDAQABAoIBAGK0aIADOU4ffJDe\n7sveiih5Fc1PATwx/QIR2QkWM1SREdx6LYclcX44V8xDanAbE44p1SkHY/CsEtYy\nXnyoXnn2FwFDQrdveY7+I6PApOPLAcKWkyLltC+hbVdj92/6YGNrm7EA/a77wruH\nmwjiivLnTG2CLecNiXSl33DA9YU4Yz+2Tza3IpTdjt8c/dz/BKKaxaWV+i9ew5VR\nioo5v51B+J8PrneCM/p8LGiLV148Njr0JqV6eFy1JuzItYMYdc3Fp+YnMzsuMZEA\n1akMcoln/ucVJyOFnCn6jx47nIoPZLl1KxX3aRDRfvrejm6W4yAkkTmR5voSRqax\njPL3rI0CgYEA9Acu4TO8xJ3uGaUad0N9JTYQVSmtAaE/g+df9LGMSzoj8X95S4xE\nQsGPqNGDm2VWADJjK4P05twZ+LfsfSKQ86wbp4/gbgnXpqB1P5Lty/B7KxiTnNwt\nwb1WGWTCukxfUSL3PRyf8uylkrg72RxKiBx4zKO3WVSLWOZWrFtn0qMCgYEA0H2p\nJs9Nv20ADOOX5tQ7+ruS6/B/Fhyj5fhflSYCAtOW7aME7+zQKJyqSQZ4b2Aub3Tp\nGIaUbRIGzjHyuTultFFWvjU3H5aI/0g1G9WKaBhNkyTIYVmMKtYyhXNvouWing8x\noraWx8TTBP8Cdnnk+QgdR2fpug8cghKupp5wvO8CgYA1JFtRL7MsHjh73TimQExA\njkWARlMmx7bNQtXis8eZmk+5h8kiaqly4DQoz3eZn7fa0x5Fm7b5j3UYdPVLSvvG\nFPTwyKRXUk1kPA1MivK+NuCbwf5jao+MYW8emJLPf1JCmRq+dD1g6aglC3n9Dewt\nOAYWipCjI4Y1FfRKFJ3HgQKBgEAb47+DTyzln3ZXJYZdDHR06SCTuwBZnixAy2NZ\nZJTp6yb3UbVU5E0Yn2QFEVNuB9lN4b8g4tMHEACnazN6G+HugPXL9z9HUqjs0yfT\n6dNIZdIxJUyJ9IfXhYFzlYhJhE+F7IVUD9kttJV8tI0pvja1QAuM8Fm9+84jYIDr\nh08RAoGAMYbjKHbtejcHBwt1kIcSss0cDmlZbBleJo8tdmdg4ndf5GE9N4/EL7tq\nm2zYSfr7OVdnOwRhoO+xF/6d1L7+TR1wz+k2fuMsI71aM5Ocp1nYTutjIkBTcldZ\nZzvjOgZWng5icuRLQQiDSKG5uqazqL/xGXkijb4kp4WW6myWY3c=\n-----END RSA PRIVATE KEY-----\n","publicKey":"-----BEGIN RSA PUBLIC KEY-----\nMIIBCgKCAQEAxr2e1pqfLwwUCwHUdx56Mxnq5Kzc2EBwqN6jIPjiqVaG5eVqhujD\nhdqwMq0hnpBSPzLnvjiOtEh7Bwhx0MjuC/GRPTM9oNWPYD4PcjX5ofrubyLRBI97\nqD0SbyzUWzeyBi6R5tpW8LK1MJXNbnYlz5WouEiC4mY77ulri0EN2hCq80wgfvtE\njEvELcKBqIytKH3rutIzfAyqXD7LSQ8UDoNh9GHyIfq8Zj32gWVk2MiPI3+G8kQJ\nDmD8CKEasnrGVdSJBQUg3xDAtOibPXLP+07AIsKYMon35hVNvQNQPS7ru/BkRhhG\np2R44zqj6L9mxYbSrhFAaKDedu8oVe1aLQIDAQAB\n-----END RSA PUBLIC KEY-----\n","publicKeyUri":"http://localhost:8080/users/admin#main-key"}
{"type":"account","id":"01F8MH1H7YV1Z7D2C8K2730QBF","createdAt":"2021-09-06T10:00:53.985643Z","username":"the_mighty_zork","locked":true,"language":"en","uri":"http://localhost:8080/users/the_mighty_zork","url":"http://localhost:8080/@the_mighty_zork","inboxURI":"http://localhost:8080/users/the_mighty_zork/inbox","outboxURI":"http://localhost:8080/users/the_mighty_zork/outbox","followingUri":"http://localhost:8080/users/the_mighty_zork/following","followersUri":"http://localhost:8080/users/the_mighty_zork/followers","featuredCollectionUri":"http://localhost:8080/users/the_mighty_zork/collections/featured","actorType":"Person","privateKey":"-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEApBmF8U+or+E0mgUMH3LE4uRIWzeV9rhYnvSMm9OpOsxwJiss\n5mEA/NtPHvQlq2UwrqXX89Wvu94K9EzZ4VyWYQGdxaiPpt17vRqUfsHUnXkY0pvC\nC9zt/aNlJtdt2xm+7PTC0YQd4+E1FX3aaoUPJL8MXzNlpJzaUtuwLZe1iBmFfatZ\nFHptEgc4nlf6TNLTzj3Yw1/7zIGVS8Vi7VquHc0Xo8dRiL2RxCGzLWnwL6GlrxY1\ntMhsUg467XeoiwegFCpcIhAhPFREKoTnCEksL/N0rpXl7m6CAy5uqBGs5mMXnXlq\nefr58l0j2dU6zc60LCHH9TJC+roXsKJhy9sx/QIDAQABAoIBAFa+UypbFG1cW2Tr\nNBxPm7ngOEtXl8MicV4dIVKh0TwOo13ZxtNFBbOj7jALmPn/9HrtmbkABPQHDL1U\n/nt9aNSAeTjpwH3RaD5vFX3n0g8n2zJBOZLxxzAjNi4RBLYj5uP1AiKkdvRlsJza\nuSFDkty2zMBqN9mLPHE+RePj5Qa6tjYfIQqQzu/+YnYMlXHoC2yHNKsvz6S5FhVj\nv5zATv2JlJQH3RSmhuPOah73iQnKCLzYYEAHleawKrCg/rZ3ht37Guvabeq7MqQN\nvi9pJdAA+RMxPsboHajskePjOTYJgKQSxEAMRTMfBR40aZxklxQL0EoBd1Y3CHXh\nfMg0xWECgYEA0ORrpJ1A2WNQwKcDDeBBsaJqWF4EraoFzYrugKZrAYEeVyuGD0zq\nARUaWkZTZ1f6wQ10i1WxAuKlBEds7QsLdZzLsA4um4JlBroCZiYfPnmTtb8op1LY\nFqeYTByvAmnfWWTuOI67GX9ruLg8tEGuz38kuQVSxYs51its3tScNPUCgYEAyRst\nwRbqpOqnwoRoS6pxv0Vpc3nUcfaVYwsg/qobJkiwAdlUYeE7alvEY926VW4cvU/X\nhy3L1punAqnyLI7uuqCefXEbNxO0Cebyy4Kv2Ye1uzl0OHsJczSNdfpNqfAIKwtN\nHLCYDGCsluQhz+I/5Pd0dT+JDPPW9hKS2HG7o+kCgYBqugn1VRLo/sEnbS02TbnC\n1ESZWY/yWsgUOEObH2vUnO+vgeFAt/9nBi0sqnm6d0z6jbFZ7zI9UycUhJm2ksoM\nEUxQay6M7ZZIVYkcP6X++YbqePyAYOdey8oYOR+BkC45MkQ0SVh2so+LFTaOsnBq\nO3+7uGiN3ZBzSESbpO0acQKBgQCONrsXZeZO82XpB4tdns3LbgGRWKEkajTgEnml\nvZNvck2NMSwb/5PttbFe0ei4CyMluPV4MamJPQ9Qse+BFR67OWR63uZY/4T8z6X4\nxpUmZnLcUFfgrRlUr+AtgvEy8HxGPDquxC7x6deC6RcEFEIM3/UqCOEZGMJ1x1Ky\n31LLKQKBgGCKwVgQ8+4JyHZFkie3YdHhxJDokgY+Opb0HNnoBY/lZ54UMCCJQPS2\n0XPSu651j/3adr3RQneU04gF6U2/D5JzFEV0kUsqZ4Zy2EEU0LU4ibus0gyomSpK\niWhU4QrC/M4ELxYZinlNu3ThPWNQ/PMNteVWfdgOcV7uUWl0ViFp\n-----END RSA PRIVATE KEY-----\n","publicKey":"-----BEGIN RSA PUBLIC KEY-----\nMIIBCgKCAQEApBmF8U+or+E0mgUMH3LE4uRIWzeV9rhYnvSMm9OpOsxwJiss5mEA\n/NtPHvQlq2UwrqXX89Wvu94K9EzZ4VyWYQGdxaiPpt17vRqUfsHUnXkY0pvCC9zt\n/aNlJtdt2xm+7PTC0YQd4+E1FX3aaoUPJL8MXzNlpJzaUtuwLZe1iBmFfatZFHpt\nEgc4nlf6TNLTzj3Yw1/7zIGVS8Vi7VquHc0Xo8dRiL2RxCGzLWnwL6GlrxY1tMhs\nUg467XeoiwegFCpcIhAhPFREKoTnCEksL/N0rpXl7m6CAy5uqBGs5mMXnXlqefr5\n8l0j2dU6zc60LCHH9TJC+roXsKJhy9sx/QIDAQAB\n-----END RSA PUBLIC KEY-----\n","publicKeyUri":"http://localhost:8080/users/the_mighty_zork#main-key"}
{"type":"block","id":"01FEXXET6XXMF7G2V3ASZP3YQW","createdAt":"2021-09-08T09:00:53.965362Z","uri":"http://localhost:8080/users/1happyturtle/blocks/01FEXXET6XXMF7G2V3ASZP3YQW","accountId":"01F8MH5NBDF2MV7CTC4Q5128HF","targetAccountId":"01F8MH5ZK5VRH73AKHQM6Y9VNX"}
{"type":"account","id":"01F8MH5ZK5VRH73AKHQM6Y9VNX","createdAt":"2021-08-31T12:00:53.985646Z","username":"foss_satan","domain":"fossbros-anonymous.io","locked":true,"language":"en","uri":"http://fossbros-anonymous.io/users/foss_satan","url":"http://fossbros-anonymous.io/@foss_satan","inboxURI":"http://fossbros-anonymous.io/users/foss_satan/inbox","outboxURI":"http://fossbros-anonymous.io/users/foss_satan/outbox","followingUri":"http://fossbros-anonymous.io/users/foss_satan/following","followersUri":"http://fossbros-anonymous.io/users/foss_satan/followers","featuredCollectionUri":"http://fossbros-anonymous.io/users/foss_satan/collections/featured","actorType":"Person","publicKey":"-----BEGIN RSA PUBLIC KEY-----\nMIIBCgKCAQEA2OyVgkaIL9VohXKYTh319j4OouHRX/8QC7piXj71k7q5RDzEyvis\nVZBc5/C1/crCpxt895i0Ai2CiXQx+dISV7s/JBhAGl8s7TQ8jLlMuptrI0+sdkBC\nlu8pU0qQmoeXVnlquOzNmqGufUxIDtLXLZDN17qf/7vWA23q4d0tG5KQhGGGKiVM\n61Ufvr9MmgPBSpyUvYMAulFlz1264L49aGWeVgOz3qUQzqtxjrP0kaIbeyt56miP\nKr5AqkRgSsXci+FAo6suxR5gzo9NgleNkbZWF9MQyKlawukPwZUDSh396vtNQMee\n/4mto7mAXw8iio0IacrYO3F7iyewXnmI/QIDAQAB\n-----END RSA PUBLIC KEY-----\n","publicKeyUri":"http://fossbros-anonymous.io/users/foss_satan/main-key"}
{"type":"follow","id":"01F8PYDCE8XE23GRE5DPZJDZDP","createdAt":"2021-09-08T09:00:54.749465Z","uri":"http://localhost:8080/users/the_mighty_zork/follow/01F8PYDCE8XE23GRE5DPZJDZDP","accountId":"01F8MH1H7YV1Z7D2C8K2730QBF","targetAccountId":"01F8MH5NBDF2MV7CTC4Q5128HF"}
{"type":"follow","id":"01F8PY8RHWRQZV038T4E8T9YK8","createdAt":"2021-09-06T12:00:54.749459Z","uri":"http://localhost:8080/users/the_mighty_zork/follow/01F8PY8RHWRQZV038T4E8T9YK8","accountId":"01F8MH1H7YV1Z7D2C8K2730QBF","targetAccountId":"01F8MH17FWEB39HZJ76B6VXSKF"}
{"type":"domainBlock","id":"01FF22EQM7X8E3RX1XGPN7S87D","createdAt":"2021-09-08T10:00:53.968971Z","domain":"replyguys.com","createdByAccountID":"01F8MH17FWEB39HZJ76B6VXSKF","privateComment":"i blocked this domain because they keep replying with pushy + unwarranted linux advice","publicComment":"reply-guying to tech posts","obfuscate":false}
{"type":"user","id":"01F8MGYG9E893WRHW0TAEXR8GJ","createdAt":"2021-09-08T10:00:53.97247Z","accountID":"01F8MH0BBE4FHXPH513MBVFHB0","encryptedPassword":"$2y$10$ggWz5QWwnx6kzb9g0tnIJurFtE0dhr5Zfeaqs9iFuUIXzafQlJVZS","locale":"en","lastEmailedAt":"0001-01-01T00:00:00Z","confirmationToken":"a5a280bd-34be-44a3-8330-a57eaf61b8dd","confirmationTokenSentAt":"2021-09-08T10:00:53.972472Z","unconfirmedEmail":"weed_lord420@example.org","moderator":false,"admin":false,"disabled":false,"approved":false}
{"type":"user","id":"01F8MGWYWKVKS3VS8DV1AMYPGE","createdAt":"2021-09-05T10:00:53.972475Z","email":"admin@example.org","accountID":"01F8MH17FWEB39HZJ76B6VXSKF","encryptedPassword":"$2y$10$ggWz5QWwnx6kzb9g0tnIJurFtE0dhr5Zfeaqs9iFuUIXzafQlJVZS","currentSignInAt":"2021-09-08T09:50:53.972477Z","lastSignInAt":"2021-09-08T08:00:53.972477Z","chosenLanguages":["en"],"locale":"en","lastEmailedAt":"2021-09-08T09:30:53.972478Z","confirmedAt":"2021-09-05T10:00:53.972478Z","moderator":true,"admin":true,"disabled":false,"approved":true}
{"type":"user","id":"01F8MGVGPHQ2D3P3X0454H54Z5","createdAt":"2021-09-06T22:00:53.97248Z","email":"zork@example.org","accountID":"01F8MH1H7YV1Z7D2C8K2730QBF","encryptedPassword":"$2y$10$ggWz5QWwnx6kzb9g0tnIJurFtE0dhr5Zfeaqs9iFuUIXzafQlJVZS","currentSignInAt":"2021-09-08T09:30:53.972481Z","lastSignInAt":"2021-09-08T08:00:53.972481Z","chosenLanguages":["en"],"locale":"en","lastEmailedAt":"2021-09-08T09:05:53.972482Z","confirmationTokenSentAt":"2021-09-06T22:00:53.972483Z","confirmedAt":"2021-09-07T00:00:53.972482Z","moderator":false,"admin":false,"disabled":false,"approved":true}
{"type":"user","id":"01F8MH1VYJAE00TVVGMM5JNJ8X","createdAt":"2021-09-06T22:00:53.972485Z","email":"tortle.dude@example.org","accountID":"01F8MH5NBDF2MV7CTC4Q5128HF","encryptedPassword":"$2y$10$ggWz5QWwnx6kzb9g0tnIJurFtE0dhr5Zfeaqs9iFuUIXzafQlJVZS","currentSignInAt":"2021-09-08T09:30:53.972485Z","lastSignInAt":"2021-09-08T08:00:53.972486Z","chosenLanguages":["en"],"locale":"en","lastEmailedAt":"2021-09-08T09:05:53.972487Z","confirmationTokenSentAt":"2021-09-06T22:00:53.972487Z","confirmedAt":"2021-09-07T00:00:53.972487Z","moderator":false,"admin":false,"disabled":false,"approved":true}
{"type":"instance","id":"01BZDDRPAB8J645ABY31HHF68Y","createdAt":"2021-09-08T10:00:54.763912Z","domain":"localhost:8080","title":"localhost:8080","uri":"http://localhost:8080","reputation":0}
```

### gotosocial admin import

此命令可用于将文件中的数据导入到你的 GoToSocial 数据库中。

如果数据库中尚未存在 GoToSocial 表，它们将被创建。

如果在导入过程中出现任何冲突（例如尝试导入特定帐户时已经存在），则进程将中止。

文件格式应为一系列以换行符分隔的 JSON 对象（参见上文）。

`gotosocial admin import --help`:

```text
从文件中导入数据到数据库

用法:
  gotosocial admin import [选项]

选项:
  -h, --help          导入命令的帮助信息
      --path string   要导入/导出文件的路径
```

示例:

```bash
gotosocial admin import --path example.json --config-path config.yaml
```

### gotosocial admin media list-attachments

可用于列出实例上的本站、外站或所有媒体附件的存储路径（包括头像和头图）。

`local-only` 和 `remote-only` 可用作过滤器；它们不能同时被设置。

如果既未设置 `local-only` 也未设置 `remote-only`，则将列出实例上的所有媒体附件。

你可能希望在运行此命令时将 `GTS_LOG_LEVEL` 设置为 `warn` 或 `error`，否则会记录大量你可能不需要的信息日志。

`gotosocial admin media list-attachments --help`:

```text
列出本站、外站或所有附件

用法:
  gotosocial admin media list-attachments [选项]

选项:
  -h, --help          list-attachments 命令的帮助信息
      --local-only    仅列出本站附件/表情；如果指定，则 remote-only 不能为 true
      --remote-only   仅列出外站附件/表情；如果指定，则 local-only 不能为 true
```

示例输出:

```text
/gotosocial/062G5WYKY35KKD12EMSM3F8PJ8/attachment/original/01PFPMWK2FF0D9WMHEJHR07C3R.jpg
/gotosocial/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg
/gotosocial/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/original/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.jpg
/gotosocial/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/original/01F8MH8RMYQ6MSNY3JM2XT1CQ5.jpg
/gotosocial/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/original/01F8MH7TDVANYKWVE8VVKFPJTJ.gif
/gotosocial/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg
/gotosocial/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/original/01F8MH58A357CV5K7R7TJMSH6S.jpg
/gotosocial/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/original/01CDR64G398ADCHXK08WWTHEZ5.gif
```

### gotosocial admin media list-emojis

用于列出您实例上的本站、外站或所有表情符号的存储路径。

`local-only` 和 `remote-only` 可用作过滤器；它们不能同时设置。

如果未设置 `local-only` 或 `remote-only`，将列出您实例上的所有表情符号。

您可能需要在运行时将 `GTS_LOG_LEVEL` 设置为 `warn` 或 `error`，否则将记录许多您可能不需要的信息消息。

`gotosocial admin media list-emojis --help`:

```text
列出本站、外站或所有表情符号

用法：
  gotosocial admin media list-emojis [标志]

标志：
  -h, --help          获取 list-emojis 帮助信息
      --local-only    仅列出本站附件/表情符号；如果指定，则 remote-only 不能为真
      --remote-only   仅列出外站附件/表情符号；如果指定，则 local-only 不能为真
```

示例输出：

```text
/gotosocial/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01GD5KP5CQEE1R3X43Y1EHS2CW.png
/gotosocial/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png
```

### gotosocial admin media prune orphaned

此命令可用于删除您 GoToSocial 中的孤立媒体。

孤立媒体定义为存储中使用 GoToSocial 格式的键存在，但没有相应数据库条目的媒体。这对于删除可能在以前安装中遗留的文件，或错误放置在存储中的文件非常有用。

!!! warning "需要停止服务器"
    
    此命令仅在 GoToSocial 未运行时起作用，因为它需要获取存储的独占锁。
    
    在运行此命令之前，请先停止 GoToSocial！

```text
删除存储中的孤立媒体

用法：
  gotosocial admin media prune orphaned [标志]

标志：
      --dry-run   执行试运行，仅记录可删除项目的数量（默认值为 true）
  -h, --help      获取 orphaned 帮助信息
```

默认情况下，此命令执行试运行，将记录可以删除的项目数量。要真正执行删除，请在命令中添加 `--dry-run=false`。

示例（试运行）：

```bash
gotosocial admin media prune orphaned
```

示例（实际执行）：

```bash
gotosocial admin media prune orphaned --dry-run=false
```

### gotosocial admin media prune remote

此命令可用于删除您 GoToSocial 中未使用/过时的外站媒体。

过时媒体是指外站实例中早于 `media-remote-cache-days` 的头像/头图/状态附件。

如果需要，这些项目将会在之后按需重新获取。

未使用媒体是指当前账号或状态未使用的头像/头图/状态附件。

!!! warning "需要停止服务器"
    
    此命令仅在 GoToSocial 未运行时起作用，因为它需要获取存储的独占锁。
    
    在运行此命令之前，请先停止 GoToSocial！

```text
从存储中删除未使用/过时的外站媒体，时间早于指定天数

用法：
  gotosocial admin media prune remote [标志]

标志：
      --dry-run   执行试运行，仅记录可删除项目的数量（默认值为 true）
  -h, --help      获取 remote 帮助信息
```

默认情况下，此命令执行试运行，将记录可以删除的项目数量。要真正执行删除，请在命令中添加 `--dry-run=false`。

示例（试运行）：

```bash
gotosocial admin media prune remote
```

示例（实际执行）：

```bash
gotosocial admin media prune remote --dry-run=false
```
