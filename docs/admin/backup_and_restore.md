# Backup and Restore

As the GoToSocial database contains the instance as well as all user signing keys it is vital to back it up. If you lose these keys you'll never be able to federate from this domain again. Don't forget to also encrypt your backups in order to keep the data safe at rest.

Aside from disaster recovery, there are other good reasons to keep backups. Some potential scenarios for you to consider:

* You want to close down your instance but you might create it again later and you don't want to break federation.
* You need to migrate to a different database for some reason (Postgres => SQLite or vice versa).
* You're about to hack around on your instance and you want to make a quick backup so you don't lose everything if you mess up.

## What to backup

### Database

Most backup tools have built-in support for common databases like PostgreSQL and SQLite. Ensure you review their documentation first as they often spell out certain considerations and conditions that need to be met for backups to complete and restore successfully.

### Media

Local media should be backed up. You can use the [GoToSocial CLI](cli.md#gotosocial-admin-media-list-local) to list all media files that belong to your instance and its users.

Remote media does not have to be backed up. This can be a good way to keep the size of your backups down. Remote media will be fetched from the origin instance, much like how it'll be fetched again if it got pruned due to media retention.

## How to backup

You can go about this a few different ways:

* Imaging the VMs/machines your instance and database runs on
* Dumping GoToSocial's state with the CLI
* Backing up database and media files
* Backup software

Though setting up backup software can be a bit more work, it's by far the best option. It ensures consistent and encrypted backups and can protect you against filesystem corruption in a way that taking disk snapshots and copying the raw database and media files won't.

### Image your disk

If you're running GoToSocial on a VPS (a remote machine in the cloud), arguably the easiest way to preserve all of your database entries and media is to image the disk attached to the VPS. This will preserve the whole disk. Many VPS providers offer the option of automatically creating backups on a timer, so you'll always be able to restore if your data is lost.

Advantages:

* Relatively easy to do.
* Easy to automate (depending on your vps).
* Keep complete media + database entries.

Disadvantages:

* Can cost extra depending on your VPS.
* Will probably also preserve stuff you don't need, from other programs running on the same machine.
* Vendor lock-in, difficult to move the data around.

### Use the GoToSocial CLI

The GoToSocial CLI tool also provides commands for backing up and restoring data from your instance, which will preserve the *bare-minimum* necessary data to backup and restore your instance, without breaking federation with other instances.

What will be **kept**:

* All local account entries, including private and public keys.
* Followed/following remote accounts, including public keys.
* Follows/follow requests.
* Domain blocks.
* Account blocks.
* Account suspensions.
* User + password entries, email addresses.

What will be **dropped**:

* All statuses.
* Media.
* Faves.
* Bookmarks.
* Pins.
* Applications.
* Tokens.

The backup file produced will be in the form of a line-separated series of JSON objects (not a JSON array!). For example:

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

For information on how to use the commands to import/export, see [here](cli.md#gotosocial-admin-export). Though the `export` command won't backup media, you can use the [`media list-local`](cli.md#gotosocial-admin-media-list-local) command to figure out which media files you should keep.

Advantages:

* Database agnostic: exported data is in a somewhat generic format, and the `import` command can be used to insert this data into either a Postgres or an SQLite database.
* Lightweight: only what is needed is preserved, so backup files can be quite small (even small enough to send in an email). Backup/import commands just take a few seconds to run.
* Easily readable format: the output is just JSON.

Disadvantages:

* Loss of statuses/faves/etc: don't do a backup/restore this way unless you're willing to drop stuff.
* You need to use the GtS CLI tool to insert data back into a database, unless you write custom tooling for it.


### Back up your database files and media

Regardless of whether you're using PostgreSQL or SQLite as your GoToSocial database, it's possible to simply back up the database files directly by using something like [rclone](https://rclone.org/), or following best practices for [backing up Postgres data](https://www.postgresql.org/docs/15/backup.html) or [SQLite data](https://sqlite.org/backup.html).

Use the GoToSocial CLI's media [`list-attachments`](cli.md#gotosocial-admin-media-list-attachments) and [`list-emojis`](cli.md#gotosocial-admin-media-list-emojis) commands to get a list of media files you need to safeguard.

Advantages:

* Backups are relatively portable - you can move data from one machine to another.
* Well-documented procedure with a lot of guides and tooling available.
* Lots of different ways of doing your backups, depending on what you need.

Disadvantages:

* Can be a bit fiddly to set up initially.
* You need to figure out where to keep your backups.
* Restoring from backups can be a pain.
* Unless you back up media as well, references to media attachments in your db will be broken.

### Backup software

Backup software is created with the specific purpose of helping you create, manage and restore your backups. It typically knows how to safely backup your database so you don't have to be an expert on how to do PostgreSQL or SQLite backups. It can backup from the filesystem too.

Though the same advantages and disadvantages roughly apply as with backing up the database files directly, this approach does have some nice extras:

* Backups are highly portable and can be used to restore the database from 0
* Backups happen on a regular schedule and with configurable retention policies
* Backups are incremental and compressed to save on storage and bandwidth
* Backups are encrypted
* Built-in tooling to list your snapshots and restore from them

!!! tip
    [Rsync.net](https://rsync.net/), [BorgBase](https://www.borgbase.com/) and [Hetzner Storage](https://www.hetzner.com/storage/storage-box) provide affordable storage that you can use as a backup target. Rsync.net has a special Borg-only backup product that is much cheaper than their regular storage product. If you only want to use them for backups managed with Borg, [sign up here instead](https://www.rsync.net/products/borg.html).

#### Borgmatic

[Borgmatic](https://torsion.org/borgmatic/) is a utility to help perform backups using [Borg](https://www.borgbackup.org/). It's driven by a declarative configuration file using YAML. BorgBase, Rsync.net and Hetzner all support Borg.

!!! warning
    When initialising the Borg repository, ensure you set it up with a strong encryption key and store that key somewhere safely. Without it you won't be able to decrypt your backups in the future. The ArchWiki entry on Borgmatic explains how to safely pass your encryption key to Borgmatic without storing it plain text in its configuration file.

How to backup databases with Borgmatic has its own [documentation page](https://torsion.org/borgmatic/docs/how-to/backup-your-databases/) that you should review. A simple `config.yaml` for Borgmatic with GoToSocial using SQLite looks like this:

```yaml
location:
  repositories:
    - path: ssh://<find it in your provider control panel>
      label: <anything but typically the provider, for example borgbase>
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

For PostgreSQL, you'll want to use `postgresql_databases` instead.

The file mentioned in `patterns_from` can be created by transforming the output from the GoToSocial CLI media [`list-attachments`](cli.md#gotosocial-admin-media-list-attachments) and [`list-emojis`](cli.md#gotosocial-admin-media-list-emojis) commands. In order to generate the right patterns you can use the [`media-to-borg-patterns.py`](https://github.com/superseriousbusiness/gotosocial/tree/main/example/borgmatic/media-to-borg-patterns.py) script. How Borg patterns work is explained in [their documentation](https://man.archlinux.org/man/borg-patterns.1).

You'll need to put that file on your GoToSocial instance and make sure the file is executable. It requires Python 3 which you will already have if you have Borg and Borgmatic installed. It only depends on the Python standard library.

!!! note
    For this to work reliably, you should ensure that the [storage-local-base-path](../configuration/storage.md) in your GoToSocial configuration uses an absolute path. Otherwise you'll have to tweak the paths yourself.

```sh
$ gotosocial --config-path /path/to/config.yaml admin media list-attachments --local-only | \
    /path/to/media-to-borg-patterns.py \
    <storage-local-base-path>
```

This will output a pattern set looking roughly like this to your console:

```
R <storage-local-base-path>
+ pp:<storage-local-base-path>/<account ID>
- <storage-local-base-path>/*
```

!!! tip
    You can view the help by passing `--help` to `media-to-borg-patterns.py`. It can write the output to a file directly by passing the location of a file as the last argument to the script.

Given this set of patterns, Borg will start looking for files starting from `<storage-local-base-path>`. Anything that matches the path prefix, `pp:` will be included. Everything else will match the last pattern, excluding it from the archive.

On a single-user instance, you can run this command once and inline the patterns directly in your Borgmatic configuration [using the `patterns` key](https://torsion.org/borgmatic/docs/reference/configuration/). On multi-user instances you should run this after a user signs up. Alternatively, you can run it every time before you do a backup.

If you're running Borgmatic as a systemd service, you can [create a drop-in](https://wiki.archlinux.org/title/systemd#Drop-in_files) for `borgmatic.service` and run the pattern generation before the backup is started with:

```ini
[Service]
ExecStartPre=/path/to/gotosocial --config-path /path/to/config.yaml admin media list-attachments --local-only | /path/to/media-to-borg-patterns.py <storage-local-base-path> /etc/borgmatic/gotosocial_patterns
```

Documentation that's good to review:

* Borgmatic [configuration reference](https://torsion.org/borgmatic/docs/reference/configuration/)
* ArchWiki entry [on Borgmatic](https://wiki.archlinux.org/title/Borgmatic)
* ArchWiki entry [on Borg](https://wiki.archlinux.org/title/Borg_backup)
* BorgBase [documentation](https://docs.borgbase.com/)
* Hetzner community guide on [setting up Borgmatic](https://community.hetzner.com/tutorials/install-and-configure-borgmatic)
