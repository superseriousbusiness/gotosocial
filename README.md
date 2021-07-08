# GoToSocial

![patrons](https://img.shields.io/liberapay/patrons/dumpsterqueer.svg?logo=liberapay) ![receives](https://img.shields.io/liberapay/receives/dumpsterqueer.svg?logo=liberapay)

Federated social media software.

![Sloth logo made by Freepik from www.flaticon.com](./web/assets/sloth.png)

GoToSocial is a Fediverse server project, written in Golang. It provides a lightweight, customizable, and safety-focused alternative to existing projects such as [Mastodon](https://joinmastodon.org/), [Pleroma](https://pleroma.social/), [Friendica](https://friendica.net), [PixelFed](https://pixelfed.org/) etc.

One of the key differences between GoToSocial and those other projects is that GoToSocial doesn't include an integrated client front-end (ie., a webapp).

Instead, like the Matrix.org's [Synapse](https://github.com/matrix-org/synapse) project, it provides only a server implementation, some static pages, and a well-documented API. On top of this API, developers are free to build any front-end implementation or mobile application that they wish.

Because the server implementation is as generic and flexible/configurable as possible, GoToSocial provides the basis for many different types of social media experience, whether Tumblr-like, Facebook-like, or Twitter-like.

## Features

A grab-bag of things that are already included or will be included in the first release(s) of the project:

* Mastodon API compatible, which means full support for apps you already know and love like [Tusky](https://tusky.app/) and [Pinafore](https://pinafore.social/).
* Various federation modes:
  * 'Normal' federation
  * Allowlist-only federation
  * Zero federation.
* Granular post settings:
  * Local-only posts.
  * Rebloggable/boostable toggle.
  * 'Likeable' toggle.
  * 'Replyable' toggle.
* Easy customizability for admins, without messing around in the source code:
  * Adjustable post length.
  * Media upload size settings.
* Built-in, automatic LetsEncrypt support (no messing around with Nginx or Certbot).
* Good performance on lower-powered machines like Raspberry Pi, old laptops and tiny VPSes (the test VPS has 1gb of ram and 1 cpu core).
* Subscribeable and shareable allow/denylists for federation.
* Strict privacy enforcement for posts and strict blocking logic.
* HTTP signature authentication by default (equivalent to Mastodon's [Secure Mode](https://docs.joinmastodon.org/spec/security/#http) being always-on).
* No external dependencies apart from a database. Binary + static assets only.

### Wishlist

These cool things will be implemented if time allows (because we really want them):

* Groups and group posting!
* Reputation-based 'slow' federation.
* User-selectable custom templates for rendering public posts:
  * Twitter-style
  * Blogpost
  * Gallery
  * Etc.

## Status

Work began on the project around February 2021, and the project is still in prerelease.

At this point, GoToSocial is already deployable and very useable, and it federates cleanly with other Fediverse servers.

For a detailed view on what's implemented and what's not, and progress made towards a first v0.1.0 (beta) release, see [here](./PROGRESS.md).

## Getting Started

Proper documentation for running and maintaining GoToSocial will be forthcoming in the first release.

For now (if you want to run it pre-alpha, like a beast), check out the [quick and dirty getting started guide](./GETTINGSTARTED.md).

## Contact

For questions and comments, you can reach out to tobi on the Fediverse [here](https://ondergrond.org/@dumpsterqueer), mail [admin@gotosocial.org](mailto:admin@gotosocial.org), or [join our Matrix channel](https://matrix.to/#/!gotosocial:ondergrond.org).

For bugs and feature requests, please check to see if there's [already an issue](https://github.com/superseriousbusiness/gotosocial/issues), and if not, open one or use one of the above channels to make a request (if you don't have a Github account).

## Credits

The following libraries and frameworks are used by GoToSocial, with gratitude 💕

* [buckket/go-blurhash](https://github.com/buckket/go-blurhash); used for generating image blurhashes.
* [gin-gonic/gin](https://github.com/gin-gonic/gin); speedy router engine.
  * [gin-contrib/cors](https://github.com/gin-contrib/cors); Gin CORS middleware.
  * [gin-contrib/sessions](https://github.com/gin-contrib/sessions); Gin sessions middleware.
  * [gin-contrib/static](https://github.com/gin-contrib/static); Gin static page middleware.
* [go-fed/activity](https://github.com/go-fed/activity); Golang ActivityPub/ActivityStreams library.
* [go-fed/httpsig](https://github.com/go-fed/httpsig); secure HTTP signature library.
* [go-pg/pg](https://github.com/go-pg/pg); Postgres ORM library.
* [google/uuid](https://github.com/google/uuid); UUID generation.
* [gorilla/websocket](https://github.com/gorilla/websocket); Websocket connectivity.
* [h2non/filetype](https://github.com/h2non/filetype); filetype checking.
* [oklog/ulid](https://github.com/oklog/ulid); sequential, database-friendly ID generation.
* [sirupsen/logrus](https://github.com/sirupsen/logrus); logging.
* [stretchr/testify](https://github.com/stretchr/testify); test framework.
* [superseriousbusiness/exifremove](https://github.com/superseriousbusiness/exifremove) forked from [scottleedavis/go-exif-remove](https://github.com/scottleedavis/go-exif-remove); EXIF data removal.
* [superseriousbusiness/oauth2](https://github.com/superseriousbusiness/oauth2) forked from [go-oauth2/oauth2](https://github.com/go-oauth2/oauth2); oauth server framework and token handling.
* [urfave/cli](https://github.com/urfave/cli); command-line interface framework.
* [wagslane/go-password-validator](https://github.com/wagslane/go-password-validator); password strength validation.

## Sponsorship + Funding

Currently, this project is funded using Liberapay, to put bread on the table while work continues on it.

If you want to sponsor this project, you can do so [here](https://liberapay.com/dumpsterqueer/)! `<3`

### Image Attribution

Logo made by [Freepik](https://www.freepik.com) from [www.flaticon.com](https://www.flaticon.com/).

### License

GoToSocial is licensed under the [GNU AGPL v3](LICENSE).
