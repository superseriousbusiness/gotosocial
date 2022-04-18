# Docker

The Official GoToSocial docker images are provided through [docker hub](https://hub.docker.com/r/superseriousbusiness/gotosocial "docker hub gotosocial").

GoToSocial can be configured using [Environment Variables](../configuration/index.md#environment-variables) if you wish, allowing your GoToSocial configuration to be embedded inside your docker container configuration.

## Run with Docker Compose (recommended)
This guide will lead you through the installation with [docker compose](https://docs.docker.com/compose/ "Docker Compose Docs"), so you might want to follow the next Steps.

### Create a Working Dir
You need a Working Directory in which the data of the PostgreSQL and the GoToSocial container will be located, so create this directory for example with the following command. 
The directory can be located where you want it to be later.

```shell
mkdir -p /docker/gotosocial
cd /docker/gotosocial
```
### Get the latest docker-compose.yaml and config.yaml
You can get an example [docker-compose.yaml](../../example/docker-compose/docker-compose.yaml "Example docker-compose.yaml") and [config.yaml](../../example/config.yaml "Example config.yaml") here, which you can download with wget for example.

```shell
wget https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/example/docker-compose/docker-compose.yaml
wget https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/example/config.yaml
```

### Edit the docker-compose.yaml
You can modify the docker-compose.yaml to your needs, but in any case you should generate a Postgres password and bind it as environment variable into the postgreSQL container. For this we can write the password directly into the docker-compose.yaml like in the example or we create an [.env file](https://docs.docker.com/compose/environment-variables/#the-env-file "Docker Docs") that will load the environment variables into the container. You may also want to check the current [GoToSocial version](https://github.com/superseriousbusiness/gotosocial/releases) and adjust the image in docker-compose.yaml.

```shell
$EDITOR docker-compose.yaml
```
### Edit the config.yaml
When we want to use the config.yaml, we should make the following changes to config.yaml.
| Config Option   | Value  |
| --------------- | ------ |
| host            | Hostname of your Inctanse e.g. gts.example.com |
| account-domain  | Domain to use when federating profiles e.g. gts.example.com |
| trusted-proxies | We need to trust our host machine and the Docker Network e.g.<br>- "127.0.0.1/32"<br>- "10.0.0.0/8"<br>- "172.16.0.0/12"<br>- "192.168.0.0/16" |
| db-address      | gotosocial_postgres |
| db-user         | gotosocial |
| db-password     | same password as postgres environment $POSTGRES_PASSWORD |

```shell
$EDITOR config.yaml
```
### Start GoToSocial

```shell
docker-compose up -d
```

After running this command, you should get an output like:
```shell
❯ docker-compose up -d
[+] Running 2/2
 ⠿ Container docker1-gotosocial_postgres-1  Started
 ⠿ Container docker1-gotosocial-1           Started
```

this names can be used to create your first user described below.

### Create your first User

Take the names from above command `docker-compose up -d` and replace $CONTAINER_NAME with the name e.g. `docker1-gotosocial-1`

```shell
# Creates a User
docker exec -ti $CONTAINER_NAME /gotosocial/gotosocial --config-path /config/config.yaml admin account create --username $USERNAME --email $USEREMAIL --password $SuperSecurePassword
# Confirms the User, so that the User can LogIn
docker exec -ti $CONTAINER_NAME /gotosocial/gotosocial --config-path /config/config.yaml admin account confirm --username $USERNAME
# Makes the User to an Admin
docker exec -ti $CONTAINER_NAME/gotosocial/gotosocial --config-path /config/config.yaml admin account promote --username $USERNAME
```

#### Lost the Name of the Container
If you forgot what the container name of your GoToSocial container was, you can figure it out with the command `docker ps -f NAME=gotosocial`.
If you execute the command, you will get an output similar to the following:

```shell
CONTAINER ID   IMAGE                                      COMMAND                  CREATED          STATUS          PORTS                      NAMES
e190f1e6335f   superseriousbusiness/gotosocial:$VERSION   "/gotosocial/gotosoc…"   12 minutes ago   Up 12 minutes   127.0.0.1:8080->8080/tcp   docker-compose-gotosocial-1
5a2c56181ada   postgres:14-alpine                         "docker-entrypoint.s…"   22 minutes ago   Up 19 minutes   5432/tcp                   docker-compose-gotosocial_postgres-1
```
Now you take the container name from the container with image superseriousbusiness/gotosocial:$VERSION and build ourselves the following commands.

## Run with Docker Run

You can run GoToSocial direct with `docker run` command.

<details>
  <summary>docker run with --env flag</summary>

```shell
docker run -e GTS_PORT='8080' -e GTS_PROTOCOL='https' -e GTS_TRUSTED_PROXIES='0.0.0.0/0' -e GTS_HOST='gotosocial.example.com' -e GTS_ACCOUNT_DOMAIN='gotosocial.example.com' -e GTS_DB_TYPE='sqlite' -e GTS_DB_ADDRESS='/gotosocial/database/sqlite.db' -e GTS_STORAGE_SERVE_PROTOCOL='https' -e GTS_STORAGE_SERVE_HOST='gotosocial.example.com' -e GTS_STORAGE_SERVE_BASE_PATH='/gotosocial/storage' -e GTS_LETSENCRYPT_ENABLED='false' -v $(pwd)/storage/:/gotosocial/storage/ -v $(pwd)/database/:/gotosocial/database/ -p 127.0.0.1:8080:8080 superseriousbusiness/gotosocial:0.2.0
```

</details>

<details>
  <summary>docker run with .env-file</summary>

```
docker run --env-file ./.env -v $(pwd)/storage/:/gotosocial/storage/ -v $(pwd)/database/:/gotosocial/database/ -p 127.0.0.1:8080:8080 superseriousbusiness/gotosocial:0.2.0
```

</details>

<details>
  <summary>Example .env File</summary>

```shell
$EDITOR .env
```

```
GTS_PORT=8080
GTS_PROTOCOL=https
GTS_TRUSTED_PROXIES=127.0.0.1 # should be the host machine and the Docker Network e.g. "127.0.0.1/32", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"
GTS_HOST=gotosocial.example.com
GTS_ACCOUNT_DOMAIN=gotosocial.example.com
GTS_DB_TYPE=sqlite
GTS_DB_ADDRESS=/gotosocial/database/sqlite.db
GTS_STORAGE_SERVE_BASE_PATH=/gotosocial/storage
GTS_LETSENCRYPT_ENABLED=false
```
</details>

## (optional) Reverse Proxy

If you want to run other webservers on port 433 or want to add an additional layer of security you might want to use [nginx](./nginx.md) or [Apache httpd](./apache-httpd.md) as reverse proxy
