# Frequently Asked Questions

- **Where's the user interface?** GoToSocial is just a bare server for the most part and is designed to be used thru external applications. [Pinafore](https://pinafore.social) and [Tusky](https://tusky.app/) are the best-supported, but anything that supports the Mastodon API should work, other than the features GoToSocial doesn't yet have. Permalinks and profile pages are served directly thru GoToSocial as well as the admin panel, but most interaction goes thru the apps.

- **Why aren't my posts showing up on my profile page?** Unlike Mastodon, the default post visibility is Unlisted. If you want something to be visible on your profile page, the post must have Public visibility.

- **Why aren't my posts showing up on other servers?** First check the visibility as noted above. TODO: explain how to debug common federation issues

- **Why am I getting frequent http 429 error responses?** GoToSocial is configured to use per-IP [rate limiting](./api/ratelimiting.md) by default, but in certain situations it can't accurately identify the remote IP and will treat all connections as coming from the same place. In those cases, the rate limiting needs to be disabled or reconfigured.

- **Why am I getting frequent http 503 error responses?** Code 503 is returned to callers when your instance is under heavy load and requests are being throttled. This behavior can be tuned as desired, or turned off entirely, see [here](./api/throttling.md).

- **My instance is deployed and I'm logged in to a client but my timelines are empty, what's up there?** To see posts, you have to start following people! Once you've followed a few people and they've posted or boosted things, you'll start seeing them in your timelines. Right now GoToSocial doesn't have a way of 'backfilling' posts -- that is, fetching previous posts from other instances -- so you'll only see new posts of people you follow. If you want to interact with an older post of theirs, you can copy the link to the post from their web profile, and paste it in to your client's search bar.

- **How can I sign up for a server?** Right now the only way to create an account is by the server's admin to run a command directly on the server. A web-based signup flow is in the roadmap but not implemented yet.

- **Why's it still in alpha?** Take a look at the [list of open bugs](https://github.com/superseriousbusiness/gotosocial/issues?q=is%3Aissue+is%3Aopen+label%3Abug) and the [roadmap](https://github.com/superseriousbusiness/gotosocial/blob/main/ROADMAP.md) for a more detailed rundown, but the main missing features at the time of this writing are:
    * reporting posts to admins
    * muting conversations
    * backfill of posts
    * web-based signup
    * profile metadata fields
    * lists of users
    * pinning posts to your profile
    * polls
    * scheduling posts
    * account migration
    * federated hashtag search
    * shared block lists across servers
