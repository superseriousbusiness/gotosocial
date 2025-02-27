# Migration

GoToSocial supports account migration using the `Move` activity.

This allows you to move an account to your GoToSocial account, or to move your GoToSocial account to another account.

Moving is software agnostic, so you can migrate your account to or from any software that supports the `Move` activity, regardless of the exact software. Eg., you can move your GoToSocial account to a Mastodon account, move your Mastodon account to a GoToSocial account, move your GoToSocial account to or from Akkoma, Misskey, GoToSocial, etc.

!!! tip
    Depending on the software that a target account is hosted on, target account URIs for both aliasing and moves should look something like `https://mastodon.example.org/users/account_you_are_moving_to`. If you are unsure what format to use, check with the admin of the instance you are moving or aliasing to.

!!! warning
    GoToSocial uses an account move cooldown of 7 days to prevent excessive instance hopping (and potential block evasion).
    
    If either one of the accounts involved in a new move attempt have moved in the last seven days, GoToSocial will refuse to trigger a move to or from either account until seven days have passed since the previous move.

## Move your GoToSocial account to another account (move *from* GoToSocial)

Using the move account settings, you can trigger the migration of your GoToSocial account to the given target account URI.

In order for the move to be successful:

1. The target account (the account you are moving to) must be aliased back to your current account (the account you are moving from).
2. The target account must be reachable from your current account, ie., not blocked by you, not blocking you, not suspended, not on a domain that is blocked by your current instance.

Moving your account will send a message out from your current account, to your current followers, indicating that they should follow the target account instead. Depending on the server software used by your followers, they may then automatically send a follow (request) to the target account, and unfollow your current account.

Currently, **only your followers will be carried over to the new account**. Other things like your following list, media, bookmarks, faves, blocks, etc, will not be carried over. You may [import your posts](./importing_posts.md) as a separate operation.

Once your account has moved, the web view of your current (now old) account will show a notice that you have moved, and to where.

Your old statuses and media will still be visible on the web view of the account you've moved from, unless you delete them manually. If you prefer, you can ask the admin of the instance you've moved from to suspend/delete your account after the move has gone through.

If necessary, you can retry an account move using the same target account URI. This will send the move message out again. This is useful in cases where your followers may not have received the move message due to network issues or other temporary outage. 

!!! danger "Moving your account is an irreversible, permanent action!"
    
    From the moment you trigger an account move from GoToSocial, you will have only basic read- and delete-level permissions on the account you've moved from.
    
    You will still be able to log in to your old account and see your own posts, faves, bookmarks, blocks, and lists.
    
    You will also be able to edit your profile, delete and/or unpin your own posts, and unboost, unfave, and unbookmark posts.
    
    However, you will not be able to take any action that involves creating something, such as writing, boosting, bookmarking, or faving a post, following someone, uploading media, creating a list, etc.
    
    Additionally, you will not be able to view any timelines (home, tag, public, list), or use the search functionality.

## Move an account to your GoToSocial account (move *to* GoToSocial)

To successfully trigger the migration of another account to your GoToSocial account, you must first create an **alias** linking your GoToSocial account back to the account you want to move from, to indicate to the account you're moving from that you also own the GoToSocial account you want to move to.

To do this, you must first log in to the GoToSocial settings panel with your GoToSocial account. Eg., if your GoToSocial instance is at `https://example.org`, you should log in to the settings panel at `https://example.org/settings`.

From there, go to the "Migration" section, and look at the "Alias Account" subsection:

![The Alias Account subsection, showing a filled-in account alias.](../public/migration-aliasing.png)

In the first free account alias box, enter the URL of the account you wish to move **from**. This indicates that the account you wish to move from belongs to you, ie., you are "also known as" the account.

For example, if you are moving **from** the account "@dumpsterqueer" which is on the instance "ondergrond.org", you should enter the value "https://ondergrond.org/@dumpsterqueer" or "https://ondergrond.org/users/dumpsterqueer" as an account alias, as in the image above.

Once you have entered your alias, click the "Save account aliases" button. If all goes well, a tick will be shown on the button. If not, an error will be shown which should help you figure out what went wrong.

Once you have created the account alias from your GoToSocial account, pointing back to the account you wish to move **from**, you can use the settings panel on the other account's instance to trigger the move to your GoToSocial account.

On Mastodon, the "Account migration" settings section looks something like this:

![The Mastodon "Account migration" settings page.](../public/migration-mastodon.png)

If you were moving to a GoToSocial account from a Mastodon account, you would fill in the "Handle of the new account" field with the `@[username]@[domain]` value of your GoToSocial account. For example, if your GoToSocial account has username "@someone" and it's on the instance "example.org", you would enter `@someone@example.org` here.

Once you have triggered the move from your other account to your GoToSocial account, the only thing you have left to do is accept follow requests from your old account's followers on your new (GoToSocial) account.

!!! tip
    To save yourself some trouble, consider setting your GoToSocial account to not require approval for new follow requests, just before triggering the migration. Once the migration is complete, turn approval of follow requests back on. Otherwise, you will have to manually approve every migrated follower from your old account.

!!! tip
    After moving your account, you may wish to import your list of followed accounts from your previous account into your GoToSocial account. [See here](./settings.md#import) for details on how to do this via the settings panel.
