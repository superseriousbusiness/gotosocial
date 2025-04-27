# Application sandboxing

By sandboxing the GoToSocial binary it's possible to control which parts of the system GoToSocial can access, and limit which things it can read and write. This can be helpful to ensure that even in the face of a security issue in GoToSocial, an attacker is severely hindered in escalating their privileges and gaining a foothold on your system.

!!! note
    As GoToSocial is still early in its development, the sandboxing policies we ship may get out of date. If you happen to run into this, please raise an issue on the issue tracker or better yet submit a PR to help us fix it.

Different distributions have different sandboxing mechanisms they prefer and support:

* **AppArmor** for the Debian or Ubuntu family of distributions or OpenSuSE, including when running with Docker
* **SELinux** for the Red Hat/Fedora/CentOS family of distributions or Gentoo

!!! warning "Containers and sandboxing"
    Running GoToSocial as a container does not in and of itself provide much additional security. Despite their name, "containers do not contain". Containers are a distribution mechanism, not a security sandbox. To further secure your container you can instruct the container runtime to load the AppArmor profile and look into limiting which syscalls can be used using a seccomp profile.

## AppArmor

We ship an example AppArmor policy for GoToSocial, which you can retrieve and install as follows:

```sh
$ curl -LO 'https://codeberg.org/superseriousbusiness/gotosocial/raw/main/example/apparmor/gotosocial'
$ sudo install -o root -g root gotosocial /etc/apparmor.d/gotosocial
$ sudo apparmor_parser -Kr /etc/apparmor.d/gotosocial
```

!!! tip
    The provided AppArmor example is just intended to get you started. It will still need to be edited depending on your exact setup; consult the comments in the example profile file for more information.

With the policy installed, you'll need to configure your system to use it to constrain the permissions GoToSocial has.

You can disable the policy like this:

```sh
$ sudo apparmor_parser -R /etc/apparmor.d/gotosocial
$ sudo rm -vi /etc/apparmor.d/gotosocial
```
Don't forget to roll back any configuration changes you made that load the AppArmor policy.

### systemd

Add the following to the systemd service, or create an override:

```ini
[Service]
...
AppArmorProfile=gotosocial
```

Reload systemd and restart GoToSocial:

```
$ systemctl daemon-reload
$ systemctl restart gotosocial
```

### Containers

!!! tip
    You should review the [Docker](https://docs.docker.com/engine/security/apparmor/) or [Podman](https://docs.podman.io/en/latest/markdown/options/security-opt.html) documentation on AppArmor.

When using our example Compose file, you can tell it to load the AppArmor policy by tweaking it like so:

```yaml
services:
  gotosocial:
    ...
    security_opt:
      - apparmor=gotosocial
```

When launching the container with `docker run` or `podman run`, you'll need the `--security-opt="apparmor=gotosocial"` command line flag.

## SELinux

!!! note
    SELinux can only be used in combination with the [binary installation](../../getting_started/installation/metal.md) method. SELinux cannot be used to constrain GoToSocial when running in a container.

The SELinux policy is maintained by the community in the [`lzap/gotosocial-selinux`](https://github.com/lzap/gotosocial-selinux) repository on GitHub. Make sure to read its documentation, review the policy before using it and use their issue tracker for any support requests around the SELinux policy.
