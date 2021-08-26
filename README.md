# GoToSocial

![patrons](https://img.shields.io/liberapay/patrons/dumpsterqueer.svg?logo=liberapay) ![receives](https://img.shields.io/liberapay/receives/dumpsterqueer.svg?logo=liberapay)

GoToSocial is an [ActivityPub](https://activitypub.rocks/) social network server, written in Golang.

<p align="middle">
  <img src="./docs/assets/sloth.png" width="300"/>
</p>

GoToSocial provides a lightweight, customizable, and safety-focused entryway into the [Fediverse](https://en.wikipedia.org/wiki/Fediverse), and is comparable to (but distinct from) existing projects such as [Mastodon](https://joinmastodon.org/), [Pleroma](https://pleroma.social/), [Friendica](https://friendica.net), and [PixelFed](https://pixelfed.org/).

With GoToSocial, you can keep in touch with your friends, post, read, and share images and articles, without being tracked or advertised to.

Documentation is at [docs.gotosocial.org](https://docs.gotosocial.org). You can skip straight to the API documentation [here](https://docs.gotosocial.org/en/latest/api/swagger/).

## Features

### Federation

Because GoToSocial uses the [ActivityPub](https://activitypub.rocks/) protocol, you can hang out not just with people on your home server, but with people all over the [Fediverse](https://en.wikipedia.org/wiki/Fediverse), seamlessly.

### Mastodon API compatible

Full support for modern, elegant apps like [Tusky](https://tusky.app/) and [Pinafore](https://pinafore.social/).

Tusky                                                        |  Pinafore
:-----------------------------------------------------------:|:------------------------------------------------------------------:
![An image of GoToSocial in Tusky](./docs/assets/tusky.png)  | ![An image of GoToSocial in Pinafore](./docs/assets/pinafore.png)

### Granular post settings

You should be able to choose how your posts can be interacted with:

* Local-only posts.
* Rebloggable/boostable toggle.
* 'Likeable' toggle.
* 'Replyable' toggle.

### Easy customizability for admins

* Adjustable post length.
* Media upload size settings.

### LetsEncrypt

 Built-in, automatic support for secure HTTPS with [LetsEncrypt](https://letsencrypt.org/).

### Light footprint and good performance

Plays nice with lower-powered machines like Raspberry Pi, old laptops and tiny VPSes.

### Easy to deploy

No external dependencies apart from a database. Just download the binary + assets (or Docker container), and run.

### HTTP signature authentication

Protect your data.

### User Safety

Strict privacy enforcement for posts and strict blocking logic.

### Subscribeable and shareable allow/denylists for federation

Import and export allowlists and denylists. Subscribe to community-created blocklists (think Adblocker, but for federation!).

### Various federation modes

* 'Normal' federation; discover new servers.
* Allowlist-only federation; choose which servers you talk to.
* Zero federation; keep your server private.

### Wishlist

These cool things will be implemented if time allows (because we really want them):

* **Groups** and group posting!
* Reputation-based 'slow' federation.
* Community decision making for federation and moderation actions.
* User-selectable custom templates for rendering public posts:
  * Twitter-style
  * Blogpost
  * Gallery
  * Etc.

## Design Ethos

One of the key differences between GoToSocial and other federated server projects is that GoToSocial doesn't include an integrated client front-end (ie., a webapp).

Instead, like Matrix.org's [Synapse](https://github.com/matrix-org/synapse) project, it provides only a server implementation, some static pages, and a well-documented API. On top of this API, developers are free to build any front-end implementation or mobile application that they wish.

Because the server implementation is as generic and flexible/configurable as possible, GoToSocial provides the basis for many different types of social media experience, whether Tumblr-like, Facebook-like, or Twitter-like.

## Status

Work began on the project around February 2021, and the project is still in prerelease.

At this point, GoToSocial is already deployable and very useable, and it federates cleanly with most other Fediverse servers (not yet all).

For a detailed view on what's implemented and what's not, and progress made towards a first v0.1.0 (beta) release, see [here](./PROGRESS.md).

## Getting Started

Proper documentation for running and maintaining GoToSocial will be forthcoming in the first release.

For now (if you want to run it pre-alpha, like a beast), check out the [quick and dirty getting started guide](https://docs.gotosocial.org/en/latest/installation_guide/quick_and_dirty/).

## Contributing

You wanna contribute to GtS? Great! ❤️❤️❤️ Check out the issues page to see if there's anything you wanna jump in on, and read the [CONTRIBUTING.md](./CONTRIBUTING.md) file for guidelines and setting up your dev environment.

## Contact

For questions and comments, you can [join our Matrix channel](https://matrix.to/#/#gotosocial:superseriousbusiness.org) at `#gotosocial:superseriousbusiness.org`. This is the quickest way to reach the devs. You can also mail [admin@gotosocial.org](mailto:admin@gotosocial.org).

For bugs and feature requests, please check to see if there's [already an issue](https://github.com/superseriousbusiness/gotosocial/issues), and if not, open one or use one of the above channels to make a request (if you don't have a Github account).

## Credits

The following libraries and frameworks are used by GoToSocial, with gratitude 💕

* [buckket/go-blurhash](https://github.com/buckket/go-blurhash); used for generating image blurhashes. [GPL-3.0 License](https://spdx.org/licenses/GPL-3.0-only.html).
* [coreos/go-oidc](https://github.com/coreos/go-oidc); OIDC client library. [Apache-2.0 License](https://spdx.org/licenses/Apache-2.0.html).
* [gin-gonic/gin](https://github.com/gin-gonic/gin); speedy router engine. [MIT License](https://spdx.org/licenses/MIT.html).
  * [gin-contrib/cors](https://github.com/gin-contrib/cors); Gin CORS middleware. [MIT License](https://spdx.org/licenses/MIT.html).
  * [gin-contrib/sessions](https://github.com/gin-contrib/sessions); Gin sessions middleware. [MIT License](https://spdx.org/licenses/MIT.html)
  * [gin-contrib/static](https://github.com/gin-contrib/static); Gin static page middleware. [MIT License](https://spdx.org/licenses/MIT.html)
* [go-fed/activity](https://github.com/go-fed/activity); Golang ActivityPub/ActivityStreams library. [BSD-3-Clause License](https://spdx.org/licenses/BSD-3-Clause.html).
* [go-fed/httpsig](https://github.com/go-fed/httpsig); secure HTTP signature library. [BSD-3-Clause License](https://spdx.org/licenses/BSD-3-Clause.html).
* [google/uuid](https://github.com/google/uuid); UUID generation. [BSD-3-Clause License](https://spdx.org/licenses/BSD-3-Clause.html)
* [gorilla/websocket](https://github.com/gorilla/websocket); Websocket connectivity. [BSD-2-Clause License](https://spdx.org/licenses/BSD-2-Clause.html).
* [h2non/filetype](https://github.com/h2non/filetype); filetype checking. [MIT License](https://spdx.org/licenses/MIT.html).
* [jackc/pgx](https://github.com/jackc/pgx); Postgres driver. [MIT License](https://spdx.org/licenses/MIT.html).
* [microcosm-cc/bluemonday](https://github.com/microcosm-cc/bluemonday); HTML user-input sanitization. [BSD-3-Clause License](https://spdx.org/licenses/BSD-3-Clause.html).
* [mvdan/xurls](https://github.com/mvdan/xurls); URL parsing regular expressions. [BSD-3-Clause License](https://spdx.org/licenses/BSD-3-Clause.html).
* [nfnt/resize](https://github.com/nfnt/resize); convenient image resizing. [ISC License](https://spdx.org/licenses/ISC.html).
* [oklog/ulid](https://github.com/oklog/ulid); sequential, database-friendly ID generation. [Apache-2.0 License](https://spdx.org/licenses/Apache-2.0.html).
* [ReneKroon/ttlcache](https://github.com/ReneKroon/ttlcache); in-memory caching. [MIT License](https://spdx.org/licenses/MIT.html).
* [russross/blackfriday](https://github.com/russross/blackfriday); markdown parsing for statuses. [Simplified BSD License](https://spdx.org/licenses/BSD-2-Clause.html).
* [sirupsen/logrus](https://github.com/sirupsen/logrus); logging. [MIT License](https://spdx.org/licenses/MIT.html).
* [stretchr/testify](https://github.com/stretchr/testify); test framework. [MIT License](https://spdx.org/licenses/MIT.html).
* [superseriousbusiness/exifremove](https://github.com/superseriousbusiness/exifremove) forked from [scottleedavis/go-exif-remove](https://github.com/scottleedavis/go-exif-remove); EXIF data removal. [MIT License](https://spdx.org/licenses/MIT.html).
* [superseriousbusiness/oauth2](https://github.com/superseriousbusiness/oauth2) forked from [go-oauth2/oauth2](https://github.com/go-oauth2/oauth2); oauth server framework and token handling. [MIT License](https://spdx.org/licenses/MIT.html).
* [go-swagger/go-swagger](https://github.com/go-swagger/go-swagger); Swagger OpenAPI spec generation. [Apache-2.0 License](https://spdx.org/licenses/Apache-2.0.html).
* [tdewolff/minify](https://github.com/tdewolff/minify); HTML minification. [MIT License](https://spdx.org/licenses/MIT.html).
* [uptrace/bun](https://github.com/uptrace/bun); database ORM. [BSD-2-Clause License](https://spdx.org/licenses/BSD-2-Clause.html).
* [urfave/cli](https://github.com/urfave/cli); command-line interface framework. [MIT License](https://spdx.org/licenses/MIT.html).
* [wagslane/go-password-validator](https://github.com/wagslane/go-password-validator); password strength validation. [MIT License](https://spdx.org/licenses/MIT.html).
* [modernc.org/sqlite](sqlite); cgo-free port of SQLite. [Other License](https://gitlab.com/cznic/sqlite/-/blob/master/LICENSE).
  * [modernc.org/ccgo](ccgo); c99 AST -> Go translater. [BSD-3-Clause License](https://spdx.org/licenses/BSD-3-Clause.html).
  * [modernc.org/libc](libc); C-runtime services. [BSD-3-Clause License](https://spdx.org/licenses/BSD-3-Clause.html).

### Image Attribution

Sloth logo made by [Freepik](https://www.freepik.com) from [www.flaticon.com](https://www.flaticon.com/).

## Sponsorship + Funding

Currently, this project is funded using Liberapay, to put bread on the table while work continues on it.

If you want to sponsor this project, you can do so [here](https://liberapay.com/dumpsterqueer/)! `<3`

## License

GoToSocial is licensed under the [GNU AGPL v3 LICENSE](LICENSE).

Copyright (C) 2021 the GoToSocial Authors.
