# Bare metal

This guide walks you through getting GoToSocial up and running on bare metal using the official binary releases.

## Prepare VPS

In a terminal on the VPS or your homeserver, make the directory that GoToSocial will run from, the directory it will use as storage, and the directory it will store LetsEncrypt certificates in.

This means we need the following hierarchy:

```
.
└── gotosocial
    └── storage
        └── certs
```

You can create that in one go with:

```bash
mkdir -p /gotosocial/storage/certs
```

If you don't have root permissions on the machine, use something like `~/gotosocial` instead.

## Download Release

In a terminal on the VPS or your homeserver, change into the base directory for GoToSocial you just created above:

```bash
cd /gotosocial
```

Now, download the latest GoToSocial release archive corresponding to the operating system and architecture you're running on.

!!! tip
    You can find the list of releases [right here](https://codeberg.org/superseriousbusiness/gotosocial/releases), arranged with the newest release at the top.

For example, to download a version for running on 64-bit Linux:

```bash
GTS_VERSION=X.Y.Z # replace this
GTS_TARGET=linux_amd64
wget https://codeberg.org/superseriousbusiness/gotosocial/releases/download/v${GTS_VERSION}/gotosocial_${GTS_VERSION}_${GTS_TARGET}.tar.gz
```

Then extract it:

```bash
tar -xzf gotosocial_${GTS_VERSION}_${GTS_TARGET}.tar.gz
```

This will put the `gotosocial` binary in your current directory, in addition to the `web` folder, which contains assets for the web frontend, and an `example` folder, which contains a sample configuration file.

!!! danger
    If you prefer to use a snapshot build of GoToSocial based on whatever code is currently on main, you can download [recent binary .tar.gz files from here](https://minio.s3.superseriousbusiness.org/browser/gotosocial-snapshots) (keyed by commit hash). Only do this if you know what you're doing, otherwise just take a stable release.

## Edit Configuration File

Create a new configuration file, based on the `config.yaml` from the `example` folder. You can copy the whole file, but make sure you only retain settings you've changed. This makes it easier to review configuration changes on release upgrades.

You'll probably need to change the following settings:

- Set `host` to whatever hostname you're going to be running the server on (eg., `example.org`).
- Set `port` to `443`.
- Set `storage-local-base-path` to the storage directory you created above (eg., `/gotosocial/storage`).
- Set `letsencrypt-enabled` to `true`.
- Set `letsencrypt-cert-dir` to the certificate storage directory you created above (eg., `/gotosocial/storage/certs`).

The above options assume you're using SQLite as your database. If you want to use Postgres instead, see [here](../../configuration/database.md) for the config options.

!!! info "Optional configuration"
    
    There are many other configuration options documented in the config.yaml file, which you can use to further customize the behavior of your GoToSocial instance. These use sensible defaults where possible, so you don't necessarily need to make any changes to them right now, but here are a few you may be interested in:
    
    - `instance-languages`: array of [BCP47 language tags](https://en.wikipedia.org/wiki/IETF_language_tag) which determines the preferred languages of your instance.
    - `media-remote-cache-days`: number of days to keep remote media cached in storage.
    - `smtp-*`: settings to allow your GoToSocial instance to connect to an email server and send notification emails.

    If you decide to set/change any of these variables later on, be sure to restart your GoToSocial instance after making the changes.

## Run the Binary

You can now run the binary.

Start the GoToSocial server with the following command:

```bash
./gotosocial --config-path ./config.yaml server start
```

The server should now start up and you should be able to access the splash page by navigating to your domain in the browser. Note that it might take up to a minute or so for your LetsEncrypt certificates to be created for the first time, so refresh a few times if necessary.

Note that for this example we're assuming that we're allowed to run on port 443 (standard https port), and that nothing else is running on this port.

## Create your user

You can use the GoToSocial binary to also create and promote your user account. This is all documented in our [Creating users](../user_creation.md) guide.

## Login

You should now be able to log in to your instance using the email address and password of the account you just created.

## (Optional) Enable the systemd service

If you don't like manually starting GoToSocial on every boot you might want to create a systemd service that does that for you.

First stop your GoToSocial instance.

Then create a new user and group for your GoToSocial installation:

```bash
sudo useradd -r gotosocial
sudo groupadd gotosocial
sudo usermod -a -G gotosocial gotosocial
```

Then make them the owner of your GoToSocial installation since they will need to read and write in it:

```bash
sudo chown -R gotosocial:gotosocial /gotosocial
```

You can find a `gotosocial.service` file in the `example` folder on [our repository](https://codeberg.org/superseriousbusiness/gotosocial/raw/branch/main/example/gotosocial.service) or your installation.

Copy it to `/etc/systemd/system/gotosocial.service`:

```bash
sudo cp /gotosocial/example/gotosocial.service /etc/systemd/system/
```

Then use `sudoedit /etc/systemd/system/gotosocial.service` to open the file in an editor. If you installed GoToSocial in a directory different from the `/gotosocial` path used in this guide, change the `ExecStart=` and `WorkingDirectory=` lines according to your installation.

!!! info "Running on ports 80 and 443"
    
    If you've been following this guide word for word, your GoToSocial instance will be configured to bind to ports 443 and 80, which are known as privileged ports. To allow the GoToSocial user to bind to these, you need to uncomment the line about `CAP_NET_BIND_SERVICE` in the service file by removing the leading `#`.
    
    Before:
    
    ```
    #AmbientCapabilities=CAP_NET_BIND_SERVICE
    ```
    
    After:
    
    ```
    AmbientCapabilities=CAP_NET_BIND_SERVICE
    ```
    
    If you later decide to run GoToSocial using a reverse proxy (see below) you may want to re-comment this line to remove the privileges, since the reverse proxy will bind to the privileged ports instead.

After you're done editing, save and close the file, and run the following command to enable the service:

```bash
sudo systemctl enable --now gotosocial.service
```

GoToSocial should now be up and running.

## (Optional) Reverse proxy

If you want to run other webservers on port 443 or want to add an additional layer of security you might want to use a [reverse proxy](../reverse_proxy/index.md). We have guides available for a couple of popular open source options and will gladly take pull requests to add more.
