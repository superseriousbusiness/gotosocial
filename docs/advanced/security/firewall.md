# Firewall

You should deploy a firewall on your instance to close off any open ports and give you a mechanism to ban potentially misbehaving clients. Many firewall frontends will also automatically install some rules that block obvious malicious packets.

It can be helpful to deploy tools that monitor your log files for certain patterns and automatically ban clients exhibiting certain behaviour. This can be use to monitor your SSH and web server access logs for things like SSH brute-force attacks.

## Ports

For GoToSocial, you'll want to ensure port `443` remains open. Without it, nobody will be able to reach your instance. Federation will fail and client apps won't be able to work at all.

If you [provision TLS certificates](../certificates.md) using ACME or GoToSocial's built-in Lets Encrypt support, you'll also need port `80` to be open.

In order to access your instance over SSH, you'll need to keep the port your SSH daemon is bound on open too. By default this is port `22`.

## ICMP

[Internet Control Message Protocol](https://en.wikipedia.org/wiki/Internet_Control_Message_Protocol) are exchanged between machines in order to detect certain network conditions or troubleshoot things. Many firewalls have a tendency of blocking ICMP entirely but this is undesirable. A few ICMP types should be allowed and you can use your firewall to configure rate limiting for them instead.

### IPv4

In order for things to work reliably, your firewall must allow:

* ICMP Type 3: "Destination Unreachable" and also aids in Path-MTU Discovery
* ICMP Type 4: "Source Quench"

If you want to be able to ping things or be pinged, you should also allow:

* ICMP Type 0: "Echo Reply"
* ICMP Type 8: "Echo Request"

For traceroute to work, it can be helpful to also allow:

* ICMP Type 11: "Time Exceeded"

### IPv6

ICMP is heavily relied on by all parts of the IPv6 stack and things will break in exciting and hard to debug ways if you block it. [RFC 4890](https://www.rfc-editor.org/rfc/rfc4890) was specifically written to address this and is worthwhile to review.

Roughly speaking, you must always allow:

* ICMP Type 1: "Destination Unreachable"
* ICMP Type 2: "Packet Too Big"
* ICMP Type 3, code 0: "Time Exceeded"
* ICMP Type 4, code 1, 2: "Parameter Problem"

For ping, you should allow:

* ICMP Type 128: "Echo Request"
* ICMP Type 129: "Echo Response"

## Firewall configuration

On Linux, firewalling is typically done using either [iptables](https://en.wikipedia.org/wiki/Iptables) or the more modern and faster [nftables](https://en.wikipedia.org/wiki/Nftables) as the backend. Most distributions are switching to nftables and many firewall frontends can be configured to use nftables instead. You'll need to refer to your distribution's documentation on the matter, but typically there will be an `iptables` or `nftables` service your init-system can start with a predefined location to load firewall rules from.

Doing this by hand using raw iptables or nftables rules offers the most control but can be challenging if you're not familiar with these systems. In order to help with that, a number of configuration frontends exist that you can use.

On the Debian and Ubuntu as well as openSUSE family of distributions, UFW is commonly used. It's a simple firewall front-end and many tutorials targeting those distributions will be using it.

For the Red Hat/CentOS family of distributions, firewalld is typically used. It's a much more advanced firewall configuration utility which also has a desktop GUI and [Cockpit integration](https://cockpit-project.org/).

Despite distribution preferences, you can use UFW, firewalld or something else entirely with any Linux distribution.

* [Ubuntu Wiki](https://wiki.ubuntu.com/UncomplicatedFirewall?action=show&redirect=UbuntuFirewall) on UFW
* [ArchWiki](https://wiki.archlinux.org/title/Uncomplicated_Firewall) on UFW
* Digital Ocean guide on [using UFW with Ubuntu 22.04](https://www.digitalocean.com/community/tutorials/how-to-set-up-a-firewall-with-ufw-on-ubuntu-22-04)
* [firewalld](https://firewalld.org/) project homepage and documentation
* [ArchWiki](https://wiki.archlinux.org/title/firewalld) on firewalld
* [Using and configuring firewalld](https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/9/html/configuring_firewalls_and_packet_filters/using-and-configuring-firewalld_firewall-packet-filters) from Red Hat
* Linode guide [on using firewalld](https://www.linode.com/docs/guides/introduction-to-firewalld-on-centos/)

## Brute-force protection

[fail2ban](https://www.fail2ban.org) and [SSHGuard](https://www.sshguard.net/) can be set up to monitor your log files for attempts to brute-force logins and other malicious behaviour. They can be configured to automatically insert firewall rules to block malicious IP addresses, either for a specific period of time or even indefinitely.

SSHGuard was initially designed just for SSH, but nowadays supports a variety of services. Fail2ban tends to support anything you can generate consistent log lines for, whereas SSHGuard's signature approach can catch more sophisticated or stealthy attacks as it computes an attack score over time.

Both SSHGuard and fail2ban ship with "backends" that can target iptables and nftables directly, or work with your frontend of choice like UFW or firewalld on Linux or pf on \*BSD. Make sure you review their documentation on how to correctly configure it.

* [ArchWiki](https://wiki.archlinux.org/title/Fail2ban) on fail2ban
* DigitalOcean guide on how to protect SSH with [fail2ban on Ubuntu](https://www.digitalocean.com/community/tutorial_collections/how-to-protect-ssh-with-fail2ban)
* Linode guide on how to [secure your server with fail2ban](https://www.linode.com/docs/guides/using-fail2ban-to-secure-your-server-a-tutorial/)
* [ArchWiki](https://wiki.archlinux.org/title/sshguard) on sshguard
* [FreeBSD manual](https://man.freebsd.org/cgi/man.cgi?query=sshguard&sektion=8&manpath=FreeBSD+13.2-RELEASE+and+Ports) for sshguard
* [SSHGuard setup](https://manpages.ubuntu.com/manpages/lunar/en/man7/sshguard-setup.7.html) manual for Ubuntu

For fail2ban, you can use the following regex, which triggers fail2ban on failed logins and not another 'Unauthorized' errors (API for example):

```regex
statusCode=401 path=/auth/sign_in clientIP=<HOST> .* msg=\"Unauthorized:
```

## IP blocking

GoToSocial implements rate-limiting in order to try and protect your instance from one party taking up all your processing capacity. However, if you know this traffic isn't legitimate or coming from an instance you don't wish to federate with anyway, you can block the IP(s) the traffic is originating from instead and spare GoToSocial from having to do any work.

### Linux

Blocking IPs is done with iptables or nftables. If you're using a firewall frontend like UFW or firewalld, use their facilities to block an IP.

In iptables, people tend to add a `DROP` rule for an IP in the `filter` table on the `INPUT` chain. On nftables, it's often done on a table with a chain with the `ip` or `ip6` address family.

In both those cases the kernel has already done a lot of unnecessary processing of the incoming traffic, just for it to then be blocked by an IP match. You can do this more efficiently by using the `mangle` table with the `PREROUTING` chain in iptables, or blocking using [the `netdev` family in nftables][nftnetdev].

You can read this blog post on [how to do this with iptables][iptblock]. For nftables, you can use something like:

```
table netdev filter {
    chain ingress {
        set baddiesv4 {
            type ipv4_addr
	        flags interval
            elements = { \
                1.0.0.0/8, \
                2.2.2.2/32 \
            }
        }
        set baddiesv6 {
            type ipv6_addr
	        flags interval
            elements = { \
                2620:4f:8000::/48, \
                fc00::/7 \
            }
        }

        type filter hook ingress device <interface name> priority -500;
        ip saddr @baddiesv4 drop
        ip6 saddr @baddiesv6 drop
    }
}
```

When using iptables, adding many rules slows things down significantly, including reloading the firewall when adding/removing rules. Since you may wish to block many IP addresses, use [the `ipset` module][ipset] and add a single block rule for the set instead: `-m set --match-set <set name>`.

[nftnetdev]: https://wiki.nftables.org/wiki-nftables/index.php/Nftables_families#netdev
[iptblock]: https://javapipe.com/blog/iptables-ddos-protection/
[ipset]: https://wiki.archlinux.org/title/Ipset

### BSDs

When using pf, you can create a persistent table, typically named `<badhosts>`, to which you add the IP addresses you want to block. Tables can also read from other files, so it's possible to keep the list of IPs outside of your main `pf.conf`.

An example of how to do this can be found [in the pf manual][manpf].

[manpf]: https://man.openbsd.org/pf.conf#TABLES
