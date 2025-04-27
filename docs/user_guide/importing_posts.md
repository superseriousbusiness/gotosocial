# Importing posts from previous instances

As of v0.18.0, GoToSocial can import your archived posts from previous instances.

!!! tip
    This process creates a *copy* of your previous posts. ActivityPub as deployed in the Fediverse in early 2025 does not yet have a way to *move* posts, or make existing posts retrievable from a new location. If your previous instance goes away, so do the original posts — but you'll still have your imported copies.

## What you can import

- Your top-level posts
- Your replies to your own posts
- The original creation dates for those posts
- Media attached to those posts (photos, audio, video)
- Emojis used in those posts (if they're already available on your instance)

Importing your posts is quiet by design: imported posts don't get pushed to your followers on remote instances, aren't inserted into timelines on your instance, and don't generate notifications for your subscribers. This means you can import a large set of posts without annoying your followers. However, once posts are imported, you can boost them or share their URLs like any other post of yours.

## What you can't import

- Replies to other accounts
- Mentions of other accounts
- Boosts (of your posts or those from other accounts)
- Other accounts' likes or boosts of your posts
- Posts with polls
- Posts dated to or before the Unix epoch (Jan. 1, 1970)

The limitations on interactions with other accounts are an intentional politeness measure: allowing replies to other people in backdated posts could result in confusion ("hey, i remember this conversation, but not with this account"). Additionally, if you imported a bunch of replies to someone in a back and forth thread, and then boosted them, causing them to be visible to that person's instance, they could end up getting slammed with a lot of mention or pending mention notifications, years after the original conversation was over. For similar reasons, replaying likes or boosts would be unavoidably spammy.

## How to import your posts

The process currently requires third-party tools which use the GTS API. In the future, we may integrate this into GoToSocial itself: please follow [issue #2](https://codeberg.org/superseriousbusiness/gotosocial/issues/2) for updates.

[`slurp`](https://github.com/VyrCossont/slurp) (by GTS developer Vyr Cossont) can import [post archives from Mastodon](https://github.com/VyrCossont/slurp?tab=readme-ov-file#importing-a-mastodon-archive) as well as [from Pixelfed](https://github.com/VyrCossont/slurp?tab=readme-ov-file#importing-a-pixelfed-archive). Please consult `slurp`'s docs, [Mastodon's instructions for exporting your data](https://docs.joinmastodon.org/user/moving/#export), and ["Importing Pixelfed Posts to GoToSocial with Slurp" by Jeff Sikes](https://box464.com/posts/gotosocial-slurp/) for more details. You'll need to be familiar with command-line basics, and have Git and a Go compiler installed.

!!! warning
    If importing from Pixelfed, note that Pixelfed archives don't contain your photos, so your original instance and account must still work at the time of import.

## For developers

You can use GoToSocial's backdating feature through the `scheduled_at` parameter to `POST /v1/statuses/create`. If this date-time parameter is set and the date is in the *past*, the post will be treated as a backdated import, and the `scheduled_at` date will be used to set the post's creation date and ID. (GoToSocial uses [ULIDs](https://github.com/ulid/spec) for IDs, which may be sorted lexicographically to sort them by time.) Additionally, the post will not be pushed to followers or timelines, or generate notifications. The return type when creating a backdated post is a `status`, as when posting normally.

Since this process uses the GTS API, the original post doesn't have to be an ActivityPub activity, and could come from a blog, Cohost, Bluesky, Usenet, etc.
