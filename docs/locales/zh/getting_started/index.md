# 部署注意事项

在部署 GoToSocial 之前，有几个关键点需要你仔细考虑，因为这些选择将影响你如何运行和管理 GoToSocial。

!!! danger "危险"

    在同一域名上切换不同实现是不被 Fediverse 支持的。这意味着如果你在 example.org 上运行 GoToSocial，而尝试切换到其他实现如 Pleroma/Akkoma、Misskey/Calckey 等，你会遇到联合问题。
    
    同理，如果你已经在 example.org 上运行了其他 ActivityPub 实现，就不应该尝试在那个域名上切换到 GoToSocial。

## 服务器 / VPS 系统要求

GoToSocial 致力于为在小型设备上运行的使用场景优化，因此我们尽量确保系统需求处于合理低位，同时仍提供人们期望的社交媒体服务器功能。

下面是系统需求的详细信息，但简而言之，你应该选择至少有 **1 个 CPU 核心**、大约 **1GB 内存**（可能会因操作系统而异），和 **15GB-20GB 的存储空间**（一般可满足前几年的使用）。

### 内存

**简短结论：系统上总共 1GB 的 RAM 应该足够，但你可能降至 512MB 也能运行。**

对于小型站点（1-20 个活跃用户），GoToSocial 在内部缓存被填满的情况下，大约会占用 250MB 到 350MB 的 RAM：

在上图中，你可以看到 RAM 的使用会在负载期间突增。这种情况会在例如某个贴文被一个拥有众多粉丝的人转发时，或嵌入的 `ffmpeg` 二进制文件在解码或重新编码媒体文件为缩略图时（特别是较大视频文件）出现。

你应该在这种情况下预留一些余量，若有需要，可以[配置一些交换内存](#交换内存)。

!!! tip "提示"
    在内存受限的环境中，你可以将 `cache.memory-target` 设置为低于默认的 100MB （查看数据库配置选项[这里](../configuration/database.md#settings)）。设置为 50MB 已被证明可以正常运行。

    这将使总内存使用稍微降低，但代价是某些请求的延迟略高，因为 GtS 需要更频繁地访问数据库。

!!! info "为什么 `htop` 显示的内存使用比图中高？"
    如果你在服务器上运行 `top` 或 `htop` 或其他系统资源监测工具，GoToSocial 显示的保留内存可能比图中高。然而，这并不总是反映 GoToSocial 实际*使用*的内存。这种差异是由于 Go 运行时会比较保守地将内存释放回操作系统，因为与立即释放并在稍后需要时重新请求相比，保留空闲内存通常更划算。

### CPU

**简短结论：1 个不错的 CPU 核心应该足够。**

CPU 负载主要在处理媒体时（尤其是编码 blurhash）和/或同时处理大量联合请求时较高。只要不在同一台机器上运行其他 CPU 密集型任务，1 个 CPU 核心就能胜任。

### 存储

**简短结论：15GB-20GB 可用存储空间应足够使用几年。**

GoToSocial 使用存储来保存其数据库文件，以及存储和服务媒体文件，例如头像和附件等。你可以[配置 GtS 使用 S3 存储桶来存储媒体](../configuration/storage.md)。

对于媒体存储，以及[缓存的外站媒体文件存储](../admin/media_caching.md)，你应该预算大约 5GB-10GB 的空间。GoToSocial 会自动执行自我清理，在一段时间后从缓存中删除未使用的外站媒体。如果存储空间是个问题，你可以[调整媒体清理行为](../admin/media_caching.md#清理)以更频繁地清理和/或减少外站媒体的缓存时间。

!!! info "附注"
    如果你的 sqlite.db 文件或 Postgres 容量在一开始增长很快，请不要惊慌，这是正常的。当你首次部署实例并开始联合时，你的实例会迅速发现并存储来自其他实例的账号和贴文。然而，随着实例的长期部署，这种增长会逐渐减缓，因为你会自然而然地看到更少的新账号（即，你的实例尚未见过并因此尚未在数据库中存储的账号）。

### 单板计算机

GoToSocial 的轻量系统要求意味着它在配置良好的单板计算机上运行良好。如果在单板计算机上运行，你应该确保 GoToSocial 使用 USB 驱动器（最好是 SSD）来存储其数据库文件和媒体，而不是 SD 卡存储，因为后者通常太慢，不适合运行数据库。

### VPS

如果你决定使用 VPS，可以为自己建立一个便宜的运行 Linux 的服务器。大多数每月 €2-€5 的 VPS 能够为个人 GoToSocial 实例提供出色的性能。

[Hostwinds](https://www.hostwinds.com/) 是一个不错的选择：价格便宜，而且他们免费提供静态 IP 地址。

[Greenhost](https://greenhost.net) 也是一个好选择：它完全无 CO2 排放，但价格稍高。他们的 1GB、1 个 CPU 的 VPS 对于单个用户或小型实例来说效果很好。

!!! warning "云存储卷"
    并非所有的云 VPS 存储都相同，声称基于 SSD 的存储并不一定适合作为 GoToSocial 实例的运行环境。
    
    [Hetzner 云卷的性能](https://github.com/superseriousbusiness/gotosocial/issues/2471#issuecomment-1891098323)没有保证，且延迟波动较大。这会导致你的 GoToSocial 实例表现不佳。

!!! danger "Oracle 免费套餐"
    如果你打算与多个其他实例和用户联合，[Oracle 云免费套餐](https://www.oracle.com/cloud/free/) 服务器不适合用于 GoToSocial 部署。
    
    在 Oracle 云免费套餐上运行的 GoToSocial 管理员报告说，他们的实例在中等负载期间非常慢或无响应。这很可能是由于内存或存储延迟，导致即使是简单的数据库查询也要很长时间才能运行。

### 发行版系统要求

请务必检查你的发行版的系统需求，特别是内存。许多发行版有基线要求，在不满足它们的系统上运行会造成问题，除非你进行进一步的调整和优化。

Linux:

* [Arch Linux][archreq]: `512MB` RAM
* [Debian][debreq]: `786MB` RAM
* [Ubuntu][ubireq]: `1GB` RAM
* [RHEL 8+][rhelreq] 及其衍生版本: `1.5GB` RAM
* [Fedora][fedorareq]: `2GB` RAM

BSD 家族的发行版在内存要求方面记录较少，但普遍预期 `128MB` 以上就足够。

[archreq]: https://wiki.archlinux.org/title/installation_guide
[debreq]: https://www.debian.org/releases/stable/amd64/ch02s05.en.html
[ubireq]: https://ubuntu.com/server/docs/installation
[rhelreq]: https://access.redhat.com/articles/rhel-limits#minimum-required-memory-3
[fedorareq]: https://docs.fedoraproject.org/en-US/fedora/latest/release-notes/welcome/Hardware_Overview/#hardware_overview-specs

## 数据库

GoToSocial 支持 SQLite 和 Postgres 作为数据库驱动。尽管理论上可以在 SQLite 和 Postgres 之间切换，但我们目前没有工具支持这项操作，因此你在开始时应慎重考虑数据库的选择。

SQLite 是默认的驱动，并已被证明在 1-30 用户范围内的实例表现出色。

!!! danger "网络存储上的 SQLite"
    不要将你的 SQLite 数据库放在外部存储上，无论是 NFS/Samba、iSCSI 卷，如 Ceph/Gluster，或者你的云供应商的网络卷存储解决方案。

    更多信息参见[网络存储上的 SQLite](../advanced/sqlite-networked-storage.md)。

如果你计划在一个实例上托管更多用户，你可能希望改用 Postgres，因为它提供了数据库集群和冗余的可能性，尽管这会增加一些复杂性。

无论你选择哪种数据库驱动，为了获得良好的性能，它们都应在快速、稳定的低延迟存储上运行。虽然可以在网络附加存储上运行数据库，但这会增加可变延迟和网络拥堵，还有源存储上的潜在 I/O 争用。

!!! tip "提示"
    请[备份你的数据库](../admin/backup_and_restore.md)。数据库包含实例和任何用户账户的加密密钥。如果丢失这些密钥，你将无法再次从同一域进行联合！

## 域名

为了和其他实例进行联合，你需要一个域名，如 `example.org`。你可以通过任何域名注册商注册域名，例如 [Namecheap](https://www.namecheap.com/)。确保你选择的注册商也允许你管理 DNS 条目，以便将你的域指向运行 GoToSocial 实例的服务器 IP。

用户通常会出现在顶级域下，例如 `@me@example.org`，但这不是必须的。完全可以在 `@me@social.example.org` 下创建用户。很多人更喜欢在顶级域下创建用户，因为输入起来更短，但你可以使用任何你控制的（子域）。

可以拥有形如 `@me@example.org` 的用户名，但让 GoToSocial 运行在 `social.example.org`。这通过区分 API 域（称为“实例域名”）和用户名用的域（称为“账号域名”）来实现。

如果你打算这样部署 GoToSocial 实例，请阅读[分域部署](../advanced/host-account-domain.md)文档以了解详细信息。

!!! danger "危险"
    无法在联合已经事实发生后安全地更改实例域名和账号域名。这需要重新生成数据库，并在任何已联合的服务器造成混乱情况。一旦你的实例域名和账号域名设置好，便不可更改。

## TLS

为了实现联合，你必须使用 TLS。大多数实现，包括 GoToSocial，通常会拒绝通过未加密的传输进行联合。

GoToSocial 内置 Lets Encrypt 证书配置支持。它也可以从磁盘加载证书。如果你有连接到 GoToSocial 的反向代理，可以在代理层处理 TLS。

!!! tip "提示"
    请确保配置使用现代版本的 TLS，TLSv1.2 及更高版本，以确保服务器和客户端之间的通信安全。当 GoToSocial 处理 TLS 终端时，这会自动为你配置。如果使用反向代理，请使用 [Mozilla SSL 配置生成器](https://ssl-config.mozilla.org/)。

## 端口

GoToSocial 需要开放端口 `80` 和 `443`。

* `80` 用于 Lets Encrypt。因此，如果不使用内置的 Lets Encrypt 配置，则不需要开放。
* `443` 用于通过 TLS 服务 API，并且是与其联合的任何实例尝试连接的端口。

如果你无法在机器上开放 `443` 和 `80` 端口，不要担心！你可以在 GoToSocial 中配置这些端口，但还需要配置端口转发，以将 `443` 和 `80` 上的流量准确转发到你选择的端口。

!!! tip "提示"
    你应该在机器上配置防火墙，并配置一些防范暴力 SSH 登录尝试的保护措施。参阅我们的[防火墙文档](../advanced/security/firewall.md)以获取配置建议和可帮助你的工具。

## 集群 / 多节点部署

GoToSocial 不支持[集群或任何形式的多节点部署](https://github.com/superseriousbusiness/gotosocial/issues/1749)。

尽管多个 GtS 实例可以使用相同的 Postgres 数据库和共享的本地存储或相同的对象桶，但 GtS 依赖于大量的内部缓存以保持高效。没有同步这些实例缓存的机制。没有它，你会得到各种奇怪和不一致的行为。不要这样做！

## 调优

除了[示例配置文件](https://github.com/superseriousbusiness/gotosocial/blob/main/example/config.yaml)中的众多实例调优选项之外，你还可以对运行 GoToSocial 实例的机器进行额外的调优。

### 交换内存

除非你在进行这种调优并处理由不使用交换内存可能产生的问题方面有经验，否则你应该按照你的发行版本或主机供应商的建议配置适量的交换内存。如果你的发行版本或主机供应商没有提供指导，你可以使用以下经验法则为服务器配置交换内存：

* 小于 2GB 的 RAM：交换内存 = RAM × 2
* 大于 2GB 的 RAM：交换内存 = RAM，最高可达 8G

!!! tip "配置交换内存活跃度"
    Linux 的内存交换得很早。这在服务器上通常是不必要的，并且在数据库的情况下可能导致不必要的延迟。虽然在需要时让系统进行交换是好的，但可以通过告诉它在早期交换时保守一些来帮助提升性能。这个在 Linux 上的配置是通过更改 `vm.swappiness` [sysctl][sysctl] 值完成的。
    
    默认值是 `60`。你可以将其降低到 `10` 作为起点并留意观察。运行更低的值也是可能的，但这可能没有必要。要使该值持久化，你需要在 `/etc/sysctl.d/` 中放置一个配置文件。

虽然可以在没有交换内存的情况下运行系统，但为了安全地做到这一点并确保一致的性能和服务可用性，你需要相应调整内核、系统和工作负载。这需要对内核的内存管理系统及你所运行的工作负载的内存使用模式有良好的理解。

!!! tip "提示"
    交换内存用于确保内核可以高效地回收内存。这在系统没有经历内存争用时也很有用，比如在进程启动时仅使用过的内存腾出。这允许更多活跃使用的东西被缓存于内存中。内存交换不是让你的程序变慢的原因。内存争用才是造成缓慢的原因。

[sysctl]: https://man7.org/linux/man-pages/man8/sysctl.8.html

### 内存和 CPU 限制

可以限制 GoToSocial 实例可以消耗的内存或 CPU 的数量。在 Linux 上可以使用 [CGroups v2 资源控制器][cgv2] 来做到这一点。

你可以使用 [systemd 资源控制设置][systemdcgv2]、[OpenRC cgroup 支持][openrccgv2] 或 [libcgroup CLI][libcg] 为进程配置限制。如果你想在系统经历内存压力时保护 GoToSocial，可以查看 [`memory.low`][cgv2mem]。

[cgv2]: https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html
[systemdcgv2]: https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html
[openrccgv2]: https://wiki.gentoo.org/wiki/OpenRC/CGroups
[libcg]: https://github.com/libcgroup/libcgroup/
[cgv2mem]: https://docs.kernel.org/admin-guide/cgroup-v2.html#memory-interface-files
