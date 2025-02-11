# 备份和恢复

由于 GoToSocial 数据库包含了实例以及所有用户的签名密钥，备份是至关重要的。如果这些密钥丢失，您将无法再从此域名进行联合与交流。请记住加密备份，以确保数据在静置时的安全。

除了灾难恢复之外，还有其他一些保持备份的好理由。您可以考虑以下几种可能的场景：

* 您想关闭实例，但可能会在以后重新创建，并且不希望破坏联合功能。
* 出于某种原因，您需要迁移到不同的数据库（从 Postgres 到 SQLite 或反之亦然）。
* 您准备对实例进行一些调整，希望快速备份以防出错导致数据丢失。

## 需要备份的内容

### 数据库

大多数备份工具都内置对常见的数据库的支持，如 PostgreSQL 和 SQLite。请确保首先查看他们的文档，因为它们通常会详细说明备份成功完成和恢复所需满足的某些注意事项和条件。

### 媒体文件

本站媒体文件应被备份。您可以使用 [GoToSocial CLI](cli.md#gotosocial-admin-media-list-local) 列出属于您的实例及其用户的所有媒体文件。

外站媒体无需备份。这可以有效地控制备份大小。外站媒体会从原始实例获取，就像因为媒体保留而被修剪掉后再次获取一样。

## 如何备份

您可以通过以下几种方式进行备份：

* 为实例和数据库运行所在的 VM/机器制作镜像
* 使用 CLI 转储 GoToSocial 的状态
* 备份数据库和媒体文件
* 使用备份软件

尽管设置备份软件可能需要更多工作，但这无疑是最佳选择。它确保了一致和加密的备份，并可以抵御文件系统损坏，而磁盘快照和复制原始数据库及媒体文件无法做到这一点。

### 镜像您的磁盘

如果您在 VPS（云端的远程机器）上运行 GoToSocial，制作附加到 VPS 的磁盘镜像可能是保存所有数据库条目和媒体的最简单方法。这将保留整个磁盘。许多 VPS 提供商提供定时自动创建备份的选项，因此即使数据丢失，您也可以随时恢复。

优点：

* 相对容易操作。
* 易于自动化（视具体的 VPS 而定）。
* 保留完整的媒体和数据库条目。

缺点：

* 可能会根据您的 VPS 产生额外费用。
* 可能也会保留您不需要的其他程序运行的数据。
* 与供应商绑定，数据难以迁移。

### 使用 GoToSocial CLI

GoToSocial CLI 工具还提供了从实例备份和恢复数据的命令，这将保留所有必要的最少数据以备份和恢复您的实例，而不破坏与其他实例的联邦。

将**保留**的内容有：

* 所有本站账户条目，包括私钥和公钥。
* 关注/被关注的外站账户，包括公钥。
* 关注/关注请求。
* 实例屏蔽列表。
* 账户屏蔽列表。
* 账户封禁列表。
* 用户名和密码条目，电子邮件地址。

将**丢弃**的：

* 所有贴文。
* 媒体。
* 收藏。
* 书签。
* 置顶。
* 应用程序。
* 令牌。

生成的备份文件将是一个以行分隔的 JSON 对象序列（而不是 JSON 数组！）。例如：

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

有关如何使用命令导入/导出的信息，参见[此处](cli.md#gotosocial-admin-export)。尽管 `export` 命令不会备份媒体，但可以使用 [`media list-local`](cli.md#gotosocial-admin-media-list-local) 命令来确定应该保留哪些媒体文件。

优势：

* 数据库无关：导出的数据采用了一种通用格式，可以使用 `import` 命令将这些数据插入 Postgres 或 SQLite 数据库中。
* 轻量级：仅保留必要的内容，因此备份文件可以非常小（甚至小到可以通过电子邮件发送）。备份/导入命令只需几秒钟即可运行。
* 易读格式：输出是 JSON 格式。

劣势：

* 贴文/收藏等的丢失：除非你愿意丢失某些内容，否则不要这样进行备份/恢复。
* 需要使用 GoToSocial CLI 工具将数据插回数据库，除非你为此编写自定义工具。

### 备份你的数据库文件和媒体

无论你是使用 PostgreSQL 还是 SQLite 作为 GoToSocial 数据库，都可以直接使用类似 [rclone](https://rclone.org/) 的工具来备份数据库文件，或者遵循 [Postgres 数据备份的最佳实践](https://www.postgresql.org/docs/15/backup.html) 或 [SQLite 数据备份的最佳实践](https://sqlite.org/backup.html)。

使用 GoToSocial CLI 的媒体 [`list-attachments`](cli.md#gotosocial-admin-media-list-attachments) 和 [`list-emojis`](cli.md#gotosocial-admin-media-list-emojis) 命令来获取需要保护的媒体文件列表。

优势：

* 备份相对便携 - 可以将数据从一台机器迁移到另一台。
* 有大量文档和工具可供参考。
* 可以根据需要有多种备份方式。

劣势：

* 初次设置可能有点复杂。
* 需要确定备份的存放位置。
* 从备份中恢复可能比较麻烦。
* 除非同时备份媒体，否则数据库中的媒体附件引用将会中断。

### 备份软件

备份软件专为帮助你创建、管理和恢复备份而设计。它通常知道如何安全地备份数据库，因此你不用成为 PostgreSQL 或 SQLite 备份专家。它也可以从文件系统中进行备份。

尽管与直接备份数据库文件的方法大致具有相同的优点和缺点，这种方法还有一些额外的好处：

* 备份高度便携，可以从零开始恢复数据库。
* 备份按照常规计划进行，并有可配置的保留策略。
* 备份是增量和压缩的，以节省存储和带宽。
* 备份是加密的。
* 内置工具可以列出快照并从中恢复。

!!! tip "提示"
    [Rsync.net](https://rsync.net/)、[BorgBase](https://www.borgbase.com/) 和 [Hetzner Storage](https://www.hetzner.com/storage/storage-box) 提供了可用于备份的经济实惠的存储。Rsync.net 有一种专门为 Borg 设计的备份产品，比他们的常规存储产品便宜得多。如果你只想使用 Borg 管理的备份，请在[此处注册](https://www.rsync.net/products/borg.html)。

#### Borgmatic

[Borgmatic](https://torsion.org/borgmatic/) 是一个帮助使用 [Borg](https://www.borgbackup.org/) 进行备份的工具。它通过使用 YAML 的声明性配置文件驱动。BorgBase、Rsync.net 和 Hetzner 都支持 Borg。

!!! warning "警告"
    初始化 Borg 仓库时，确保使用强加密密钥进行设置，并将密钥安全地存放在某处。否则将无法在将来解密备份。ArchWiki 上关于 Borgmatic 的条目解释了如何安全地将你的加密密钥传递给 Borgmatic，而不在配置文件中以明文形式存储它。

如何使用 Borgmatic 备份数据库有其[单独的文档页面](https://torsion.org/borgmatic/docs/how-to/backup-your-databases/)，你应当在备份前查看一下。对于使用 SQLite 的 GoToSocial，Borgmatic 的简单 `config.yaml` 如下：


```yaml
location:
  repositories:
    - path: ssh://<在你的提供商控制面板中查找ssh地址>
      label: <可以是任意内容，但通常是提供商名称，例如 borgbase>
  patterns_from:
  - /etc/borgmatic/gotosocial_patterns

storage:
  compression: auto,zstd
  archive_name_format: '{hostname}-{now:%Y-%m-%d-%H%M%S}'
  retries: 5
  retry_wait: 30

retention:
    keep_daily: 7
    keep_weekly: 6
    keep_monthly: 12

hooks:
  before_backup:
      - /usr/bin/systemctl stop gotosocial
  after_backup:
      - /usr/bin/systemctl start gotosocial
  sqlite_databases:
    - name: gotosocial
      path: /path/to/sqlite.db
```

对于 PostgreSQL，你应该使用 `postgresql_databases`。

`patterns_from` 中提到的文件可以通过转换 GoToSocial CLI 媒体命令 [`list-attachments`](cli.md#gotosocial-admin-media-list-attachments) 和 [`list-emojis`](cli.md#gotosocial-admin-media-list-emojis) 的输出来创建。要生成正确的模式，您可以使用 [`media-to-borg-patterns.py`](https://github.com/superseriousbusiness/gotosocial/tree/main/example/borgmatic/media-to-borg-patterns.py) 脚本。有关 Borg 模式如何工作的详情，参见 [他们的文档](https://man.archlinux.org/man/borg-patterns.1)。

您需要将该文件放在您的 GoToSocial 实例上，并确保该文件是可执行的。它需要 Python 3，安装 Borg 和 Borgmatic 后您应该已经具备。它仅依赖于 Python 标准库。

!!! note "注意"
    为了确保可靠运行，您应确保 GoToSocial 配置中的 [storage-local-base-path](../configuration/storage.md) 使用的是绝对路径。否则您将需要自己调整路径。

```sh
$ gotosocial --config-path /path/to/config.yaml admin media list-attachments --local-only | \
    /path/to/media-to-borg-patterns.py \
    <storage-local-base-path>
```

这将在控制台上输出一个类似于以下内容的模式集：

```
R <storage-local-base-path>
+ pp:<storage-local-base-path>/<account ID>
- <storage-local-base-path>/*
```

!!! tip "提示"
    你可以通过向 `media-to-borg-patterns.py` 传递 `--help` 来查看帮助。通过将文件位置作为脚本的最后一个参数，也可以将输出直接写入文件。

给定这组模式，Borg 将从 `<storage-local-base-path>` 开始寻找文件。任何匹配路径前缀 `pp:` 的都会被包括进去。其他的则会匹配最后一个模式，从存档中排除。

在单用户实例中，你可以运行此命令一次，并直接在 Borgmatic 配置中内联模式 [使用 `patterns` 键](https://torsion.org/borgmatic/docs/reference/configuration/)。在多用户实例中，你应该在用户注册后运行此命令。或者，每次备份前都可以运行它。

如果你将 Borgmatic 作为 systemd 服务运行，可以为 `borgmatic.service` [创建一个 drop-in](https://wiki.archlinux.org/title/systemd#Drop-in_files)，在备份开始前运行模式生成：


```ini
[Service]
ExecStartPre=/path/to/gotosocial --config-path /path/to/config.yaml admin media list-attachments --local-only | /path/to/media-to-borg-patterns.py <storage-local-base-path> /etc/borgmatic/gotosocial_patterns
```

建议查看的文档：

* Borgmatic [配置参考](https://torsion.org/borgmatic/docs/reference/configuration/)
* ArchWiki 关于 [Borgmatic 的条目](https://wiki.archlinux.org/title/Borgmatic)
* ArchWiki 关于 [Borg 的条目](https://wiki.archlinux.org/title/Borg_backup)
* BorgBase [文档](https://docs.borgbase.com/)
* Hetzner 社区指南关于 [设置 Borgmatic](https://community.hetzner.com/tutorials/install-and-configure-borgmatic)
