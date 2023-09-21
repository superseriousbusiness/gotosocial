# Admin Settings Panel

The GoToSocial admin settings panel uses the [admin API](https://docs.gotosocial.org/en/latest/api/swagger/#operations-tag-admin) to manage your instance. It's combined with the [user settings panel](../user_guide/settings.md) and uses the same OAuth mechanism as normal clients (with scope: admin).

## Setting admin account permissions and logging in

To use the admin settings panel, your account has to be promoted to admin:

```bash
./gotosocial --config-path ./config.yaml admin account promote --username YOUR_USERNAME
```

In order for the promotion to 'take', you may need to restart your instance after running the command.

After this, you can navigate to `https://[your-instance-name.org]/settings`, enter your domain in the login field, and login like you would with any other client. You should now see the admin settings.

## Moderation

Instance moderation settings.

### Reports

![List of reports for testing, one resolved and one open.](../assets/admin-settings-reports.png)

The reports section shows a list of reports, originating from your local users, or remote instances (shown anonymously as just the name of the instance, without specific username).

Clicking a report shows if it was resolved (with the reasoning if available), more information, and a list of reported toots if selected by the reporting user. You can also use this view to mark a report as resolved, and fill in a comment. Whatever comment you enter here will be visible to the user that created the report, if that user is from your instance.

Clicking on the username of the reported account opens that account in the 'Accounts' view, allowing you to perform moderation actions on it.

### Accounts

You can use this section to search for an account and perform moderation actions on it.

### Federation

![List of suspended instances, with a field to filter/add new blocks. Below is a link to the bulk import/export interface](../assets/admin-settings-federation.png)

In the federation section you can create, delete, and review explicit domain blocks and domain allows.

For more detail on federation settings, and specifically how domain allows and domain blocks work in combination, please see [the federation modes section](./federation_modes.md), and [the domain blocks section](./domain_blocks.md).

#### Domain Blocks

You can enter a domain to suspend in the search field, which will filter the list to show you if you already have a block for it.

Clicking 'suspend' gives you a form to add a public and/or private comment, and submit to add the block. Adding a suspension will suspend all the currently known accounts on the instance, and prevent any new interactions with any user on the blocked instance.

#### Domain Allows

The domain allows section works much like the domain blocks section, described above, only for explicit domain allows rather than domain blocks.

#### Bulk import/export

Through the link at the bottom of the Federation section (or going to `/settings/admin/federation/import-export`) you can do bulk import/export of blocklists and allowlists.

![List of domains included in an import, providing ways to select some or all of them, change their domains, and update the use of subdomains.](../assets/admin-settings-federation-import-export.png)

Upon importing a list, either through the input field or from a file, you can review the entries in the list before importing a subset. You'll also be warned for entries that use subdomains, providing an easy way to change them to the main domain.

## Administration

Instance administration settings.

### Actions

Run one-off administrative actions.

#### Media

You can use this section run a media action to clean up the remote media cache using the specified number of days. Media older than the given number of days will be removed from storage (s3 or local). Media removed in this way will be refetched again later if the media is required again. This action is functionally identical to the media cleanup that runs every night, automatically.

#### Keys

You can use this section to expire/invalidate public keys from the selected remote instance. The next time your instance receives a signed request using an expired key, it will attempt to fetch and store the public key again.

### Custom Emoji

Custom Emoji will be automatically fetched when included in remote toots, but to use them in your own posts they have to be enabled on your instance.

#### Local

![Local custom emoji section, showing an overview of custom emoji sorted by category. There are a lot of garfields.](../assets/admin-settings-emoji-local.png)

This section shows an overview of all the custom emoji enabled on your instance, sorted by their category. Clicking an emoji shows it's details, and provides options to change the category or image, or delete it completely. The shortcode cannot be updated here, you would have to upload it with the new shortcode yourself (and optionally delete the old one).

Below the overview you can upload your own custom emoji, after previewing how they look in a toot. PNG and (animated) GIF's are supported.

#### Remote

![Remote custom emoji section, showing a list of 3 emoji parsed from the entered toot, garfield, blobfoxbox and blobhajmlem. They can be selected, their shortcode can be tweaked, and they can be assigned to a category, before submitting as a copy or delete operation](../assets/admin-settings-emoji-remote.png)

Through the 'remote' section, you can look up a link to any remote toots (provided the instance isn't suspended). If they use any custom emoji they will be listed, providing an easy way to copy them to the local emoji (for use in your own toots), or disable them ( hiding them from toots).

**Note:** as the testrig server does not federate, this feature can't be used in development (500: Internal Server Error).

### Instance Settings

![Screenshot of the GoToSocial admin panel, showing the fields to change an instance's settings](../assets/admin-settings.png)

Here you can set various metadata for your instance, like the displayed name/title, thumbnail image, description (HTML accepted), and contact username and email.
