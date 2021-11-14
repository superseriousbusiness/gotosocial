# Configuration Overview

GoToSocial aims to be as configurable as possible, to fit lots of different use cases.

We try to provide sensible defaults wherever possible, but you can't run a GoToSocial instance without managing *some* configuration.

In this section, we describe the different methods available for configuring GoToSocial, 

## Configuration Methods

There are three different methods for configuring a GoToSocial instance, which can be combined depending on your setup.

### Configuration File

The easiest way to configure GoToSocial is to pass a configuration file to to the `gotosocial server start` command, for example:

```bash
gotosocial --config-path ./config.yaml server start
```

The command expects a file in [YAML](https://en.wikipedia.org/wiki/YAML) format.

An example configuration file, with an explanation of each of the config fields, with default and example values, can be found [here](https://github.com/superseriousbusiness/gotosocial/blob/main/example/config.yaml).

This example file is included with release downloads, so you can just copy it and edit it to your needs without having to worry too much about what the hell YAML is.

### Environment Variables

You can also configure GoToSocial by setting [environment variables](https://en.wikipedia.org/wiki/Environment_variable). These environment variables generally follow the format:

```text
GTS_[CONFIG_SECTION]_[FIELD_NAME]
```

So for example, instead of setting `media.maxImageSize` to `2097152` in your config.yaml, you could set the environment variable:

```text
GTS_MEDIA_MAX_IMAGE_SIZE=2097152
```

If you're in doubt about any of the names of these environment variables, just check `gotosocial --help`. They're all listed there.

### Command Line Flags

Finally, you can set configuration values using command-line flags, which you pass directly when you're running a `gotosocial` command. For example, instead of setting `media.maxImageSize` in your config.yaml, or with an environment variable, you can pass the value directly through the command line:

```bash
gotosocial --media-max-image-size 2097152 server start
```

Note the weird order of the above command; `gotosocial` first, then all global variables (like the ones in the config.yaml), then the command you want to execute.

If you're in doubt about which flags are available, check `gotosocial --help`.

## Priority

The above configuration methods override each other in the order in which they were listed.

```text
command line flags > environment variables > config file
```

That is, if you set `media.maxImageSize` to `2097152` in your config file, but then *ALSO* set the environment variable `GTS_MEDIA_MAX_IMAGE_SIZE=9999999`, then the final value will be `9999999`, because environment variables have a *higher priority* than values set in config.yaml.

Command line flags have the highest priority, so if you set `--media-max-image-size 13121312`, then the final value will be `13121312` regardless of what you've set elsewhere.

This means in cases where you want to just try changing one thing, but don't want to edit your config file, you can temporarily use an environment variable or a command line flag to set that one thing.

## Default Values

Reasonable default values are provided for *most* of the configuration parameters, except in cases where a custom value is absolutely required.

See the [example config file](https://github.com/superseriousbusiness/gotosocial/blob/main/example/config.yaml) for the default values, or run `gotosocial --help`.
