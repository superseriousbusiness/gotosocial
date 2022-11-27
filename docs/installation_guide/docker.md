# Docker

The official GoToSocial Docker images are provided through [Docker Hub](https://hub.docker.com/r/superseriousbusiness/gotosocial).

Docker images are currently available for the following OS + architecture combinations:

Linux

- 386
- amd64
- arm6
- arm7
- arm64v8

FreeBSD

- amd64

Before following this guide, you should read the [system requirements](./index.md).

This guide assumes that you're using Linux.

## Run with Docker Compose

You can run GoToSocial using any orchestration system that can manage Docker containers ([Kubernetes](https://kubernetes.io/), [Nomad](https://www.nomadproject.io/), etc).

For simplicity's sake, this guide will lead you through the installation with [Docker Compose](https://docs.docker.com/compose), using SQLite as your database.

### Create a Working Dir

You need a working directory in which your docker-compose file will be located, and a directory for GoToSocial to store data in, so create these directories with the following command:

```bash
mkdir -p ~/gotosocial/data
```

Now change to the working directory you created:

```bash
cd ~/gotosocial
```

### Get the latest docker-compose.yaml

Use `wget` to download the latest [docker-compose.yaml](https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/example/docker-compose/docker-compose.yaml) example, which we'll customize for our needs:

```bash
wget https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/example/docker-compose/docker-compose.yaml
```

### Edit the docker-compose.yaml

Because GoToSocial can be configured using [Environment Variables](../configuration/index.md#environment-variables), we can skip mounting a config.yaml file into the container, to make our configuration simpler. We just need to edit the docker-compose.yaml file to change a few things.

First open the docker-compose.yaml file in your editor of choice. For example:

```bash
nano docker-compose.yaml
```

#### Version

If desired, update the GoToSocial Docker image tag to the version of GtS you want to use.

`latest`   - the default. This points to the latest stable release of GoToSocial.

`snapshot` - points to whatever code is currently on the main branch. Not guaranteed to be stable, will often be broken. Use with caution.

You can also replace `latest` with a specific GoToSocial version number. This is recommended when you want to make sure that you don't update your GoToSocial version by accident, which can cause problems.

The list of releases can be found [right here](https://github.com/superseriousbusiness/gotosocial/releases), with the newest release at the top. Replace `latest` in the docker-compose.yaml with the number of the desired release (without the leading `v` or trailing version name). So for example if you want to run [v0.3.1 Sleepy Sloth](https://github.com/superseriousbusiness/gotosocial/releases/tag/v0.3.1) for whatever reason, you should replace:

```text
image: superseriousbusiness/gotosocial:latest
```

with:

```text
image: superseriousbusiness/gotosocial:0.3.1
```

#### Host

Change the `GTS_HOST` environment variable to the domain you are running GoToSocial on.

#### User (optional / probably not necessary)

By default, Dockerized GoToSocial runs with Linux user/group `1000:1000`, which is fine in most cases. If you want to run as a different user/group, you should change the `user` field in the docker-compose.yaml accordingly.

For example, let's say you created the `~/gotosocial/data` directory for a user with id `1001`, and group id `1001`. If you now try to run GoToSocial without changing the `user` field, it will get a permissions error trying to open its database file in the directory. In this case, you would have to change the `user` field of the docker compose file to `1001:1001`.

#### LetsEncrypt (optional)

If you want to use [LetsEncrypt](../configuration/letsencrypt.md) for ssl certificates (https), you should also:

1. Change the value of `GTS_LETSENCRYPT_ENABLED` to `"true"`.
2. Remove the `#` before `- "80:80"` in the `ports` section.
3. (Optional) Set `GTS_LETSENCRYPT_EMAIL_ADDRESS` to a valid email address to receive certificate expiry warnings etc.

#### Reverse proxies

The default port bindings are for exposing GoToSocial directly and publicly. Remove the `#` in front the line that forwards `127.0.0.1:8080:8080` which makes port `8080` available only to the local host. Change that `127.0.0.1` if the reverse proxy is somewhere else.

To ensure [rate limiting](../api/ratelimiting.md) by IP works, remove the `#` in front of `GTS_TRUSTED_PROXIES` and set it to the IP the requests from the reverse proxy are coming from. That's usually the value of the `Gateway` field of the docker network.

```text
$ docker network inspect gotosocial_gotosocial
[
    {
        "Name": "gotosocial_gotosocial",
        [...]
        "IPAM": {
            "Driver": "default",
            "Options": null,
            "Config": [
                {
                    "Subnet": "172.19.0.0/16",
                    "Gateway": "172.19.0.1"
                }
            ]
        },
        [...]
```

In the example above, it would be `172.19.0.1`.

If unsure, skip the trusted proxies step, continue with the next sections, and once it's running get the `clientIP` from the docker logs.

### Start GoToSocial

With those small changes out of the way, you can now start GoToSocial with the following command:

```shell
docker-compose up -d
```

After running this command, you should get an output like:

```text
Creating network "gotosocial_gotosocial" with the default driver
Creating gotosocial ... done
```

If you want to follow the logs of GoToSocial, you can use:

```bash
docker logs -f gotosocial
```

If everything is OK, you should see something similar to the following:

```text
time=2022-04-19T09:48:35Z level=info msg=connected to SQLITE database
time=2022-04-19T09:48:35Z level=info msg=MIGRATED DATABASE TO group #1 (20211113114307, 20220214175650, 20220305130328, 20220315160814) func=doMigration
time=2022-04-19T09:48:36Z level=info msg=instance account example.org CREATED with id 01EXX0TJ9PPPXF2C4N2MMMVK50
time=2022-04-19T09:48:36Z level=info msg=created instance instance example.org with id 01PQT31C7BZJ1Q2Z4BMEV90ZCV
time=2022-04-19T09:48:36Z level=info msg=media manager cron logger: start[]
time=2022-04-19T09:48:36Z level=info msg=media manager cron logger: schedule[now 2022-04-19 09:48:36.096127852 +0000 UTC entry 1 next 2022-04-20 00:00:00 +0000 UTC]
time=2022-04-19T09:48:36Z level=info msg=started media manager remote cache cleanup job: will run next at 2022-04-20 00:00:00 +0000 UTC
time=2022-04-19T09:48:36Z level=info msg=listening on 0.0.0.0:8080
```

### Create your first User

Now that GoToSocial is running, you can execute commands inside the running container to create and promote your admin user.

First create a user (replace the username, email, and password with appropriate values):

```bash
docker exec -it gotosocial /gotosocial/gotosocial admin account create --username some_username --email someone@example.org --password 'some_very_good_password'
```

If you are running a version older than 0.6.0, you will need to manually confirm as well:

```bash
./gotosocial --config-path ./config.yaml admin account confirm --username some_username
```

Replace `some_username` with the username of the account you just created.

Now promote the user you just created to admin privileges:

```bash
docker exec -it gotosocial /gotosocial/gotosocial admin account promote --username some_username
```

When running these commands, you'll get a bit of output like the following:

```text
time=2022-04-19T09:53:29Z level=info msg=connected to SQLITE database
time=2022-04-19T09:53:29Z level=info msg=there are no new migrations to run func=doMigration
time=2022-04-19T09:53:29Z level=info msg=closing db connection
```

This is normal and indicates that the commands ran as expected.

### Done

GoToSocial should now be running on your machine! To verify this, open your browser and go to `http://localhost:443`. You should see the GoToSocial landing page. Well done!

## (Optional) Reverse Proxy

If you want to run other webservers on port 443, or want to add an additional layer of security you might want to add [NGINX](https://nginx.org), [Traefik](https://doc.traefik.io/traefik/), or [Apache httpd](https://httpd.apache.org/) into your docker-compose to use as a reverse proxy.
