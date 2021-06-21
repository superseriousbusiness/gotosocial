# GoToSocial

![patrons](https://img.shields.io/liberapay/patrons/dumpsterqueer.svg?logo=liberapay) ![receives](https://img.shields.io/liberapay/receives/dumpsterqueer.svg?logo=liberapay)

Federated social media software.

![Sloth logo made by Freepik from www.flaticon.com](./assets/sloth.png)

GoToSocial is a Fediverse server project, written in Golang. It provides an alternative to existing projects such as [Mastodon](https://joinmastodon.org/), [Pleroma](https://pleroma.social/), [Friendica](https://friendica.net), [PixelFed](https://pixelfed.org/) etc.

One of the key differences between GoToSocial and those other projects is that GoToSocial doesn't include an integrated client front-end (ie., a webapp). Instead, like the Matrix.org's [Synapse](https://github.com/matrix-org/synapse) project, it provides only a server implementation, some static web pages for profiles and posts, and a well-documented API. On this API, developers are free to build any front-end implementation or mobile application that they wish.

Because the server implementation is as generic and flexible/configurable as possible, GoToSocial provides the basis for many different types of social media experience, whether Tumblr-like, Facebook-like, or Twitter-like.

## Features Wishlist

A grab-bag of things that are already included or will be included in the project if time allows:

* Various federation modes, including reputation-based 'slow' federation, 'normal' federation, allowlist-only federation, and zero federation.
* Local-only posting, and granular post settings including 'rebloggable/boostable', 'likeable', 'replyable'.
* Character limit for posts that's easy for admins to configure (no messing around in the source code).
* Groups and group posting!
* Built-in, automatic LetsEncrypt support (no messing around with Nginx or Certbot).
* Good performance on lower-powered machines like Raspberry Pi, old laptops, tiny VPSes (the test VPS has 1gb of ram and 1 cpu core).
* Subscribeable and shareable allowlists/denylists for federation.

## Implementation Status

Things are moving on the project! As of June 2021 you can now:

* Build and deploy GoToSocial as a binary, with automatic LetsEncrypt certificate support built-in.
* Connect to the running instance via Tusky or Pinafore, using email address and password (stored encrypted).
* Post/delete posts.
* Reply/delete replies.
* Fave/unfave posts.
* Post images and gifs.
* Boost stuff/unboost stuff.
* Set your profile info (including header and avatar).
* Follow people/unfollow people.
* Accept follow requests from people.
* Post followers only/direct/public/unlocked.
* Customize posts with further flags: federated (y/n), replyable (y/n), likeable (y/n), boostable (y/n) -- not supported through Pinafore/Tusky yet.
* Get notifications for mentions/replies/likes/boosts.
* View local timeline.
* View and scroll home timeline (with ~10ms latency hell yeah).
* Stream new posts, notifications and deletes through a websockets connection via Pinafore.
* Federation support and interoperability with Mastodon and others.

In other words, a deployed GoToSocial instance is already pretty useable!

For a detailed view on progress made towards a v0.1.0 (beta) release, see [here](./PROGRESS.md).

## Contact

For questions and comments, you can reach out to tobi on the Fediverse [here](https://ondergrond.org/@dumpsterqueer) or mail admin@gotosocial.org.

## Sponsorship

Currently, this project is funded using Liberapay, to put bread on the table while work continues on it.

If you want to sponsor this project, you can do so [here](https://liberapay.com/dumpsterqueer/)! `<3`

### Image Attribution

Logo made by [Freepik](https://www.freepik.com) from [www.flaticon.com](https://www.flaticon.com/).
