# Settings

GoToSocial provides a settings interface where you can update your post and profile settings, add an avatar and header image, write a bio for your account, and so on.

You can access the Settings at `https://my-instance.example.com/settings` with your own GoToSocial instance. It uses the same OAuth mechanism as normal clients.

You will be prompted to log in with your email address and password after providing the instance url.

## Profile

![Screenshot of the profile section of the user settings interface, showing a preview of the avatar, header and display name, and providing form fields to change them](../assets/user-settings-profile-info.png)

In the profile section you can change your display name, avatar and header images. You can also choose to enable manually approving follow requests, and opt-in to providing a public RSS feed of your posts.

### Set Avatar / Header

To set an avatar or header image, click on the `Browse` button in the appropriate section, and use the file browser to select an image.

Currently, supported image formats are `gif`, `png`, `webp`, and `jpeg`/`jpg`.

A preview of the image as it will appear on your profile will be shown. If you're happy with your choices, click on the `Save profile info` button at the bottom of the page.

If you navigate to your profile and refresh the page, your new avatar / header will be shown. It might take a bit longer for the update to federate out to remote instances.

### Select Theme

GoToSocial provides themes for you to choose from for the web view of your profile, to change your profile's appearance and vibe.

To choose a theme, just select it from the profile settings page, and click/tap "Save profile info" at the bottom of the page. When you look at your profile in the web view (you may need to refresh the page), you'll see the new theme applied, and so will anyone else visiting your profile.

!!! tip "Adding more themes"
    Instance admins can add more themes by dropping css files into the `web/assets/themes` folder. See the [themes](../admin/themes.md) part of the admin docs for more information.

### Basic Information

#### Display Name

Your display name is a short handle shown alongside your username on your profile.

While your username cannot be changed once it's created, your display name can.

Your display name can also contain spaces, capital letters, emojis, and so on.

It's a great place to put a nickname or full name. For example, if your username is `@miranda`, your display name could be something like `Miranda Priestly`.

#### Bio

Your bio is a longer text that introduces your account and your self. Your bio is a good place to:

- Give an indication of the sort of things you post about.
- Mention your approximate age / location.
- Link to any of your other accounts or profiles elsewhere.
- Describe your boundaries and preferences when it comes to other people interacting with you.
- Link hashtags that you often use when you post.

The bio accepts either `plain` or `markdown` formatting. This is set by the default post format setting described in [Post Settings](#post-settings).

#### Profile Fields

Profile fields are a series of name/value pairs that will appear on your profile, and be federated to remote instances.

This is a perfect place to put things like:

- Links to your website(s)
- Links to crowdfunding / donation pages
- Your age
- Pronouns

Some examples:

- Alias : handler walter
- My Website : https://example.org
- Age : 99
- Pronouns : she/her
- My other account : @someone@somewhere.com

### Visibility and Privacy

#### Manually Approve Follow Requests (aka Lock Your Account)

This checkbox allows you to decide whether or not you want to manually review follow requests to your account.

When this is **not checked**, new follow requests are approved automatically without your intervention. This is useful for more public-facing accounts or cases where you don't really post anything sensitive or private.

When it is **checked**, you must manually approve new follow requests, and you can deny follow requests from accounts you don't want to follow you. This is useful for private accounts where you post personal things to followers only.

This option is often referred to on the fediverse as "locking" your account.

After ticking or unticking the checkbox, be sure to click on the `Save profile info` button at the bottom to save your new settings.

#### Enable RSS Feed of Public Posts

RSS feeds for users are disabled by default, but can be opted into with this checkbox. For more information see [RSS](./rss.md).

This feed only includes posts set as 'Public' (see [Privacy Settings](./posts.md#privacy-settings)).

!!! warning
    Exposing your RSS feed allows *anyone* to subscribe to updates on your Public posts anonymously, bypassing follows and follow requests.

#### Mark Account as Discoverable by Search Engines and Directories

This setting updates the 'discoverable' flag on your account.

Checking the discoverable box for your account does the following:

- Update robots meta tags for your account, allowing it to be indexed by search engines and appear in search engine results.
- Indicate to remote instances that your account may be included in public directories and indexes.

Turning on the discoverable flag may take a week or more to propagate; your account will not immediately appear in search engine results.

!!! tip
    Discoverable is set to false by default for new accounts, to avoid exposing them to crawlers. Setting it to true is useful for public-facing accounts where you actually *want* to be crawled.

!!! info
    The discoverable setting is about **discoverability of your account**, not searchability of your posts. It has nothing to do with indexing of your posts for search by Mastodon instances, or other federated instances that use full text search!

### Advanced

#### Custom CSS

If enabled on your instance by the instance administrator, custom CSS allows you to further customize the way your profile looks when visited through a browser.

When this setting is not enabled by the instance administrator, the text input box is read-only and custom CSS will not be applied.

See the [Custom CSS](./custom_css.md) page for some tips on writing custom CSS for your profile.

!!! tip
    Any custom CSS you add in this box will be applied *after* your selected theme, so you can pick a preset theme that you like and then make your own tweaks!

## Post Settings

![Screenshot of the user settings section, providing drop-down menu's to select default post settings, and form fields to change your password](../assets/user-settings-post-settings.png)

In the 'Settings' section, you can set various defaults for new posts.

The default post language setting allows you to indicate to other fediverse users which language your posts are usually written in. This is helpful for fediverse users who speak (for example) Korean, and would prefer to filter out posts written in other languages.

The default post privacy setting allows you to set the default privacy for new posts. This is useful when you generally prefer to post public or followers-only, but you don't want to have to remember to set the privacy every time you post. Remember, this is only the default: no matter what you set here, you can still set the privacy individually for new posts if desired. For more information on post privacy settings, see the [page on Posts](./posts.md).

The default post format setting allows you to set which text interpreter should be used when parsing your posts.

The plain (default) setting provides standard post formatting, similar to what many other fediverse servers use. This is great for general purpose posting: you can write short, twitter-style posts, or multi-paragraph essays, insert links, and mention other accounts using their username.

The markdown setting indicates that your posts should be parsed as Markdown, which is a markup language that gives you more options for customizing the layout and appearance of your posts. For more information on the differences between plain and markdown post formats, see the [posts page](posts.md).

When you are finished updating your post settings, remember to click the `Save post settings` button at the bottom of the section to save your changes.

## Password Change

You can use the Password Change section of the User Settings Panel to set a new password for your account.

For more information on the way GoToSocial manages passwords, please see the [Password management document](./password_management.md).

## Migration

In the migration section you can manage settings related to aliasing and/or migrating your account to another account.

!!! tip
    Depending on the software that a target account is hosted on, target account URIs for both aliasing and moves should look something like `https://mastodon.example.org/users/account_you_are_moving_to`. If you are unsure what format to use, check with the admin of the instance you are moving or aliasing to.

### Alias Account

You can use this section to create an alias from your GoToSocial account to other accounts elsewhere, indicating that you are also known as those accounts.

**Not implemented yet**: Alias information for accounts you enter here will be shown on the web view of your profile, but only if the target accounts are also aliased back to your account first. This is to prevent accounts from claiming to be aliased to other accounts that they don't actually control.

### Move Account

Using the move account settings, you can trigger the migration of your current account to the given target account URI.

In order for the move to be successful, the target account (the account you are moving to) must be aliased back to your current account (the account you are moving from). The target account must also be reachable from your current account, ie., not blocked by you, not suspended by your current instance, and not on a domain that is blocked by your current instance. The target account does not have to be on a GoToSocial instance.

GoToSocial uses an account move cooldown of 7 days. If either your current account or the target account have recently been involved in a move, you will not be able to trigger a move to the target account until seven days have passed.

Moving your account will send a message out from your current account, to your current followers, indicating that they should follow the target account instead. Depending on the server software used by your followers, they may then automatically send a follow (request) to the target account, and unfollow your current account.

Currently, **only your followers will be carried over to the new account**. Other things like your following list, statuses, media, bookmarks, faves, blocks, etc, will not be carried over.

Once your account has moved, the web view of your current (now old) account will show a notice that you have moved, and to where.

Your old statuses and media will still be visible on the web view of the account you've moved from, unless you delete them manually. If you prefer, you can ask the admin of the instance you've moved from to suspend/delete your account after the move has gone through.

If necessary, you can retry an account move using the same target account URI. This will send the move message out again.

!!! danger "Moving your account is an irreversible, permanent action!"
    
    From the moment you trigger an account move, you will have only basic read- and delete-level permissions on the account you've moved from.
    
    You will still be able to log in to your old account and see your own posts, faves, bookmarks, blocks, and lists.
    
    You will also be able to edit your profile, delete and/or unpin your own posts, and unboost, unfave, and unbookmark posts.
    
    However, you will not be able to take any action that involves creating something, such as writing, boosting, bookmarking, or faving a post, following someone, uploading media, creating a list, etc.
    
    Additionally, you will not be able to view any timelines (home, tag, public, list), or use the search functionality.

## Admins

If your account has been promoted to admin, this interface will also show sections related to admin actions, see [Admin Settings](../admin/settings.md).
