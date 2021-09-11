# GoToSocial <!-- omit in toc -->

![patrons](https://img.shields.io/liberapay/patrons/GoToSocial.svg?logo=liberapay) ![receives](https://img.shields.io/liberapay/receives/GoToSocial.svg?logo=liberapay)

GoToSocial is an [ActivityPub](https://activitypub.rocks/) social network server, written in Golang.

With GoToSocial, you can keep in touch with your friends, post, read, and share images and articles. All without being tracked or advertised to!

<p align="middle">
  <img src="./docs/assets/sloth.png" width="300"/>
</p>

Documentation is at [docs.gotosocial.org](https://docs.gotosocial.org). You can skip straight to the API documentation [here](https://docs.gotosocial.org/en/latest/api/swagger/).

## Table of Contents <!-- omit in toc -->

- [What is GoToSocial?](#what-is-gotosocial)
  - [Federation](#federation)
  - [History and Status](#history-and-status)
- [Features](#features)
  - [Mastodon API compatibility](#mastodon-api-compatibility)
  - [Granular post settings](#granular-post-settings)
  - [Customizability for admins](#customizability-for-admins)
  - [Easy to run](#easy-to-run)
  - [Safety + security features](#safety--security-features)
  - [Various federation modes](#various-federation-modes)
  - [OIDC integration](#oidc-integration)
  - [Backend-first design](#backend-first-design)
- [Wishlist](#wishlist)
- [Getting Started](#getting-started)
- [Contributing](#contributing)
- [Contact](#contact)
- [Credits](#credits)
  - [Libraries](#libraries)
  - [Image Attribution](#image-attribution)
  - [Developers](#developers)
  - [Special Thanks](#special-thanks)
- [Sponsorship + Funding](#sponsorship--funding)
- [License](#license)

## What is GoToSocial?

GoToSocial provides a lightweight, customizable, and safety-focused entryway into the [Fediverse](https://en.wikipedia.org/wiki/Fediverse), and is comparable to (but distinct from) existing projects such as [Mastodon](https://joinmastodon.org/), [Pleroma](https://pleroma.social/), [Friendica](https://friendica.net), and [PixelFed](https://pixelfed.org/).

If you've ever used something like Twitter or Tumblr (or even Myspace!) GoToSocial will probably feel familiar to you: You can follow people and have followers, you make posts which people can favourite and reply to and share, and you scroll through posts from people you follow using a timeline. You can write long posts or short posts, or just post images, it's up to you. You can also, of course, block people or otherwise limit interactions that you don't want by posting just to your friends.

**GoToSocial does NOT use algorithms or collect data about you to suggest content or 'improve your experience'**. The timeline is chronological: whatever you see at the top of your timeline is there because it's *just been posted*, not because it's been selected as interesting (or controversial) based on your personal profile.

GoToSocial is not designed for 'must-follow' influencers with tens of thousands of followers, and it's not designed to be addictive. Your timeline and your experience is shaped by who you follow and how you interact with people, not by metrics of engagement!

GoToSocial doesn't claim to be *better* than any other application, but it offers something that might be better *for you* in particular.

### Federation

Because GoToSocial uses [ActivityPub](https://activitypub.rocks/), you can hang out not just with people on your home server, but with people all over the [Fediverse](https://en.wikipedia.org/wiki/Fediverse), seamlessly.

Federation means that your home server is part of a network of servers all over the world that all communicate using the same protocol. Your data is no longer centralized on one company's servers, but resides on your own server and is shared -- as you see fit -- across a resilient web of servers run by other people.

This federated approach also means that you aren't beholden to arbitrary rules from some gigantic corporation potentially thousands of miles away. Your server has its own rules and culture; your fellow server residents are your neighbors; you will likely get to know your server admins and moderators, or be an admin yourself.

GoToSocial advocates for many small, weird, specialist servers where people can feel at home, rather than a few big and generic ones where one person's voice can get lost in the crowd.

### History and Status

This project sprang up in 2021 out of a dissatisfaction with the safety + privacy features of other Federated microblogging/social media applications, and a desire to implement something a little different.

It began as a solo project, and then picked up steam as more developers became interested and jumped on.

The project is still in prerelease, but is already deployable and very useable, and it federates cleanly with most other Fediverse servers (not yet all).

For a detailed view on what's implemented and what's not, and progress made towards a first v0.1.0 (beta) release, see [here](./PROGRESS.md).

## Features

### Mastodon API compatibility

The Mastodon API has become the de-facto standard for client communication with federated servers, so GoToSocial has implemented and extended the API with custom functionality.

In short this means full support for modern, beautiful apps like [Tusky](https://tusky.app/) and [Pinafore](https://pinafore.social/).

Tusky                                                        |  Pinafore
:-----------------------------------------------------------:|:------------------------------------------------------------------:
![An image of GoToSocial in Tusky](./docs/assets/tusky.png)  | ![An image of GoToSocial in Pinafore](./docs/assets/pinafore.png)

If you're used to using Mastodon with Tusky or Pinafore, you'll find using GoToSocial a breeze.

### Granular post settings

It's important that when you post something, you can choose who sees it.

GoToSocial offers public/unlisted/friends-only/mutuals-only/and direct posts (slide in DMs! -- with consent).

It also allows you to customize how people interact with your posts:

- Local-only posts.
- Rebloggable/boostable toggle.
- 'Likeable' toggle.
- 'Replyable' toggle.

### Customizability for admins

Lots of [config options](./example/config.yaml) for admins to play around with, including:

- Easily-adjustable post length.
- Media upload size settings.

### Easy to run

No external dependencies apart from a database (or just use SQLite!). Simply download the binary + assets (or Docker container), and run.

GoToSocial plays nice with lower-powered machines like Raspberry Pi, old laptops and tiny $5/month VPSes.

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

### Backend-first design

Unlike other federated server projects, GoToSocial doesn't include an integrated client front-end (ie., a webapp).

Instead, like Matrix.org's [Synapse](https://github.com/matrix-org/synapse) project, it provides a relatively generic backend server implementation, some beautiful static pages for profiles and posts, and a [well-documented API](https://docs.gotosocial.org/en/latest/api/swagger/).

On top of this API, web developers are encouraged to build any front-end implementation or mobile application that they wish, whether Tumblr-like, Facebook-like, Twitter-like, or something else entirely.

## Wishlist

These cool things will be implemented if time allows (because we really want them):

- **Groups** and group posting!
- Reputation-based 'slow' federation.
- Community decision making for federation and moderation actions.
- User-selectable custom templates for rendering public posts:
  - Twitter-style
  - Blogpost
  - Gallery
  - Etc.

## Getting Started

Proper documentation for running and maintaining GoToSocial will be forthcoming in the first release.

For now (if you want to run it pre-alpha, like a beast), check out the [quick and dirty getting started guide](https://docs.gotosocial.org/en/latest/installation_guide/quick_and_dirty/).

## Contributing

You wanna contribute to GtS? Great! ❤️❤️❤️ Check out the issues page to see if there's anything you wanna jump in on, and read the [CONTRIBUTING.md](./CONTRIBUTING.md) file for guidelines and setting up your dev environment.

## Contact

For questions and comments, you can [join our Matrix channel](https://matrix.to/#/#gotosocial:superseriousbusiness.org) at `#gotosocial:superseriousbusiness.org`. This is the quickest way to reach the devs. You can also mail [admin@gotosocial.org](mailto:admin@gotosocial.org).

For bugs and feature requests, please check to see if there's [already an issue](https://github.com/superseriousbusiness/gotosocial/issues), and if not, open one or use one of the above channels to make a request (if you don't have a Github account).

## Credits

### Libraries

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

### Developers

In alphabetical order:

- f0x \[[donate with liberapay](https://liberapay.com/f0x)\]
- kim
- tobi \[[donate with liberapay](https://liberapay.com/GoToSocial/)\]

### Special Thanks

Thanks to everyone who has used GtS, opened an issue, suggested something, given funding, and otherwise encouraged or supported the project!

## Sponsorship + Funding

Currently, this project is funded using Liberapay, to put bread on the table while work continues on it.

If you want to sponsor this project, you can do so [here](https://liberapay.com/GoToSocial/)! `<3`

**GoToSocial has NO CORPORATE SPONSORS and does not desire corporate sponsorship.**

## License

![the gnu AGPL logo](https://www.gnu.org/graphics/agplv3-155x51.png)

GoToSocial is free software, licensed under the [GNU AGPL v3 LICENSE](LICENSE). We encourage forking and changing the code, hacking around with it, and experimenting.

See [here](https://www.gnu.org/licenses/why-affero-gpl.html) for the differences between AGPL versus GPL licensing, and [here](https://www.gnu.org/licenses/gpl-faq.html) for FAQ's about GPL licenses, including the AGPL.

If you modify the GoToSocial source code, and run that modified code in a way that's accessible over a network, you *must* make your modifications to the source code available following the guidelines of the license:

> \[I\]f you modify the Program, your modified version must prominently offer all users interacting with it remotely through a computer network (if your version supports such interaction) an opportunity to receive the Corresponding Source of your version by providing access to the Corresponding Source from a network server at no charge, through some standard or customary means of facilitating copying of software.

Copyright (C) 2021 the GoToSocial Authors.
