# System Requirements

GoToSocial needs a domain name, and a *server* to run on, either a homeserver in your house, or a cloud server.

## Server / VPS

The system requirements for GoToSocial are fairly minimal: for a single-user instance with about 100 followers/followees, it uses somewhere between 50 to 100MB of RAM. CPU usage is only intensive when handling media (encoding blurhashes, mostly) and/or doing a lot of federation requests at the same time.

These light requirements mean GtS runs pretty well on something like a Raspberry Pi (a €40 single-board computer). It's been tested on a Raspberry Pi Zero W as well (a €9 computer smaller than a credit card), but it's not quite able to run on that. It should run on a Raspberry Pi Zero W 2 (which costs €14!), but we haven't tested that yet.

If you have an old laptop or a dusty desktop lying around that you're not using anymore, it will probably be a perfect candidate for running GoToSocial.

If you decide to use a VPS instead, you can just spin yourself up something cheap with Linux running on it.

[Hostwinds](https://www.hostwinds.com/) is a good option here: it's cheap and they throw in a static IP address for free.

[Greenhost](https://greenhost.net) is also great: it has zero co2 emissions, but is a bit more costly.

## Ports

The installation guides won't go into running [UFW](https://www.digitalocean.com/community/tutorials/how-to-set-up-a-firewall-with-ufw-on-ubuntu-18-04) and [Fail2Ban](https://linuxize.com/post/install-configure-fail2ban-on-ubuntu-20-04/) but you absolutely should do that.

For ports, you should leave `443` and `80` open. `443` is used for https requests to GoToSocial, and `80` is used for LetsEncrypt certification verification.

If you can't leave `443` and `80` open on the machine, don't worry! You can configure these ports in GoToSocial, but you'll have to also configure port forwarding to properly forward traffic on `443` and `80` to whatever ports you choose.

## Domain Name

To run a GoToSocial server, you also need a domain name, and it needs to be pointed towards your VPS or homeserver.

[Namecheap](https://www.namecheap.com/) is a good place to do this, but you can use any domain name registrar that lets you manage your own DNS.

IMPORTANT: If you want to host GoToSocial at a different host from your desired account domain (eg., you want to host GtS at `fedi.example.org` but you want your account to show up at `example.org`), please read the [advanced configuration](./advanced.md) carefully, before proceeding with installation!
