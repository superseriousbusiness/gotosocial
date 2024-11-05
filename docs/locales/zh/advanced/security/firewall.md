# 防火墙

你应该在你的实例上部署防火墙，以关闭任何开放端口，并提供一个机制来封禁可能行为不端的客户端。许多防火墙前端还会自动安装一些规则来屏蔽明显的恶意数据包。

部署工具来监控日志文件中的某些趋势，并自动封禁表现出某种行为的客户端是很有帮助的。这可以用于监控你的 SSH 和 Web 服务器访问日志，以应对如 SSH 暴力破解攻击。

## 端口

对于 GoToSocial，你需要确保端口 `443` 保持开放。没有它，任何人都无法访问你的实例。联合将失败，客户端应用程序将无法正常工作。

如果你使用 ACME 或 GoToSocial 的内置 Lets Encrypt 支持[配置 TLS 证书](../certificates.md)，你还需要开放端口 `80`。

为了通过 SSH 访问你的实例，你还需要保持 SSH 守护进程绑定的端口开放。默认情况下，SSH 端口是 `22`。

## ICMP

[ICMP](https://en.wikipedia.org/wiki/Internet_Control_Message_Protocol) 是在机器之间交换数据，以检测某些网络条件或排除故障的协议。许多防火墙倾向于完全屏蔽 ICMP，但这并不理想。应该允许一些 ICMP 类型，你可以使用防火墙为它们配置速率限制。

### IPv4

为了确保功能可靠，你的防火墙必须允许：

* ICMP 类型 3："目标不可达"，并有助于路径 MTU 发现
* ICMP 类型 4："源抑制"

如果你希望能够 ping 或被 ping，还应允许：

* ICMP 类型 0："回显应答"
* ICMP 类型 8："回显请求"

为了 traceroute 能够工作，还可以允许：

* ICMP 类型 11："时间超限"

### IPv6

IPv6 协议栈的所有部分非常依赖 ICMP，屏蔽它会导致难以调试的问题。[RFC 4890](https://www.rfc-editor.org/rfc/rfc4890) 专门为此而写，值得查看。

简单来说，你必须始终允许：

* ICMP 类型 1："目标不可达"
* ICMP 类型 2："数据包过大"
* ICMP 类型 3，代码 0："时间超限"
* ICMP 类型 4，代码 1, 2："参数问题"

对于 ping，你应该允许：

* ICMP 类型 128："回显请求"
* ICMP 类型 129："回显应答"

## 防火墙配置

在 Linux 上，通常使用 [iptables](https://en.wikipedia.org/wiki/Iptables) 或更现代、更快的 [nftables](https://en.wikipedia.org/wiki/Nftables) 作为后端进行防火墙配置。大多数发行版正在转向使用 nftables，许多防火墙前端可以配置为使用 nftables。你需要参考发行版的文档，但通常会有一个 `iptables` 或 `nftables` 服务，可以通过预定义的位置加载防火墙规则。

手动使用原始的 iptables 或 nftables 规则提供了最大的控制精度，但如果不熟悉这些系统，这样做可能会有挑战。为了帮助解决这个问题，存在许多配置前端可以使用。

在 Debian 和 Ubuntu 以及 openSUSE 系列的发行版中，通常使用 UFW。它是一个简单的防火墙前端，许多针对这些发行版的教程都会使用它。

对于 Red Hat/CentOS 系列的发行版，通常使用 firewalld。它是一个更高级的防火墙配置工具，也有桌面 GUI 和 [Cockpit 集成](https://cockpit-project.org/)。

尽管发行版有各自偏好，你可以在任何 Linux 发行版中使用 UFW、firewalld 或其他完全不同的工具。

* [Ubuntu Wiki](https://wiki.ubuntu.com/UncomplicatedFirewall?action=show&redirect=UbuntuFirewall) 关于 UFW 的介绍
* [ArchWiki](https://wiki.archlinux.org/title/Uncomplicated_Firewall) 关于 UFW 的介绍
* DigitalOcean 指南 [在 Ubuntu 22.04 上使用 UFW 建立防火墙](https://www.digitalocean.com/community/tutorials/how-to-set-up-a-firewall-with-ufw-on-ubuntu-22-04)
* [firewalld](https://firewalld.org/) 项目主页及文档
* [ArchWiki](https://wiki.archlinux.org/title/firewalld) 关于 firewalld 的介绍
* [使用和配置 firewalld](https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/9/html/configuring_firewalls_and_packet_filters/using-and-configuring-firewalld_firewall-packet-filters) 的 Red Hat 文档
* Linode 指南 [如何使用 firewalld](https://www.linode.com/docs/guides/introduction-to-firewalld-on-centos/)

## 暴力攻击防护

[fail2ban](https://www.fail2ban.org) 和 [SSHGuard](https://www.sshguard.net/) 可以配置以监控日志文件中暴力破解登录和其他恶意行为的尝试。它们可以配置为自动插入防火墙规则，以屏蔽恶意 IP 地址，屏蔽可以是暂时的，也可以是永久的。

SSHGuard 最初只为 SSH 设计，但现在支持多种服务。Fail2ban 往往支持任何可生成一致日志行的服务，而 SSHGuard 的签名方法可以捕获更复杂或隐蔽的攻击，因为它随着时间的推移计算攻击分数。

SSHGuard 和 fail2ban 都带有后端，可以直接针对 iptables 和 nftables，或与你选择的前端如 UFW 或 firewalld 在 Linux 上工作，在 \*BSD 系统上可以使用 pf。确保查看其文档以正确配置。

* [ArchWiki](https://wiki.archlinux.org/title/Fail2ban) 关于 fail2ban 的介绍
* DigitalOcean 指南如何在 Ubuntu 上使用 [fail2ban 保护 SSH](https://www.digitalocean.com/community/tutorial_collections/how-to-protect-ssh-with-fail2ban)
* Linode 指南如何使用 [fail2ban 保护服务器](https://www.linode.com/docs/guides/using-fail2ban-to-secure-your-server-a-tutorial/)
* [ArchWiki](https://wiki.archlinux.org/title/sshguard) 关于 sshguard 的介绍
* [FreeBSD 手册](https://man.freebsd.org/cgi/man.cgi?query=sshguard&sektion=8&manpath=FreeBSD+13.2-RELEASE+and+Ports) sshguard 的介绍
* [SSHGuard 设置](https://manpages.ubuntu.com/manpages/lunar/en/man7/sshguard-setup.7.html) 的 Ubuntu 手册

对于 fail2ban，可以使用以下正则表达式，该正则表达式在身份验证失败时触发 fail2ban，而不是其他“未经授权”的错误（例如 API）：

```regex
statusCode=401 path=/auth/sign_in clientIP=<HOST> .* msg=\"Unauthorized:
```

## IP 屏蔽

GoToSocial 实现了速率限制，以保护你的实例不被单个主体占用所有处理能力。然而，如果你知道这不是合法流量，或者来自你不想与之联邦的实例，你可以屏蔽流量来源的 IP，以节省 GoToSocial 的处理能力。

### Linux

屏蔽 IP 是通过 iptables 或 nftables 实现的。如果你使用 UFW 或 firewalld 等防火墙前端，请使用其功能来屏蔽 IP。

在 iptables 中，人们倾向于在 `filter` 表的 `INPUT` 链中为 IP 添加一个 `DROP` 规则。在 nftables 中，通常在一个具有 `ip` 或 `ip6` 地址族的链的表中完成。在这些情况下，内核已经对传入流量进行了大量不必要的处理，然后再通过 IP 匹配进行屏蔽。

使用 iptables 时，可以更有效地使用 `mangle` 表和 `PREROUTING` 链。你可以查看这篇博客文章，[了解它在 iptables 中的工作原理][iptblock]。对于 nftables，你可能会想要使用 [`netdev` family][nftnetdev] 进行屏蔽。

[iptblock]: https://javapipe.com/blog/iptables-ddos-protection/
[nftnetdev]: https://wiki.nftables.org/wiki-nftables/index.php/Nftables_families#netdev

#### iptables

使用 `iptables` 屏蔽 IP 的示例：

```
iptables -t mangle -A PREROUTING -s 1.0.0.0/8 -j DROP
ip6tables -t mangle -A PREROUTING -s fc00::/7 -j DROP
```

当使用 iptables 时，添加许多规则会显著降低速度，包括在添加/删除规则时重新加载防火墙。由于你可能希望屏蔽许多 IP 地址，请使用 [ipset 模块][ipset] 并为集合添加单个屏蔽规则。

[ipset]: https://ipset.netfilter.org/ipset.man.html

首先创建你的集合并添加一些 IP：

```
ipset create baddiesv4 hash:ip family inet
ipset create baddiesv6 hash:ip family inet6

ipset add baddiesv4 1.0.0.0/8
ipset add baddiesv6 fc00::/7
```

然后，更新你的 iptables 规则以针对该集合：

```
iptables -t mangle -A PREROUTING -m set --match-set baddiesv4 src -j DROP
ip6tables -t mangle -A PREROUTING -m set --match-set baddiesv6 src -j DROP
```

#### nftables

对于 nftables，你可以使用如下配置：

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

### BSDs

使用 pf 时，你可以创建一个通常命名为 `<badhosts>` 的持久化表，将需要屏蔽的 IP 地址添加到该表中。表格还可以从其他文件读取，因此可以将 IP 列表保存在主 `pf.conf` 之外。

有关如何执行此操作的示例，可以在 [pf 手册][manpf] 中找到。

[manpf]: https://man.openbsd.org/pf.conf#TABLES
