# 对 GoToSocial 可执行文件进行沙盒处理

通过对 GoToSocial 二进制文件进行沙盒化，可以控制 GoToSocial 能访问系统的哪些部分，并限制其读写权限。这有助于确保即使在 GoToSocial 出现安全问题时，攻击者也很难提升权限，进而在系统上立足。

不同发行版有其偏好的沙盒机制：

* **AppArmor** 适用于 Debian 或 Ubuntu 系列及 OpenSuSE，包括在 Docker 中的运行时
* **SELinux** 适用于 Red Hat/Fedora/CentOS 系列或 Gentoo

## AppArmor

我们提供了一个 GoToSocial 的 AppArmor 示例策略，你可以按以下步骤获取并安装：

```sh
$ curl -LO 'https://codeberg.org/superseriousbusiness/gotosocial/raw/main/example/apparmor/gotosocial'
$ sudo install -o root -g root gotosocial /etc/apparmor.d/gotosocial
$ sudo apparmor_parser -Kr /etc/apparmor.d/gotosocial
```

安装策略后，你需要配置系统以使用该策略来限制 GoToSocial 的权限。

你可以这样禁用该策略：

```sh
$ sudo apparmor_parser -R /etc/apparmor.d/gotosocial
$ sudo rm -vi /etc/apparmor.d/gotosocial
```
别忘了回滚你所做的任何加载 AppArmor 策略的配置更改。

### systemd

在 systemd 服务中添加以下内容，或创建一条覆盖规则：

```ini
[Service]
...
AppArmorProfile=gotosocial
```

重载 systemd 并重新启动 GoToSocial：

```sh
$ systemctl daemon-reload
$ systemctl restart gotosocial
```

### 容器

使用我们的示例 Compose 文件时，可以通过以下方式告知其加载 AppArmor 策略：

```yaml
services:
  gotosocial:
    ...
    security_opt:
      - apparmor=gotosocial
```

在使用 `docker run` 或 `podman run` 启动容器时，需要使用 `--security-opt="apparmor=gotosocial"` 命令行标志。

## SELinux

SELinux 策略由社区在 GitHub 上的 [`lzap/gotosocial-selinux`](https://github.com/lzap/gotosocial-selinux) 仓库维护。请务必阅读其文档，在使用前查看策略，并使用其问题跟踪器获取有关 SELinux 策略的支持请求。
