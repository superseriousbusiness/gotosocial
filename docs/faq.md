# Frequently Asked Questions

## Where's the user interface?

GoToSocial is just a bare server for the most part and is designed to be used thru external applications. [Semaphore](https://semaphore.social/) in the browser, [Tusky](https://tusky.app/) for Android and [Feditext](https://github.com/feditext/feditext) for iOS, iPadOS and macOS are the best-supported. Anything that supports the Mastodon API should work, other than the features GoToSocial doesn't yet have. Permalinks and profile pages are served directly through GoToSocial as well as the settings panel, but most interaction goes through the apps.

## Why aren't my posts showing up on my profile page?

Unlike Mastodon, the default post visibility is Unlisted. If you want something to be visible on your profile page, the post must have Public visibility.

## Why aren't my posts showing up on other servers?

First check the visibility as noted above. TODO: explain how to debug common federation issues

## Why am I getting frequent http 429 error responses?

GoToSocial is configured to use per-IP [rate limiting](./api/ratelimiting.md) by default, but in certain situations it can't accurately identify the remote IP and will treat all connections as coming from the same place. In those cases, the rate limiting needs to be disabled or reconfigured.

## Why am I getting frequent HTTP 503 error responses?

Code 503 is returned to callers when your instance is under heavy load and requests are being throttled. This behavior can be tuned as desired, or turned off entirely, see [here](./api/throttling.md).

## I keep getting a 400 Bad Request error, and I have done everything suggested by the error message. What should I do?

Verify that the `host` configuration matches the domain that GoToSocial is served from (the domain that users use to acces the server).

## I keep seeing 'dial within blocked / reserved IP range' in my server logs, and I can't connect to some instances from my instance, what do I do?

The IP address of the remote instance may be in one of the blocked "special use" IP ranges hardcoded into GoToSocial for security reasons. If you need to, you can override this in your configuration file. Have a look at the [http client docs](./configuration/httpclient.md) for this, and please read the warnings there carefully! If you add an explicit allow, you will have to restart your GoToSocial instance to make the config take effect.

## My instance is deployed and I'm logged in to a client but my timelines are empty, what's up there?

To see posts, you have to start following people! Once you've followed a few people and they've posted or boosted things, you'll start seeing them in your timelines. Right now GoToSocial doesn't have a way of 'backfilling' posts -- that is, fetching previous posts from other instances -- so you'll only see new posts of people you follow. If you want to interact with an older post of theirs, you can copy the link to the post from their web profile, and paste it in to your client's search bar.

## How can I sign up for a server?

We introduced a sign-up flow in v0.16.0. The server you want to sign up to must have enabled registrations/sign-ups, as detailed [right here](./admin/signups.md).

## Why's it still in Beta?

Take a look at the [list of open bugs](https://github.com/superseriousbusiness/gotosocial/issues?q=is%3Aissue+is%3Aopen+label%3Abug) and the [roadmap](https://github.com/superseriousbusiness/gotosocial/blob/main/ROADMAP.md) for a more detailed rundown.
