# Configuration Overview

GoToSocial aims to be as configurable as possible, to fit lots of different use cases.

We try to provide sensible defaults wherever possible, but you can't run a GoToSocial instance without managing *some* configuration.

## Configuration Methods

There are three different methods for configuring a GoToSocial instance, which can be combined depending on your setup.

### Configuration File

The easiest way to configure GoToSocial is to pass a configuration file to to the `gotosocial server start` command, for example:

```bash
gotosocial --config-path ./config.yaml server start
```

The command expects a file in [YAML](https://en.wikipedia.org/wiki/YAML) or [JSON](https://en.wikipedia.org/wiki/JSON) format.

An example configuration file, with an explanation of each of the config fields, with default and example values, can be found [here](https://github.com/superseriousbusiness/gotosocial/blob/main/example/config.yaml). This example file is also included in release downloads. 

It's recommended to create your own configuration file with only the settings you need to change for your installation. This ensures you don't have to reconcile changes to defaults or adding/updating/removing settings from your configuration file that you haven't changed away from the defaults on every release.

#### Mounting in container

It can be necessary to have a `config.yaml` in a container, as some settings aren't easily managed with environment variables or command line flags.

To do so, create a `config.yaml` on the host, mount it in the container and then tell GoToSocial to pickup the configuration file. The latter can be done by either setting the command the container is run with to `--config-path /path/inside/container/to/config.yaml` or using the `GTS_CONFIG_PATH` environment variable.

For a compose file, you can amend the configuration like so:

```yaml
services:
  gotosocial:
    command: ["--config-path", "/gotosocial/config.yaml"]
    volumes:
      - type: bind
        source: /path/on/the/host/to/config.yaml
        target: /gotosocial/config.yaml
        read_only: true
```

Or, for the environment variable:

```yaml
services:
  gotosocial:
    environment:
      GTS_CONFIG_PATH: /gotosocial/config.yaml
    volumes:
      - type: bind
        source: /path/on/the/host/to/config.yaml
        target: /gotosocial/config.yaml
        read_only: true
```

For the Docker or Podman command line, pass a [mount specification](https://docs.podman.io/en/latest/markdown/podman-run.1.html#mount-type-type-type-specific-option).

Then when using `docker run` or `podman run`, pass `--config-path /gotosocial/config.yaml` as the command, for example:

```sh
podman run \
    --mount type=bind,source=/path/on/the/host/to/config.yaml,destination=/gotosocial/config.yaml,readonly \
    docker.io/superseriousbusiness/gotosocial:latest \
    --config-path /gotosocial/config.yaml
```

Using the `GTS_CONFIG_PATH` environment variable instead:

```sh
podman run \
    --mount type=bind,source=/path/on/the/host/to/config.yaml,destination=/gotosocial/config.yaml,readonly \
    --env 'GTS_CONFIG_PATH=/gotosocial/config.yaml' \
    docker.io/superseriousbusiness/gotosocial:latest
```

### Environment Variables

You can also configure GoToSocial by setting [environment variables](https://en.wikipedia.org/wiki/Environment_variable). These environment variables follow the format:

1. Prepend `GTS_` to the config flag.
2. Uppercase-all.
3. Replace dash (`-`) with underscore (`_`).

So for example, instead of setting `media-image-max-size` to `2097152` in your config.yaml, you could set the environment variable:

```text
GTS_MEDIA_IMAGE_MAX_SIZE=2097152
```

If you're in doubt about any of the names of these environment variables, just check the `--help` for the subcommand you're using.

!!! tip "Environment variable arrays"
    
    If you need to use an environment variable to set a configuration option that accepts an array, provide each value in a comma-separated list.
    
    For example, `instance-languages` may be set in the config.yaml file as an array like so: `["nl", "de", "fr", "en"]`. To set the same values as an environment variable, use: `GTS_INSTANCE_LANGUAGES="nl,de,fr,en"`

### Command Line Flags

Finally, you can set configuration values using command-line flags, which you pass directly when you're running a `gotosocial` command. For example, instead of setting `media-image-max-size` in your config.yaml, or with an environment variable, you can pass the value directly through the command line:

```bash
gotosocial server start --media-image-max-size 2097152 
```

If you're in doubt about which flags are available, check `gotosocial --help`.

## Priority

The above configuration methods override each other in the order in which they were listed.

```text
command line flags > environment variables > config file
```

That is, if you set `media-image-max-size` to `2097152` in your config file, but then *ALSO* set the environment variable `GTS_MEDIA_MAX_IMAGE_SIZE=9999999`, then the final value will be `9999999`, because environment variables have a *higher priority* than values set in config.yaml.

Command line flags have the highest priority, so if you set `--media-image-max-size 13121312`, then the final value will be `13121312` regardless of what you've set elsewhere.

This means in cases where you want to just try changing one thing, but don't want to edit your config file, you can temporarily use an environment variable or a command line flag to set that one thing.

## Default Values

Reasonable default values are provided for *most* of the configuration parameters, except in cases where a custom value is absolutely required.

See the [example config file](https://github.com/superseriousbusiness/gotosocial/blob/main/example/config.yaml) for the default values, or run `gotosocial --help`.

## `GTS_WAZERO_COMPILATION_CACHE`

On startup, GoToSocial compiles embedded WebAssembly `ffmpeg` and `ffprobe` binaries into [Wazero](https://wazero.io/)-compatible modules, which are used for media processing without requiring any external dependencies.

To speed up startup time of GoToSocial, you can cache the compiled modules between restarts so that GoToSocial doesn't have to compile them on every startup from scratch.

You can instruct GoToSocial on where to store the Wazero artifacts by setting the environment variable `GTS_WAZERO_COMPILATION_CACHE` to a directory, which will be used by GtS to store two smallish artifacts of ~50MiB or so each (~100MiB total).

For an example of this in action, see the [docker-compose.yaml](https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/example/docker-compose/docker-compose.yaml), and the [gotosocial.service](https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/example/gotosocial.service) example files.

If you want to provide this value to GtS outside of systemd or Docker, you can do so in the following manner when starting up your GtS server:

```bash
GTS_WAZERO_COMPILATION_CACHE=~/gotosocial/.cache ./gotosocial --config-path ./config.yaml server start
```
