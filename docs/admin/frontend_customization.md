# Frontend Customization

Starting with v0.7.0, gotosocial embeds all frontend assets within the executable.
To override these, you can use directories on the filesystem, whose path is defined by [configuration variables](../configuration/web.md).
gotosocial attempts to read files from these directories first before using the embedded assets.

## Working with embedded assets using the CLI

The `gotosocial embedded` subcommand is provided to manage the embedded assets:

```
Usage:
  gotosocial embedded [command]

Available Commands:
  extract     write content of an embedded asset to a file on disk, creating a backup if the target already existed.
  list        generate list of assets embedded in this executable
  view        print content of an embedded asset
```

### Customizing the frontend

> NOTE:
> When running outside of docker, make sure to call the following commands from the user that is normally running gotosocial.

> NOTE:
> When running within docker, you should mount the `web-asset-base-dir` and `web-template-base-dir` as volumes, so customizations are persisted (The example docker-compose.yaml file already does this).
> In the following examples, treat `gotosocial` as an alias to `docker run <containername> gotosocial`.

1. Find the asset you want to modify.
    ```
    gotosocial embedded list
    ```

2. Extract the desired asset to the location where gotosocial is going to use it.
    For example, to customize the landing page of your instance, call
    ```
    gotosocial embedded extract template/index.tmpl
    ```
    With default configuration, you should have a file `web/templates/index.tmpl` next to your gotosocial executable.

3. Customize the template (templates are written in plain HTML + [go template syntax](https://pkg.go.dev/html/template)) and restart gotosocial to apply the changes.

## A word of caution

It's important to note that assets will change between gotosocial versions, and especially customized templates can become incompatible with the updated version.
We treat the API surface used in templates as private, but we try to point out major breaking changes on a best effort basis in the release notes.

### Migrating your customizations to a new release

When updating you should compare your customizations with files that ship with the new version.
For this, you can use something like the following bash snippet:

```sh
OLD_GTS="$(which gotosocial)"
NEW_GTS="/tmp/gotosocial_new"
NEW_EMBEDDED="/tmp/$($NEW_GTS --version)"
OLD_EMBEDDED="/tmp/$($OLD_GTS --version)"
CUSTOMDIR="$(dirname $(which gotosocial))/web"

CUSTOMIZED_FILES=$(find "$CUSTOMDIR" -type f)
for f in $CUSTOMIZED_FILES; do
    f=${f#"$CUSTOMDIR/"} # strip leading tree parts
    "$OLD_GTS" embedded extract --target-base-dir "$OLD_EMBEDDED" "$f"
    "$NEW_GTS" embedded extract --target-base-dir "$NEW_EMBEDDED" "$f" > /dev/null
    cmp "$OLD_EMBEDDED/$f" "$NEW_EMBEDDED/$f"
    if [[ $? -eq 0 ]]; then 
        echo "  has not changed"
    else
        echo "  differs in new version:"
        # remove -p to directly apply changes
        git merge-file -p -L customized -L old -L new "$CUSTOMDIR/$f" "$OLD_EMBEDDED/$f" "$NEW_EMBEDDED/$f"
    fi
done
```
