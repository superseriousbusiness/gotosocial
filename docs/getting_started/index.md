# Deployment considerations

Before deploying GoToSocial, it's important to think through a few things as some choices will have long-term consequences for how you run and manage GoToSocial.

!!! danger

    It's not supported across the Fediverse to switch between implementations on the same domain. This means that if you run GoToSocial on example.org, you'll run into federation issues if you try to switch to a different implementation like Pleroma/Akkoma, Misskey/Calckey etc.
    
    In that same vein, if you already have another ActivityPub implementation running on example.org you should not attempt to switch to GoToSocial on that domain.

## Database

GoToSocial supports both SQLite and Postgres and you can start using either. We do not currently have tooling to support migrating from SQLite to Postgres or vice-versa, but it is possible in theory.

SQLite is great for a single-user instance. If you're planning on hosting multiple people it's advisable to use Postgres instead. You can always use Postgres regardless of the instance size.

!!! tip
    Please [backup  your database](../admin/backup_and_restore.md). The database contains encryption keys for the instance and any user accounts. You won't be able to federate again from the same domain if you lose these keys.

## Domain name

In order to federate with others, you'll need a domain like `example.org`. You can register your domain name through any domain registrar, like [Namecheap](https://www.namecheap.com/). Make sure you pick a registrar that also lets you manage DNS entries, so you can point your domain to the IP of the server that's running your GoToSocial instance.

You'll commonly see usernames existing at the apex of the domain, for example `@me@example.org` but this is not required. It's perfectly fine to have users exist on `@me@social.example.org` instead. Many people prefer to have usernames on the apex as its shorter to type, but you can use any (subdomain) you control.

It is possible to have usernames like `@me@example.org` but have GoToSocial running on `social.example.org` instead. This is done by distinguishing between the API domain, called the "host", and the domain used for usernames, called the "account domain".

!!! danger
    It's not possible to safely change whether the host and account domain are different after the fact. It requires regenerating the database and will cause confusion for any server you have already federated with.

When using a single domain, you only need to configure the "host" in the GoToSocial configuration:

```yaml
host: "example.org"
```

When using a split domain approach, you need to configure both the "host" and the "account-domain":

```yaml
host: "social.example.org"
account-domain: "example.org"
```

## TLS

For federation to work, you have to use TLS. Most implementations, including GoToSocial, will generally refuse to federate over unencrypted transports.

GoToSocial comes with built-in support for provisioning certificates through Lets Encrypt. It can also load certificates from disk. If you have a reverse-proxy in front of GoToSocial you can handle TLS at that level instead.

!!! tip
    Make sure you configure the use of modern versions of TLS, TLSv1.2 and higher, in order to keep communications between servers and clients safe. When GoToSocial handles TLS termination this is done automatically for you. If you have a reverse-proxy in use, use the [Mozilla SSL Configuration Generator](https://ssl-config.mozilla.org/).

## Server / VPS

!!! bug "Clustering / multi-node deployments"
    GoToSocial does not support [clustering or any form of multi-node deployment](https://github.com/superseriousbusiness/gotosocial/issues/1749). Though multiple GtS instances can use the same Postgres database and either shared local storage or the same object bucket, GtS relies on a lot of internal caching to keep things fast. There is no mechanism for synchronising these caches between instances. Without it, you'll get all kinds of odd and inconsistent behaviour.

GoToSocial aims to fit in small spaces so we try and ensure that the system requirements are fairly minimal: for a single-user instance with about 100 followers/followees, it uses somewhere between 50 to 100MB of RAM. CPU usage is only intensive when handling media (encoding blurhashes, mostly) and/or doing a lot of federation requests at the same time.

These light requirements mean GtS runs pretty well on something like a Raspberry Pi (a €40 single-board computer). It's been tested on a Raspberry Pi Zero W as well (a €9 computer smaller than a credit card), but it's not quite able to run on that. It should run on a Raspberry Pi Zero W 2 (which costs €14!), but we haven't tested that yet. You can also repurpose an old laptop or desktop to run GoToSocial for you.

If you decide to use a VPS instead, you can spin yourself up something cheap with Linux running on it. Most of the VPS offerings in the €2-€5 range will perform admirably for a personal GoToSocial instance.

[Hostwinds](https://www.hostwinds.com/) is a good option here: it's cheap and they throw in a static IP address for free.

[Greenhost](https://greenhost.net) is also great: it has zero CO2 emissions, but is a bit more costly.

## Ports

GoToSocial needs ports `80` and `443` open.

* `80` is used for Lets Encrypt. As such, you don't need it if you don't use the built-in Lets Encrypt provisioning.
* `443` is used to serve the API on with TLS and is what any instance you're federating with will try to connect to.

If you can't leave `443` and `80` open on the machine, don't worry! You can configure these ports in GoToSocial, but you'll have to also configure port forwarding to properly forward traffic on `443` and `80` to whatever ports you choose.

!!! tip
    You should configure a firewall on your machine, as well as some protection against brute-force SSH login attempts and the like. Take a look at our [firewall documentation](../advanced/security/firewall.md) for pointers on what to configure and tools that can help you out.
