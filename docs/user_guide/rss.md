# RSS

RSS stands for [Really Simple Syndication](https://en.wikipedia.org/wiki/RSS). It's a very well established standard for sharing content on the web. You might recognize the jolly orange RSS logo from your favorite news websites and blogs:

![The orange RSS icon](../public/rss.svg)

If you like, you can configure your GoToSocial account to expose an RSS feed of your posts to the web. This allows people to get regular updates about your posts even when they don't have a Fediverse account. This is great when you're using GoToSocial to create longer-form, blog style posts, and you want anyone to be able to read them easily.

The RSS feed for GoToSocial profiles is turned off by default. You can enable it via the [User Settings](./settings.md) at `https://[your-instance-domain]/settings`.

When enabled, the RSS feed for your account will be available at `https://[your-instance-domain]/@[your_username]/feed.rss`. If you use an RSS reader, you can point it at this address to check that RSS is working.

## Which posts are shared via RSS?

Only your latest 20 Public posts are shared via RSS by default. Replies and reblogs/boosts are not included. Unlisted posts are not included. In other words, the only posts visible via RSS will be the same ones that are visible when you open your profile in a browser.

If you want to see more posts, you can provide our standard set of timeline paging parameters ([as per our swagger documentation](https://docs.gotosocial.org/en/latest/api/swagger)) to see beyond the first page.

You can also access Atom and JSON feeds from this same endpoint, but providing the appropriate request content-type header. i.e. `application/atom+xml` for an Atom feed, or `application/feed+json` for a JSON feed.
