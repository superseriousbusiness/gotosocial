# GoToSocial <!-- omit in toc -->

![patrons](https://img.shields.io/liberapay/patrons/dumpsterqueer.svg?logo=liberapay) ![receives](https://img.shields.io/liberapay/receives/dumpsterqueer.svg?logo=liberapay)

GoToSocial is an [ActivityPub](https://activitypub.rocks/) social network server, written in Golang.

<p align="middle">
  <img src="./docs/assets/sloth.png" width="300"/>
</p>

GoToSocial provides a lightweight, customizable, and safety-focused entryway into the [Fediverse](https://en.wikipedia.org/wiki/Fediverse), and is comparable to (but distinct from) existing projects such as [Mastodon](https://joinmastodon.org/), [Pleroma](https://pleroma.social/), [Friendica](https://friendica.net), and [PixelFed](https://pixelfed.org/).

With GoToSocial, you can keep in touch with your friends, post, read, and share images and articles. All without being tracked or advertised to!

Documentation is at [docs.gotosocial.org](https://docs.gotosocial.org). You can skip straight to the API documentation [here](https://docs.gotosocial.org/en/latest/api/swagger/).

## Table of Contents <!-- omit in toc -->

- [Features](#features)
  - [Federation](#federation)
  - [Mastodon API compatible](#mastodon-api-compatible)
  - [Granular post settings](#granular-post-settings)
  - [Easy customizability for admins](#easy-customizability-for-admins)
  - [Easy to run](#easy-to-run)
  - [Safety + security features](#safety--security-features)
  - [Various federation modes](#various-federation-modes)
  - [OIDC integration](#oidc-integration)
  - [Wishlist](#wishlist)
- [Design Ethos](#design-ethos)
- [Status](#status)
- [Getting Started](#getting-started)
- [Contributing](#contributing)
- [Contact](#contact)
- [Credits](#credits)
  - [Image Attribution](#image-attribution)
- [Sponsorship + Funding](#sponsorship--funding)
- [License](#license)

## Features

### Federation

Because GoToSocial uses [ActivityPub](https://activitypub.rocks/), you can hang out not just with people on your home server, but with people all over the [Fediverse](https://en.wikipedia.org/wiki/Fediverse), seamlessly.

Federation means that your home server is part of a network of servers all over the world that all communicate using the same protocol. Your data is no longer centralized on the servers of one gigantic corporation, but resides on your own server and is shared -- as you see fit -- across a resilient web of other home servers.

### Mastodon API compatible

The Mastodon API has become the de-facto standard for client communication with federated servers, so GoToSocial has implemented and extended the API with custom functionality.

This means: full support for modern, beautiful apps like [Tusky](https://tusky.app/) and [Pinafore](https://pinafore.social/).

Tusky                                                        |  Pinafore
:-----------------------------------------------------------:|:------------------------------------------------------------------:
![An image of GoToSocial in Tusky](./docs/assets/tusky.png)  | ![An image of GoToSocial in Pinafore](./docs/assets/pinafore.png)

If you're used to using Mastodon, you'll find using GoToSocial a breeze.

### Granular post settings

It's important that when you post something, you can choose who sees it.

GoToSocial offers public/unlisted/friends-only/mutuals-only/and direct posts (slide in DMs!).

It also allows you to customize how people interact with your posts:

- Local-only posts.
- Rebloggable/boostable toggle.
- 'Likeable' toggle.
- 'Replyable' toggle.

### Easy customizability for admins

Lots of [config options](./example/config.yaml), including:

- Adjustable post length.
- Media upload size settings.

### Easy to run

No external dependencies apart from a database (or just use SQLite!). Simply download the binary + assets (or Docker container), and run.

Plays nice with lower-powered machines like Raspberry Pi, old laptops and tiny VPSes.

### Safety + security features

- Built-in, automatic support for secure HTTPS with [LetsEncrypt](https://letsencrypt.org/).
- Strict privacy enforcement for posts and strict blocking logic.
- Import and export allowlists and denylists. Subscribe to community-created blocklists (think Adblocker, but for federation!).
- HTTP signature authentication: GoToSocial requires [HTTP Signatures](https://tools.ietf.org/id/draft-cavage-http-signatures-01.html) when sending and receiving messages, to ensure that your messages can't be tampered with and your identity can't be forged.

### Various federation modes

GoToSocial doesn't apply a one-size-fits-all approach to federation. Who your server federates with should be up to you.

- 'Normal' federation; discover new servers.
- Allowlist-only federation; choose which servers you talk to.
- Zero federation; keep your server private.

### OIDC integration

GoToSocial supports [OpenID Connect (OIDC)](https://openid.net/connect/) identity providers, meaning you can integrate it with existing user management services like [Auth0](https://auth0.com/), [Gitlab](https://docs.gitlab.com/ee/integration/openid_connect_provider.html), etc, or run your own and hook GtS up to that (we recommend [Dex](https://dexidp.io/)).

### Wishlist

These cool things will be implemented if time allows (because we really want them):

- **Groups** and group posting!
- Reputation-based 'slow' federation.
- Community decision making for federation and moderation actions.
- User-selectable custom templates for rendering public posts:
  - Twitter-style
  - Blogpost
  - Gallery
  - Etc.

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

- [buckket/go-blurhash](https://github.com/buckket/go-blurhash); used for generating image blurhashes. [GPL-3.0 License](https://spdx.org/licenses/GPL-3.0-only.html).
- [coreos/go-oidc](https://github.com/coreos/go-oidc); OIDC client library. [Apache-2.0 License](https://spdx.org/licenses/Apache-2.0.html).
- [gin-gonic/gin](https://github.com/gin-gonic/gin); speedy router engine. [MIT License](https://spdx.org/licenses/MIT.html).
  - [gin-contrib/cors](https://github.com/gin-contrib/cors); Gin CORS middleware. [MIT License](https://spdx.org/licenses/MIT.html).
  - [gin-contrib/sessions](https://github.com/gin-contrib/sessions); Gin sessions middleware. [MIT License](https://spdx.org/licenses/MIT.html)
  - [gin-contrib/static](https://github.com/gin-contrib/static); Gin static page middleware. [MIT License](https://spdx.org/licenses/MIT.html)
- [go-fed/activity](https://github.com/go-fed/activity); Golang ActivityPub/ActivityStreams library. [BSD-3-Clause License](https://spdx.org/licenses/BSD-3-Clause.html).
- [go-fed/httpsig](https://github.com/go-fed/httpsig); secure HTTP signature library. [BSD-3-Clause License](https://spdx.org/licenses/BSD-3-Clause.html).
- [google/uuid](https://github.com/google/uuid); UUID generation. [BSD-3-Clause License](https://spdx.org/licenses/BSD-3-Clause.html)
- [go-playground/validator](https://github.com/go-playground/validator); struct validation. [MIT License](https://spdx.org/licenses/MIT.html)
- [gorilla/websocket](https://github.com/gorilla/websocket); Websocket connectivity. [BSD-2-Clause License](https://spdx.org/licenses/BSD-2-Clause.html).
- [h2non/filetype](https://github.com/h2non/filetype); filetype checking. [MIT License](https://spdx.org/licenses/MIT.html).
- [jackc/pgx](https://github.com/jackc/pgx); Postgres driver. [MIT License](https://spdx.org/licenses/MIT.html).
- [microcosm-cc/bluemonday](https://github.com/microcosm-cc/bluemonday); HTML user-input sanitization. [BSD-3-Clause License](https://spdx.org/licenses/BSD-3-Clause.html).
- [mitchellh/mapstructure](https://github.com/mitchellh/mapstructure); Go interface => struct parsing. [MIT License](https://spdx.org/licenses/MIT.html).
- [modernc.org/sqlite](https://gitlab.com/cznic/sqlite); cgo-free port of SQLite. [Other License](https://gitlab.com/cznic/sqlite/-/blob/master/LICENSE).
  - [modernc.org/ccgo](https://gitlab.com/cznic/ccgo); c99 AST -> Go translater. [BSD-3-Clause License](https://spdx.org/licenses/BSD-3-Clause.html).
  - [modernc.org/libc](https://gitlab.com/cznic/libc); C-runtime services. [BSD-3-Clause License](https://spdx.org/licenses/BSD-3-Clause.html).
- [mvdan/xurls](https://github.com/mvdan/xurls); URL parsing regular expressions. [BSD-3-Clause License](https://spdx.org/licenses/BSD-3-Clause.html).
- [nfnt/resize](https://github.com/nfnt/resize); convenient image resizing. [ISC License](https://spdx.org/licenses/ISC.html).
- [oklog/ulid](https://github.com/oklog/ulid); sequential, database-friendly ID generation. [Apache-2.0 License](https://spdx.org/licenses/Apache-2.0.html).
- [ReneKroon/ttlcache](https://github.com/ReneKroon/ttlcache); in-memory caching. [MIT License](https://spdx.org/licenses/MIT.html).
- [russross/blackfriday](https://github.com/russross/blackfriday); markdown parsing for statuses. [Simplified BSD License](https://spdx.org/licenses/BSD-2-Clause.html).
- [sirupsen/logrus](https://github.com/sirupsen/logrus); logging. [MIT License](https://spdx.org/licenses/MIT.html).
- [stretchr/testify](https://github.com/stretchr/testify); test framework. [MIT License](https://spdx.org/licenses/MIT.html).
- [superseriousbusiness/exifremove](https://github.com/superseriousbusiness/exifremove) forked from [scottleedavis/go-exif-remove](https://github.com/scottleedavis/go-exif-remove); EXIF data removal. [MIT License](https://spdx.org/licenses/MIT.html).
- [superseriousbusiness/oauth2](https://github.com/superseriousbusiness/oauth2) forked from [go-oauth2/oauth2](https://github.com/go-oauth2/oauth2); oauth server framework and token handling. [MIT License](https://spdx.org/licenses/MIT.html).
- [go-swagger/go-swagger](https://github.com/go-swagger/go-swagger); Swagger OpenAPI spec generation. [Apache-2.0 License](https://spdx.org/licenses/Apache-2.0.html).
- [tdewolff/minify](https://github.com/tdewolff/minify); HTML minification. [MIT License](https://spdx.org/licenses/MIT.html).
- [uptrace/bun](https://github.com/uptrace/bun); database ORM. [BSD-2-Clause License](https://spdx.org/licenses/BSD-2-Clause.html).
- [urfave/cli](https://github.com/urfave/cli); command-line interface framework. [MIT License](https://spdx.org/licenses/MIT.html).
- [wagslane/go-password-validator](https://github.com/wagslane/go-password-validator); password strength validation. [MIT License](https://spdx.org/licenses/MIT.html).

### Image Attribution

Sloth logo made by [Freepik](https://www.freepik.com) from [www.flaticon.com](https://www.flaticon.com/).

## Sponsorship + Funding

Currently, this project is funded using Liberapay, to put bread on the table while work continues on it.

If you want to sponsor this project, you can do so [here](https://liberapay.com/dumpsterqueer/)! `<3`

## License

GoToSocial is free software, licensed under the [GNU AGPL v3 LICENSE](LICENSE).

Copyright (C) 2021 the GoToSocial Authors.
