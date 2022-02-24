# Docker

The Official GoToSocial docker images are provided through [docker hub](https://hub.docker.com/r/superseriousbusiness/gotosocial "docker hub gotosocial").

Currently the docker images only support the amd64 architechture. In the future we will add support for more architechtures. For now if you wish to run GoToSocial on, for example, a Raspberry Pi, you may [build the docker image yourself](https://github.com/ForestJohnson/gotosocial/compare/main...forest-jank-multi-arch-docker-build)

GoToSocial can be configured using [Environment Variables](../configuration/index.md#environment-variables) if you wish, allowing your GoToSocial configuration to be embedded inside your docker container configuration.

## Run with Docker Compose (recommended)
Assuming you will be using something like [docker compose](https://docs.docker.com/compose/ "Docker Compose Docs") to configure your GoToSocial docker container, you might want to follow the next Steps.

### Create a Working Dir
You need a Working Directory in which the data of the PostgreSQL and the GoToSocial container will be located, so create this directory for example with the following command. 
The directory can be located where you want it to be later.

```shell
mkdir -p /docker/gotosocial
cd /docker/gotosocial
```
### Get the latest docker-compose.yaml and config.yaml
You can get an example [docker-compose.yaml](../../example/docker-compose/docker-compose.yaml "Example docker-compose.yaml") and [config.yaml](../../example/config.yaml "Example config.yaml") here, which you can download with wget.

```shell
wget https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/example/docker-compose/docker-compose.yaml
wget https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/example/config.yaml
```

### Edit the docker-compose.yaml
You can modify the docker-compose.yaml to our needs, but in any case you should generate a Postgres password and bind it as environment variable into the postgreSQL container. For this we can write the password directly into the docker-compose.yaml like in the example or we create an [.env file](https://docs.docker.com/compose/environment-variables/#the-env-file "Docker Docs") with which we load the environment variables into the container.

```shell
$EDITOR docker-compose.yaml
```
### Edit the config.yaml
When we want to use the config.yaml, we should make the following changes to config.yaml.
| Config Option   | Value  |
| --------------- | ------ |
| host            | Hostname of your Inctanse e.g. gts.example.com |
| account-domain  | Domain to use when federating profiles e.g. gts.example.com |
| trusted-proxies | We need to trust our hostmashine and the Docker Network e.g.<br>- "127.0.0.1/32"<br>- "10.0.0.0/8"<br>- "172.16.0.0/12"<br>- "192.168.0.0/16" |
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

### Create your first User
First of all, we have to find the right Container ID so we will check the Container ID with `docker ps` or if we got tons of containers running `docker ps -f NAME=gotosocial`.
If we execute the command like this, we will get an output similar to the following:

```shell
CONTAINER ID   IMAGE                                   COMMAND                  CREATED          STATUS          PORTS                      NAMES
e190f1e6335f   superseriousbusiness/gotosocial:0.2.0   "/gotosocial/gotosoc…"   12 minutes ago   Up 12 minutes   127.0.0.1:8080->8080/tcp   docker-compose-gotosocial-1
5a2c56181ada   postgres:14-alpine                      "docker-entrypoint.s…"   22 minutes ago   Up 19 minutes   5432/tcp                   docker-compose-gotosocial_postgres-1
```

Now we take the container ID of the container with the image superseriousbusiness/gotosocial:0.2.0 and build ourselves the following commands.

```shell
# Creates a User
docker exec -ti $CONTAINER_ID /gotosocial/gotosocial --config-path /config/config.yaml admin account create --username $USERNAME --email $USEREMAIL --password $SuperSecurePassword
# Confirms the User, so that the User can LogIn
docker exec -ti $CONTAINER_ID /gotosocial/gotosocial --config-path /config/config.yaml admin account confirm --username $USERNAME
# Makes the User to an Admin
docker exec -ti $CONTAINER_ID /gotosocial/gotosocial --config-path /config/config.yaml admin account promote --username $USERNAME
```

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
GTS_TRUSTED_PROXIES=0.0.0.0/0
GTS_HOST=gotosocial.example.com
GTS_ACCOUNT_DOMAIN=gotosocial.example.com
GTS_DB_TYPE=sqlite
GTS_DB_ADDRESS=/gotosocial/database/sqlite.db
GTS_STORAGE_SERVE_PROTOCOL=https
GTS_STORAGE_SERVE_HOST=gotosocial.example.com
GTS_STORAGE_SERVE_BASE_PATH=/gotosocial/storage
GTS_LETSENCRYPT_ENABLED=false
```
</details>

## (optional) NGINX Config
The following NGINX config is just an example of what this might look like. In this case we assume that a valid SSL certificate is present. For this you can get a valid certificate from [Let's Encrypt](https://letsencrypt.org "Let's Encrypt Homepage") with the [cerbot](https://certbot.eff.org "Certbot's Homepage").

```shell
server {
  listen 80;
  listen [::]:80;
  server_name gts.example.com;

  location /.well-known/acme-challenge/ {
    default_type "text/plain";
    root /var/www/certbot;
  }
  location / { return 301 https://$host$request_uri; }
}

server {
  listen 443 ssl http2;
  listen [::]:443 ssl http2;
  server_name gts.example.com;

  #############################################################################
  # Certificates                                                              #
  # you need a certificate to run in production. see https://letsencrypt.org/ #
  #############################################################################
  ssl_certificate     /etc/letsencrypt/live/gts.example.com/fullchain.pem;
  ssl_certificate_key /etc/letsencrypt/live/gts.example.com/privkey.pem;

  location ^~ '/.well-known/acme-challenge' {
    default_type "text/plain";
    root /var/www/certbot;
  }

  ###########################################
  # Security hardening (as of Nov 15, 2020) #
  # based on Mozilla Guideline v5.6         #
  ###########################################

  ssl_protocols             TLSv1.2 TLSv1.3;
  ssl_prefer_server_ciphers on;
  ssl_ciphers "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305";
  ssl_session_timeout       1d; # defaults to 5m
  ssl_session_cache         shared:SSL:10m; # estimated to 40k sessions
  ssl_session_tickets       off;
  ssl_stapling              on;
  ssl_stapling_verify       on;
  ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;
  # HSTS (https://hstspreload.org), requires to be copied in 'location' sections that have add_header directives
  add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload";


  location / {
    proxy_pass         http://127.0.0.1:8080;

    proxy_set_header   Host             $host;
    proxy_set_header   Connection       $http_connection;
    proxy_set_header   X-Real-IP        $remote_addr;
    proxy_set_header   X-Forwarded-For  $proxy_add_x_forwarded_for;
    proxy_set_header   X-Scheme         $scheme;
  }

}
```