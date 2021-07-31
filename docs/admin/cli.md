# GtS CLI Tool

GoToSocial compiles to an executable binary.

The standard way of using this binary is to run a server with the `gotosocial server start` command.

However, this binary can also be used as an admin tool.

Here's the full output of `gotosocial --help`, without the big list of global config options.

```text
NAME:
   gotosocial - a fediverse social media server

USAGE:
   gotosocial [global options] command [command options] [arguments...]

VERSION:
   0.1.0-SNAPSHOT a940a52

COMMANDS:
   server   gotosocial server-related tasks
   admin    gotosocial admin-related tasks
   testrig  gotosocial testrig tasks
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   [a huge list of global options -- too much to show here]
```

Under `COMMANDS`, you can see the standard `server` command. But there are also commands doing admin and testing etc, which will be explained in this document.

**Please note -- for all of these commands, you will still need to set the global options correctly so that the CLI tool knows how eg., how to connect to your database, which database to use, which host and account domain to use etc.**

You can set these global options using environment variables, passing them as CLI variables after the `gotosocial` part of the command (eg., `gotosocial --host example.org [commands]`), or by just pointing the CLI tool towards your config file (eg., `gotosocial --config-path ./config.yaml [commands]`).

## gotosocial admin

Contains `account` subcommands.

### gotosocial admin account create

This command can be used to create a new account on your instance.

`gotosocial admin account create --help`:

```text
NAME:
   gotosocial admin account create - create a new account

USAGE:
   gotosocial admin account create [command options] [arguments...]

OPTIONS:
   --username value  the username to create/delete/etc
   --email value     the email address of this account
   --password value  the password to set for this account
   --help, -h        show help (default: false)
```

Example:

```bash
gotosocial admin account create \
   --username some_username \
   --email someuser@example.org \
   --password 'somelongandcomplicatedpassword'
```

### gotosocial admin account confirm

This command can be used to confirm a user+account on your instance, allowing them to log in and use the account.

`gotosocial admin account confirm --help`:

```text
NAME:
   gotosocial admin account confirm - confirm an existing account manually, thereby skipping email confirmation

USAGE:
   gotosocial admin account confirm [command options] [arguments...]

OPTIONS:
   --username value  the username to create/delete/etc
   --help, -h        show help (default: false)
```

Example:

```bash
gotosocial admin account confirm --username some_username
```

### gotosocial admin account promote

This command can be used to promote a user to admin.

`gotosocial admin account promote --help`:

```text
NAME:
   gotosocial admin account promote - promote an account to admin

USAGE:
   gotosocial admin account promote [command options] [arguments...]

OPTIONS:
   --username value  the username to create/delete/etc
   --help, -h        show help (default: false)
```

Example:

```bash
gotosocial admin account promote --username some_username
```

### gotosocial admin account demote

This command can be used to demote a user from admin to normal user.

`gotosocial admin account demote --help`:

```text
NAME:
   gotosocial admin account demote - demote an account from admin to normal user

USAGE:
   gotosocial admin account demote [command options] [arguments...]

OPTIONS:
   --username value  the username to create/delete/etc
   --help, -h        show help (default: false)
```

Example:

```bash
gotosocial admin account demote --username some_username
```

### gotosocial admin account disable

This command can be used to disable an account: prevent it from signing in or doing anything, without deleting data.

`gotosocial admin account disable --help`:

```text
NAME:
   gotosocial admin account disable - prevent an account from signing in or posting etc, but don't delete anything

USAGE:
   gotosocial admin account disable [command options] [arguments...]

OPTIONS:
   --username value  the username to create/delete/etc
   --help, -h        show help (default: false)
```

Example:

```bash
gotosocial admin account disable --username some_username
```

### gotosocial admin account suspend

This command can be used to completely remove an account's media/posts/etc and prevent it from logging in.

In other words, this 'deletes' the account (without actually removing the account entry, meaning the username cannot be used again).

`gotosocial admin account suspend --help`:

```text
NAME:
   gotosocial admin account suspend - completely remove an account and all of its posts, media, etc

USAGE:
   gotosocial admin account suspend [command options] [arguments...]

OPTIONS:
   --username value  the username to create/delete/etc
   --help, -h        show help (default: false)
```

Example:

```bash
gotosocial admin account suspend --username some_username
```

### gotosocial admin account password

This command can be used to set a new password on the given account.

`gotosocial admin account password --help`:

```text
NAME:
   gotosocial admin account password - set a new password for the given account

USAGE:
   gotosocial admin account password [command options] [arguments...]

OPTIONS:
   --username value  the username to create/delete/etc
   --password value  the password to set for this account
   --help, -h        show help (default: false)
```

Example:

```bash
gotosocial admin account password --username some_username --pasword some_really_good_password
```
