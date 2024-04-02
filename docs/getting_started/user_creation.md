# Creating users

Regardless of the installation method, you'll need to create some users. GoToSocial currently doesn't have a way for users to be created through the web UI, or for people to sign-up through the web UI.

In the meantime, you can create a user using the CLI:

```sh
./gotosocial --config-path /path/to/config.yaml \
    admin account create \
    --username some_username \
    --email some_email@whatever.org \
    --password 'SOME_PASSWORD'
```

In the above command, replace `some_username` with your desired username, `some_email@whatever.org` with the email address you want to associate with your account, and `SOME_PASSWORD` with a secure password.

If you want your user to have admin rights, you can promote them using a similar command:

```sh
./gotosocial --config-path /path/to/config.yaml \
    admin account promote --username some_username
```

Replace `some_username` with the username of the account you just created.

!!! warning "Promotion requires server restart"
    
    Due to the way caching works in GoToSocial, some admin CLI commands require a server restart after running the command in order for the changes to "take".
    
    For example, after promoting a user to admin, you will need to restart your GoToSocial server so that the new values can be loaded from the database.

!!! tip
    
    Take a look at the other available CLI commands [here](../admin/cli.md).

## Containers

When running GoToSocial from a container, you'll need to execute the above command in the container instead. How to do this varies based on your container runtime, but for Docker it should look like:

```sh
docker exec -it CONTAINER_NAME_OR_ID \
    /gotosocial/gotosocial \
    admin account create \
    --username some_username \
    --email someone@example.org \
    --password 'some_very_good_password'
```

If you followed our Docker guide, the container name will be `gotosocial`. Both the name and the ID can be retrieved through `docker ps`.
