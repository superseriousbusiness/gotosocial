# Getting Started

## Quick and Dirty

### 1: Domain Name

Get a domain name -- [Namecheap](https://www.namecheap.com/) is a good place to do this, but you can use any domain name registrar that lets you manage your own DNS.

### 2: VPS

Spin yourself up a cheap VPS with Linux running on it, or get a homeserver ready with Ubuntu Server or something similar.

[Hostwinds](https://www.hostwinds.com/) is a good option here: it's cheap and they throw in a static IP address for free.

This guide won't go into running [UFW](https://www.digitalocean.com/community/tutorials/how-to-set-up-a-firewall-with-ufw-on-ubuntu-18-04) and [Fail2Ban](https://linuxize.com/post/install-configure-fail2ban-on-ubuntu-20-04/) but you absolutely should do that. Leave ports `443` and `80` open.

### 3: DNS

Point your domain name towards the server you just spun up.

### 4: Postgres

Install [Postgres](https://www.postgresql.org/download/) on your server and run it.

If you have [Docker](https://docs.docker.com/engine/install/ubuntu/) installed on your server, this is as easy as running:

```bash
docker run -d --network host --user postgres -e POSTGRES_PASSWORD=some_password postgres 
```

### 5: Build the Binary

On your local machine (not your server), with Go installed, clone the GoToSocial repository, and build the binary with `go build ./cmd/gotosocial`.

### 6: Prepare VPS

On the VPS or your homeserver, make the directory that GoToSocial will run from, and the directory it will use as storage:

```bash
mkdir /gotosocial && mkdir /gotosocial/storage
```

### 7: Copy Binary

Copy your binary from your local machine onto the VPS, using something like the following command (where `example.org` is the domain you set up in step 1):

```bash
scp ./gotosocial root@example.org:/gotosocial/gotosocial
```

Replace `root` with whatever user you're actually running on your remote server.

### 8: Copy Web Dir

GoToSocial uses some web templates and static assets, so you need to copy these over to your VPS as well (where `example.org` is the domain you set up in step 1):

```bash
scp -r ./web root@example.org:/gotosocial/
```

### 9: Run the Binary

Everything is in place now.

First cd into the directory you created on your remote machine in step 6:

```bash
cd /gotosocial
```

Then start the GoToSocial server with the following command (where `example.org` is the domain you set up in step 1, and `some_password` is the password you set for Postgres in step 4):

```bash
./gotosocial --host example.org --storage-serve-host example.org --letsencrypt-enabled=true server start
```

The server should now start up and you should be able to access the splash page by navigating to your domain in the browser. Note that it might take up to a minute or so for your LetsEncrypt certificates to be created for the first time, so refresh a few times if necessary.

### 10: Create and confirm your user

You can use the GoToSocial binary to also create, confirm, and promote your user account.

#### Create

Run the following command to create a new account:

```bash
./gotosocial --host example.org admin account create --username some_username --email some_email@whatever.org --password SOME_PASSWORD
```

In the above command, replace `example.org` with your domain, `some_username` with your desired username, `some_email@whatever.org` with the email address you want to associate with your account, and `SOME_PASSWORD` with a secure password.

#### Confirm

Run the following command to confirm the account you just created:

```bash
./gotosocial --host example.org admin account confirm --username some_username
```

Replace `example.org` with your domain and `some_username` with the username of the account you just created.

#### Promote

If you want your user to have admin rights, you can promote them using a similar command:

```bash
./gotosocial --host example.org admin account promote --username some_username
```

Replace `example.org` with your domain and `some_username` with the username of the account you just created.

### 11. Login!

You should now be able to log in to your instance using the email address and password of the account you just created. We recommend using [Pinafore](https://pinafore.social) or Tusky for this.
