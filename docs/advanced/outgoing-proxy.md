# Outgoing HTTP proxy

GoToSocial supports canonical environment variables for configuring the use of an HTTP proxy for outgoing requets:

* `HTTP_PROXY`
* `HTTPS_PROXY`
* `NO_PROXY`

The lowercase versions of these environment variables are also recognised. `HTTPS_PROXY` takes precedence over `HTTP_PROXY` for https requests.

The environment values may be either a complete URL or a `host[:port]`, in which case the "http" scheme is assumed. The schemes "http", "https", and "socks5" are supported.

## systemd

When running with systemd, you can add the necessary environment variables using the `Environment` option in the `Service` section.

How to do so is documented in the [`systemd.exec` manual](https://www.freedesktop.org/software/systemd/man/systemd.exec.html#Environment).

## Container runtime

Environment variables can be set in the compose file under the `environment` key. You can also pass them on the CLI to Docker or Podman's `run` command with `-e KEY=VALUE` or `--env KEY=VALUE`.
