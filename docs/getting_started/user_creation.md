# Creating users

Regardless of the installation method, you'll need to create some users. GoToSocial currently doesn't have a way for users to be created through the web UI, or for people to sign-up through the web UI.

Using the CLI, you can create a user:

```sh
$ gotosocial --config-path /path/to/config.yaml \
    admin account create \
    --username some_username \
    --email some_email@whatever.org \
    --password 'SOME_PASSWORD'
```

In the above command, replace `some_username` with your desired username, `some_email@whatever.org` with the email address you want to associate with your account, and `SOME_PASSWORD` with a secure password.

If you want your user to have admin rights, you can promote them using a similar command:

```sh
$ gotosocial --config-path /path/to/config.yaml \
    admin account promote --username some_username
```

Replace `some_username` with the username of the account you just created.

!!! info
    When running these commands, you'll get a bit of output like the following:

    ```text
    time=XXXX level=info msg=connected to SQLITE database
    time=XXXX level=info msg=there are no new migrations to run func=doMigration
    time=XXXX level=info msg=closing db connection
    ```

    This is normal and indicates that the commands ran as expected.

## Containers

When running GoToSocial from a container, you'll need to execute the above command in the conatiner instead. How to do this varies based on your container runtime, but for Docker it should look like:

```sh
$ docker exec -it CONTAINER_NAME_OR_ID \
    /gotosocial/gotosocial \
    admin account create \
    --username some_username \
    --email someone@example.org \
    --password 'some_very_good_password'
```

If you followed our Docker guide, the container name will be `gotosocial`. Both the name and the ID can be retrieved through `docker ps`.
