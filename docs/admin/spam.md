# Spam Filtering

To make life a bit easier for admins trying to combat spam messages from open signup instances, GoToSocial has an experimental spam filter option.

If you or your users are being barraged by spam, try setting the option `instance-federation-spam-filter` to true in your config.yaml. You can read more about the heuristics used in the [instance config page](../configuration/instance.md).

Messages that are considered to be spam will not be stored on your instance, and will not generate notifications.

!!! warning
    Spam filters are necessarily imperfect tools, since they will likely catch at least a few legitimate messages in the filter, or indeed fail to catch some messages that *are* spam.
    
    Enabling `instance-federation-spam-filter` should be viewed as a "battening down the hatches" option for when the fediverse is facing a spam wave. Under normal circumstances, you will likely want to leave it turned off to avoid filtering out legitimate messages by accident.

!!! tip
    If you want to check what's being caught by the spam filter (if anything), grep your logs for the phrase "looked like spam".
    
    If you're [running GoToSocial as a systemd service](../getting_started/installation/metal.md#optional-enable-the-systemd-service), you can do this with the command:
    
    ```bash
    journalctl -u gotosocial --no-pager | grep 'looked like spam'
    ```
    
    If you see no output, that means no spam has been caught in the filter. Otherwise, you will see one or more log lines with links to statuses that have been filtered and dropped.
