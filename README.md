<!--overview-start-->
# GoToSocial <!-- omit in toc -->

**Update regarding corporate sponsors: we are open to sponsorship arrangements with organizations that align with our values; see the conditions below**

GoToSocial is an [ActivityPub](https://activitypub.rocks/) social network server, written in Golang.

With GoToSocial, you can keep in touch with your friends, post, read, and share images and articles. All without being tracked or advertised to!

<p align="middle">
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/sloth.webp" width="300"/>
</p>

**GoToSocial is still [BETA SOFTWARE](https://en.wikipedia.org/wiki/Software_release_life_cycle#Beta)**. It is already deployable and useable, and it federates cleanly with many other Fediverse servers (not yet all). However, many things are not yet implemented, and there are plenty of bugs! We left alpha stage around September/October 2024, and we intend to exit beta some time around 2026.

Documentation is at [docs.gotosocial.org](https://docs.gotosocial.org). You can skip straight to the API documentation [here](https://docs.gotosocial.org/en/latest/api/swagger/).

To build from source, check the [CONTRIBUTING.md](https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md) file.

Here's a screenshot of the instance landing page! Check out the project's [official account](https://gts.superseriousbusiness.org/@gotosocial) running on GoToSocial.

![Screenshot of the landing page for the GoToSocial instance goblin.technology. It shows basic information about the instance; number of users and posts etc.](https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/instancesplash.png)
<!--overview-end-->

## Table of Contents <!-- omit in toc -->

- [What is GoToSocial?](#what-is-gotosocial)
  - [Federation](#federation)
  - [History and Status](#history-and-status)
- [Features](#features)
  - [Mastodon API compatibility](#mastodon-api-compatibility)
  - [Granular post visibility settings](#granular-post-visibility-settings)
  - [Reply controls](#reply-controls)
  - [Local-only posting](#local-only-posting)
  - [RSS feed](#rss-feed)
  - [Rich text formatting](#rich-text-formatting)
  - [Themes and custom CSS](#themes-and-custom-css)
  - [Easy to run](#easy-to-run)
  - [Safety + security features](#safety--security-features)
  - [Various federation modes](#various-federation-modes)
  - [OIDC integration](#oidc-integration)
  - [Backend-first design](#backend-first-design)
- [Known Issues](#known-issues)
- [Installing GoToSocial](#installing-gotosocial)
  - [Supported Platforms](#supported-platforms)
    - [64-bit](#64-bit)
    - [BSDs](#bsds)
    - [32-bit](#32-bit)
    - [OpenBSD](#openbsd)
  - [Stable Releases](#stable-releases)
  - [Snapshot Releases](#snapshot-releases)
    - [Docker](#docker)
    - [Binary release .tar.gz](#binary-release-targz)
  - [From Source](#from-source)
  - [Third-party Packaging](#third-party-packaging)
- [Contributing](#contributing)
- [Contact](#contact)
- [Credits](#credits)
  - [Libraries](#libraries)
  - [Image Attribution and Licensing](#image-attribution-and-licensing)
  - [Team](#team)
  - [Special Thanks](#special-thanks)
- [Sponsorship + Funding](#sponsorship--funding)
  - [Crowdfunding](#crowdfunding)
  - [Corporate Sponsorship](#corporate-sponsorship)
  - [NLnet](#nlnet)
- [License](#license)

<!--body-1-start-->
## What is GoToSocial?

GoToSocial provides a lightweight, customizable, and safety-focused entryway into the [Fediverse](https://en.wikipedia.org/wiki/Fediverse), and is comparable to (but distinct from) existing projects such as [Mastodon](https://joinmastodon.org/), [Pleroma](https://pleroma.social/), [Friendica](https://friendi.ca), and [PixelFed](https://pixelfed.org/).

If you've ever used something like Twitter or Tumblr (or even Myspace!) GoToSocial will probably feel familiar to you: You can follow people and have followers, you make posts which people can favourite and reply to and share, and you scroll through posts from people you follow using a timeline. You can write long posts or short posts, or just post images, it's up to you. You can also, of course, block people or otherwise limit interactions that you don't want by posting just to your friends.

![Screenshot of the web view of a profile in GoToSocial, showing header and avatar, bio, and numbers of followers/following.](https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/profile1.png)

**GoToSocial does NOT use recommendation algorithms or collect data about you to suggest content or 'improve your experience'**. The timeline is chronological: whatever you see at the top of your timeline is there because it's *just been posted*, not because it's been selected as interesting (or controversial) based on your personal profile.

GoToSocial is not designed for 'must-follow' influencers with tens of thousands of followers, and it's not designed to be addictive. Your timeline and your experience are shaped by who you follow and how you interact with people, not by metrics of engagement!

GoToSocial doesn't claim to be *better* than any other application, but it offers something that might be better *for you* in particular.

### Federation

Because GoToSocial uses [ActivityPub](https://activitypub.rocks/), you can hang out not just with people on your home server, but with people all over the [Fediverse](https://en.wikipedia.org/wiki/Fediverse), seamlessly.

![the activitypub logo](https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/ap_logo.svg)

Federation means that your home server is part of a network of servers all over the world that all communicate using the same protocol. Your data is no longer centralized on one company's servers, but resides on your own server and is shared — as you see fit — across a resilient web of servers run by other people.

This federated approach also means that you aren't beholden to arbitrary rules from some gigantic corporation potentially thousands of miles away. Your server has its own rules and culture; your fellow server residents are your neighbors; you will likely get to know your server admins and moderators, or be an admin yourself.

GoToSocial advocates for many small, weird, specialist servers where people can feel at home, rather than a few big and generic ones where one person's voice can get lost in the crowd.

### History and Status

This project sprang up in February/March 2021 out of a dissatisfaction with the safety + privacy features of other Federated microblogging/social media applications, and a desire to implement something a little different.

It began as a solo project, and then picked up steam as more developers became interested and jumped on.

We made our first Alpha release in November 2021. We left Alpha and entered Beta in September/October 2024.

For a detailed view on what's implemented and what's not, and progress made towards [stable release](https://en.wikipedia.org/wiki/Software_release_life_cycle#Stable_release), please see [the roadmap document](https://github.com/superseriousbusiness/gotosocial/blob/main/ROADMAP.md).

---

## Features

### Mastodon API compatibility

The Mastodon API has become the de facto standard for client communication with federated servers, so GoToSocial has implemented and extended the API with custom functionality.

Though most apps that implement the Mastodon API should work, GoToSocial is tested and works reliably with beautiful apps like:

* [Tusky](https://tusky.app/) for Android
* [Pinafore](https://pinafore.social/) in the browser
* [Feditext](https://github.com/feditext/feditext) (beta) on iOS, iPadOS and macOS

If you've used Mastodon with a third-party app before, you'll find using GoToSocial a breeze.

### Granular post visibility settings

It's important that when you post something, you can choose who sees it.

GoToSocial offers public, unlisted/unlocked, followers-only, and direct posts (slide in DMs! -- with consent).

### Reply controls

GoToSocial lets you choose who can reply to your posts, via [interaction policies](https://docs.gotosocial.org/en/latest/user_guide/settings/#default-interaction-policies). You can choose to let anyone reply to your posts, let only your friends reply, and more.

![interaction policies settings](https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/user-settings-interaction-policy-1.png)

### Local-only posting

Sometimes you only want to talk to people you share an instance with. GoToSocial supports this via local-only posting, which ensures that your post stays on your instance only. (Local-only posting is currently dependent on client support.)

### RSS feed

GoToSocial lets you opt-in to exposing your profile as an RSS feed, so that people can subscribe to your public feed without missing a post.

### Rich text formatting

With GoToSocial, you can write posts using the popular, easy-to-use Markdown markup language, which lets you produce rich HTML posts with support for blockquotes, syntax-highlighted code blocks, lists, inline links, and more.

![markdown-formatted post](https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/markdown-post.png)

### Themes and custom CSS

Users can [choose from a variety of fun themes](https://docs.gotosocial.org/en/latest/user_guide/settings/#select-theme) for their profile, or even write their own [custom CSS](https://docs.gotosocial.org/en/latest/user_guide/settings/#custom-css).

It's also easy for admins to [add their own custom themes](https://docs.gotosocial.org/en/latest/admin/themes/) for users to choose from.

<details>
<summary>Show theme examples</summary>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-blurple-dark.png"/>
  <figcaption>Blurple dark</figcaption>
</figure>
<hr/>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-blurple-light.png"/>
  <figcaption>Blurple light</figcaption>
</figure>
<hr/>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-brutalist-light.png"/>
  <figcaption>Brutalist light</figcaption>
</figure>
<hr/>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-brutalist-dark.png"/>
  <figcaption>Brutalist dark</figcaption>
</figure>
<hr/>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-ecks-pee.png"/>
  <figcaption>Ecks pee</figcaption>
</figure>
<hr/>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-midnight-trip.png"/>
  <figcaption>Midnight trip</figcaption>
</figure>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-moonlight-hunt.png"/>
  <figcaption>Moonlight hunt</figcaption>
</figure>
<hr/>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-rainforest.png"/>
  <figcaption>Rainforest</figcaption>
</figure>
<hr/>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-soft.png"/>
  <figcaption>Soft</figcaption>
</figure>
<hr/>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-solarized-dark.png"/>
  <figcaption>Solarized dark</figcaption>
</figure>
<hr/>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-solarized-light.png"/>
  <figcaption>Solarized light</figcaption>
</figure>
<hr/>
<figure>
  <img src="https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/theme-sunset.png"/>
  <figcaption>Sunset</figcaption>
</figure>
<hr/>
</details>

### Easy to run

GoToSocial uses only about 250-350MiB of RAM, and requires very little CPU power, so it plays nice with single-board computers, old laptops and tiny $5/month VPSes.

![Grafana graph showing GoToSocial heap in use hovering around 250MB and spiking occasionally to 400MB-500MB.](https://raw.githubusercontent.com/superseriousbusiness/gotosocial/main/docs/overrides/public/getting-started-memory-graph.png)

No external dependencies apart from a database (or just use SQLite!).

Simply download the binary + assets (or Docker container), tweak your configuration, and run.

### Safety + security features

- Strict privacy enforcement for posts, and strict blocking logic.
- [Choose the visibility of posts on the web view of your profile](https://docs.gotosocial.org/en/latest/user_guide/settings/#visibility-level-of-posts-to-show-on-your-profile).
- [Import, export](https://docs.gotosocial.org/en/latest/admin/settings/#importexport), and [subscribe](https://docs.gotosocial.org/en/latest/admin/domain_permission_subscriptions) to community-created domain allow and domain block lists.
- HTTP signature authentication: GoToSocial requires [HTTP Signatures](https://datatracker.ietf.org/doc/html/draft-cavage-http-signatures-12) when sending and receiving messages, to ensure that your messages can't be tampered with and your identity can't be forged.
- Built-in, automatic support for secure HTTPS with [Let's Encrypt](https://letsencrypt.org/).

### Various federation modes

GoToSocial doesn't apply a one-size-fits-all approach to federation. Who your server federates with should be up to you.

- 'blocklist' mode (default): discover new servers; block servers you don't like.
- 'allowlist' mode (experimental); opt-in to federation with trusted servers.
- 'zero' federation mode; keep your server private (not yet implemented).

[See the docs for more info](https://docs.gotosocial.org/en/latest/admin/federation_modes).

### OIDC integration

GoToSocial supports [OpenID Connect (OIDC)](https://openid.net/connect/) identity providers, meaning you can integrate it with existing user management services like [Auth0](https://auth0.com/), [Gitlab](https://docs.gitlab.com/ee/integration/openid_connect_provider.html), etc., or run your own and hook GtS up to that (we recommend [Dex](https://dexidp.io/)).

### Backend-first design

Unlike other federated server projects, GoToSocial doesn't include an integrated client front-end (i.e., a web app).

Instead, like Matrix.org's [Synapse](https://github.com/matrix-org/synapse) project, it provides a relatively generic backend server implementation, some beautiful static pages for profiles and posts, and a [well-documented API](https://docs.gotosocial.org/en/latest/api/swagger/).

On top of this API, web developers are encouraged to build any front-end implementation or mobile application that they wish, whether Tumblr-like, Facebook-like, Twitter-like, or something else entirely.

---

## Known Issues

Since GoToSocial is still in beta, there are plenty of bugs. We use [GitHub issues](https://github.com/superseriousbusiness/gotosocial/issues?q=is%3Aissue+is%3Aopen+label%3Abug) to track these.

Since every ActivityPub server implementation has a slightly different interpretation of the protocol, some servers don't quite federate properly with GoToSocial yet. We're tracking these issues [in this project](https://github.com/superseriousbusiness/gotosocial/projects/4). Eventually, we want to make sure that any implementation that can federate nicely with Mastodon should also be able to federate with GoToSocial.

---

## Installing GoToSocial

Check our [getting started](https://docs.gotosocial.org/en/latest/getting_started/) documentation! And have a peruse of our [releases page](https://github.com/superseriousbusiness/gotosocial/releases).

<!--releases-start-->
### Supported Platforms

While we try to support a reasonable number of architectures and operating systems, it's not always possible to support a given platform due to library constraints or performance issues.

Platforms that we don't officially support *may* still work, but we can't test or guarantee performance or stability.

This is the current status of support offered by GoToSocial for different platforms (if something is unlisted it means we haven't checked yet so we don't know):

| OS      | Architecture            | Support level                             | Binary archive | Docker container |
| ------- | ----------------------- | ----------------------------------------- | -------------- | ---------------- |
| Linux   | x86-64/AMD64 (64-bit)   | 🟢 Full<sup>[1](#64-bit)</sup>            | Yes            | Yes              |
| Linux   | Armv8/ARM64  (64-bit)   | 🟢 Full<sup>[1](#64-bit)</sup>            | Yes            | Yes              |
| FreeBSD | x86-64/AMD64 (64-bit)   | 🟢 Full<sup>[1](#64-bit),[2](#bsds)</sup> | Yes            | No               |
| FreeBSD | Armv8/ARM64  (64-bit)   | 🟢 Full<sup>[1](#64-bit),[2](#bsds)</sup> | Yes            | No               |
| NetBSD  | x86-64/AMD64 (64-bit)   | 🟢 Full<sup>[1](#64-bit),[2](#bsds)</sup> | Yes            | No               |
| NetBSD  | Armv8/ARM64  (64-bit)   | 🟢 Full<sup>[1](#64-bit),[2](#bsds)</sup> | Yes            | No               |
| Linux   | x86-32/i386 (32-bit)    | 🟡 Partial<sup>[3](#32-bit)</sup>         | Yes            | Yes              |
| Linux   | Armv7/ARM32 (32-bit)    | 🟡 Partial<sup>[3](#32-bit)</sup>         | Yes            | Yes              |
| Linux   | Armv6/ARM32 (32-bit)    | 🟡 Partial<sup>[3](#32-bit)</sup>         | Yes            | Yes              |
| OpenBSD | Any                     | 🔴 None<sup>[4](#openbsd)</sup>           | No             | No               |

#### 64-bit

64-bit platforms require the following (now, very common) CPU features:

- x86-64 require SSE4.1 (for both media decoding and WASM SQLite)

- Armv8 require ARM64 Large System Extensions (specifically when using WASM SQLite)

Without these features, performance will suffer. In these situations, you may have some success building a binary yourself with the totally **unsupported, experimental** [nowasm](https://docs.gotosocial.org/en/latest/advanced/builds/nowasm/) tag.

#### BSDs

Mostly works, just [a few things to keep in mind](https://github.com/ncruces/go-sqlite3/wiki/Support-matrix) with WASM SQLite; check release notes carefully when installing on NetBSD or FreeBSD. If running with Postgres you should have no issues.

#### 32-bit

GtS doesn't work well on 32-bit systems like i386, or Armv6/v7, mainly due to performance of media decoding.

We don't recommend running GtS on 32-bit, but you may have some success either turning off remote media processing altogether, or building a binary yourself with the totally **unsupported, experimental** [nowasm](https://docs.gotosocial.org/en/latest/advanced/builds/nowasm/) tag.

For more guidance, check release notes when trying to install on 32-bit.

#### OpenBSD

Marked as unsupported due to performance issues (no WASM compiler, high memory usage when idle, crashes while processing media).

While we don't support running GtS on OpenBSD, you may have some success building a binary yourself with the totally **unsupported, experimental** [nowasm](https://docs.gotosocial.org/en/latest/advanced/builds/nowasm/) tag.

### Stable Releases

We package our stable releases for both binary builds and Docker containers, so that you don't have to build from source yourself.

The Docker image `superseriousbusiness/gotosocial:latest` will always correspond to the latest stable release. Since this tag is overwritten frequently, you may want to use Docker CLI flag `--pull always` to ensure that you always have the most up-to-date image every time you run using this tag. Alternatively, run `docker pull superseriousbusiness/gotosocial:latest` manually just before use.

### Snapshot Releases

We also make snapshot builds every time something is merged into the main branch, so you can run from whatever code is on main if you wish.

Please be warned that you do so at your own risk! We try to keep main working properly, but we make absolutely no guarantees. Take a stable release instead if you're unsure.

#### Docker

To run from main using Docker, use the `snapshot` Docker tag. The Docker image `superseriousbusiness/gotosocial:snapshot` will always correspond to the latest commit on main. Since this tag is overwritten frequently, you may want to use Docker CLI flag `--pull always` to ensure that you always have the most up-to-date image every time you run using this tag. Alternatively, run `docker pull superseriousbusiness/gotosocial:snapshot` manually just before use.

#### Binary release .tar.gz

To run from main using a binary release, download the appropriate .tar.gz file for your architecture from our [self-hosted Minio S3 repository](https://minio.s3.superseriousbusiness.org/browser/gotosocial-snapshots).

Snapshot binary releases in the S3 bucket are keyed by Github commit hash. To get the latest one, sort by Last Modified, or check out the list of commits [here](https://github.com/superseriousbusiness/gotosocial/commits/main), copy the SHA of the latest one, and paste it in the Minio console filter. Snapshot binary releases are expired after 28 days, to keep our hosting costs down.

### From Source

Instructions for building GoToSocial from source are in the [CONTRIBUTING.md](https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md) file.

### Third-party Packaging

Thank you so much to the cool people who have put time and energy into packaging GoToSocial! 

These packages are not maintained by GoToSocial, so please direct questions and issues to the repository maintainers (and donate to them!).

[![Packaging status](https://repology.org/badge/vertical-allrepos/gotosocial.svg)](https://repology.org/project/gotosocial/versions)

You can also deploy your own instance of GoToSocial with the help of:

- [YunoHost GoToSocial Packaging](https://github.com/YunoHost-Apps/gotosocial_ynh) by [OniriCorpe](https://github.com/OniriCorpe).
- [Ansible Playbook (MASH)](https://github.com/mother-of-all-self-hosting/mash-playbook): The playbook supports a many services, including GoToSocial. [Documentation](https://github.com/mother-of-all-self-hosting/mash-playbook/blob/main/docs/services/gotosocial.md)
- [GoToSocial Helm Chart](https://github.com/fSocietySocial/charts/tree/main/charts/gotosocial) by [0hlov3](https://github.com/0hlov3).

<!--releases-end-->
---

## Contributing

You would like to contribute to GtS? Great! ❤️❤️❤️ Check out the issues page to see if there's anything you intend to jump in on, and read the [CONTRIBUTING.md](https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md) file for guidelines and setting up your dev environment.

---

## Contact

For questions and comments, you can [join our Matrix space](https://matrix.to/#/#gotosocial-space:superseriousbusiness.org) at `#gotosocial-space:superseriousbusiness.org`. This is the quickest way to reach the devs. You can also mail [admin@gotosocial.org](mailto:admin@gotosocial.org).

For bugs and feature requests, please check to see if there's [already an issue](https://github.com/superseriousbusiness/gotosocial/issues), and if not, open one or use one of the above channels to make a request (if you don't have a Github account).

---

## Credits
<!--body-1-end-->

### Libraries

The following open source libraries, frameworks, and tools are used by GoToSocial, with gratitude 💕

- [buckket/go-blurhash](https://github.com/buckket/go-blurhash); used for generating image blurhashes. [GPL-3.0 License](https://spdx.org/licenses/GPL-3.0-only.html).
- [coreos/go-oidc](https://github.com/coreos/go-oidc); OIDC client library. [Apache-2.0 License](https://spdx.org/licenses/Apache-2.0.html).
- [DmitriyVTitov/size](https://github.com/DmitriyVTitov/size); runtime model memory size calculations. [MIT License](https://spdx.org/licenses/MIT.html).
- Gin:
  - [gin-contrib/cors](https://github.com/gin-contrib/cors); Gin CORS middleware. [MIT License](https://spdx.org/licenses/MIT.html).
  - [gin-contrib/gzip](https://github.com/gin-contrib/gzip); Gin gzip middleware. [MIT License](https://spdx.org/licenses/MIT.html).
  - [gin-contrib/sessions](https://github.com/gin-contrib/sessions); Gin sessions middleware. [MIT License](https://spdx.org/licenses/MIT.html).
  - [gin-gonic/gin](https://github.com/gin-gonic/gin); speedy router engine. [MIT License](https://spdx.org/licenses/MIT.html).
- [google/uuid](https://github.com/google/uuid); UUID generation. [BSD-3-Clause License](https://spdx.org/licenses/BSD-3-Clause.html).
- Go-Playground:
  - [go-playground/form](https://github.com/go-playground/form); funky form mapping support. [MIT License](https://spdx.org/licenses/MIT.html).
  - [go-playground/validator](https://github.com/go-playground/validator); struct validation. [MIT License](https://spdx.org/licenses/MIT.html).
- Gorilla:
  - [gorilla/feeds](https://github.com/gorilla/feeds); RSS + Atom feed generation. [BSD-2-Clause License](https://spdx.org/licenses/BSD-2-Clause.html).
  - [gorilla/websocket](https://github.com/gorilla/websocket); Websocket connectivity. [BSD-2-Clause License](https://spdx.org/licenses/BSD-2-Clause.html).
- [go-swagger/go-swagger](https://github.com/go-swagger/go-swagger); Swagger OpenAPI spec generation. [Apache-2.0 License](https://spdx.org/licenses/Apache-2.0.html).
- gruf:
  - [gruf/go-bytesize](https://codeberg.org/gruf/go-bytesize); byte size parsing / formatting. [MIT License](https://spdx.org/licenses/MIT.html).
  - [gruf/go-cache](https://codeberg.org/gruf/go-cache); LRU and TTL caches. [MIT License](https://spdx.org/licenses/MIT.html).
  - [gruf/go-debug](https://codeberg.org/gruf/go-debug); debug build tag. [MIT License](https://spdx.org/licenses/MIT.html).
  - [gruf/go-errors](https://codeberg.org/gruf/go-errors); context-like error w/ value wrapping  [MIT License](https://spdx.org/licenses/MIT.html).
  - [gruf/go-fastcopy](https://codeberg.org/gruf/go-fastcopy); performant (buffer pooled) I/O copying [MIT License](https://spdx.org/licenses/MIT.html).
  - [gruf/go-ffmpreg](https://codeberg.org/gruf/go-ffmpreg); embedded ffmpeg / ffprobe WASM binaries [GPL-3.0 License](https://spdx.org/licenses/GPL-3.0-only.html).
  - [gruf/go-kv](https://codeberg.org/gruf/go-kv); log field formatting. [MIT License](https://spdx.org/licenses/MIT.html).
  - [gruf/go-list](https://codeberg.org/gruf/go-list); generic doubly linked list. [MIT License](https://spdx.org/licenses/MIT.html).
  - [gruf/go-mutexes](https://codeberg.org/gruf/go-mutexes); safemutex & mutex map. [MIT License](https://spdx.org/licenses/MIT.html).
  - [gruf/go-runners](https://codeberg.org/gruf/go-runners); synchronization utilities. [MIT License](https://spdx.org/licenses/MIT.html).
  - [gruf/go-sched](https://codeberg.org/gruf/go-sched); task scheduler. [MIT License](https://spdx.org/licenses/MIT.html).
  - [gruf/go-storage](https://codeberg.org/gruf/go-storage); file storage backend (local & s3). [MIT License](https://spdx.org/licenses/MIT.html).
  - [gruf/go-structr](https://codeberg.org/gruf/go-structr); struct caching + queueing with automated indexing by field. [MIT License](https://spdx.org/licenses/MIT.html).
- jackc:
  - [jackc/pgconn](https://github.com/jackc/pgconn); Postgres driver. [MIT License](https://spdx.org/licenses/MIT.html).
  - [jackc/pgx](https://github.com/jackc/pgx); Postgres driver and toolkit. [MIT License](https://spdx.org/licenses/MIT.html).
- [KimMachineGun/automemlimit](https://github.com/KimMachineGun/automemlimit); cgroups memory limit checking. [MIT License](https://spdx.org/licenses/MIT.html).
- [k3a/html2text](https://github.com/k3a/html2text); HTML-to-text conversion. [MIT License](https://spdx.org/licenses/MIT.html).
- [mcuadros/go-syslog](https://github.com/mcuadros/go-syslog); Syslog server library. [MIT License](https://spdx.org/licenses/MIT.html).
- [microcosm-cc/bluemonday](https://github.com/microcosm-cc/bluemonday); HTML user-input sanitization. [BSD-3-Clause License](https://spdx.org/licenses/BSD-3-Clause.html).
- [miekg/dns](https://github.com/miekg/dns); DNS utilities. [Go License](https://go.dev/LICENSE).
- [minio/minio-go](https://github.com/minio/minio-go); S3 client SDK. [Apache-2.0 License](https://spdx.org/licenses/Apache-2.0.html).
- [mitchellh/mapstructure](https://github.com/mitchellh/mapstructure); Go interface => struct parsing. [MIT License](https://spdx.org/licenses/MIT.html).
- [modernc.org/sqlite](https://gitlab.com/cznic/sqlite); cgo-free port of SQLite. [Other License](https://gitlab.com/cznic/sqlite/-/blob/master/LICENSE).
- [mvdan.cc/xurls](https://github.com/mvdan/xurls); URL parsing regular expressions. [BSD-3-Clause License](https://spdx.org/licenses/BSD-3-Clause.html).
- [oklog/ulid](https://github.com/oklog/ulid); sequential, database-friendly ID generation. [Apache-2.0 License](https://spdx.org/licenses/Apache-2.0.html).
- [open-telemetry/opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go); OpenTelemetry API + SDK. [Apache-2.0 License](https://spdx.org/licenses/Apache-2.0.html).
- spf13:
  - [spf13/cobra](https://github.com/spf13/cobra); command-line tooling. [Apache-2.0 License](https://spdx.org/licenses/Apache-2.0.html).
  - [spf13/viper](https://github.com/spf13/viper); configuration management. [Apache-2.0 License](https://spdx.org/licenses/Apache-2.0.html).
- [stretchr/testify](https://github.com/stretchr/testify); test framework. [MIT License](https://spdx.org/licenses/MIT.html).
- superseriousbusiness:
  - [superseriousbusiness/activity](https://github.com/superseriousbusiness/activity) forked from [go-fed/activity](https://github.com/go-fed/activity); Golang ActivityPub/ActivityStreams library. [BSD-3-Clause License](https://spdx.org/licenses/BSD-3-Clause.html).
  - [superseriousbusiness/exif-terminator](https://codeberg.org/superseriousbusiness/exif-terminator); EXIF data removal. [GNU AGPL v3 LICENSE](https://spdx.org/licenses/AGPL-3.0-or-later.html).
  - [superseriousbusiness/httpsig](https://github.com/superseriousbusiness/httpsig) forked from [go-fed/httpsig](https://github.com/go-fed/httpsig); secure HTTP signature library. [BSD-3-Clause License](https://spdx.org/licenses/BSD-3-Clause.html).
  - [superseriousbusiness/oauth2](https://github.com/superseriousbusiness/oauth2) forked from [go-oauth2/oauth2](https://github.com/go-oauth2/oauth2); OAuth server framework and token handling. [MIT License](https://spdx.org/licenses/MIT.html).
- [temoto/robotstxt](https://github.com/temoto/robotstxt); robots.txt parsing. [MIT License](https://spdx.org/licenses/MIT.html).
- [tdewolff/minify](https://github.com/tdewolff/minify); HTML minification for Markdown-submitted posts. [MIT License](https://spdx.org/licenses/MIT.html).
- [uber-go/automaxprocs](https://github.com/uber-go/automaxprocs); GOMAXPROCS automation. [MIT License](https://spdx.org/licenses/MIT.html).
- [ulule/limiter](https://github.com/ulule/limiter); http rate limit middleware. [MIT License](https://spdx.org/licenses/MIT.html).
- [uptrace/bun](https://github.com/uptrace/bun); database ORM. [BSD-2-Clause License](https://spdx.org/licenses/BSD-2-Clause.html).
- [wagslane/go-password-validator](https://github.com/wagslane/go-password-validator); password strength validation. [MIT License](https://spdx.org/licenses/MIT.html).
- [yuin/goldmark](https://github.com/yuin/goldmark); markdown parser. [MIT License](https://spdx.org/licenses/MIT.html).

<!--body-2-start-->
### Image Attribution and Licensing

Sloth logo by [Anna Abramek](https://abramek.art/).

<a rel="license" href="http://creativecommons.org/licenses/by-sa/4.0/"><img alt="Creative Commons License" style="border-width:0" src="https://i.creativecommons.org/l/by-sa/4.0/88x31.png" /></a><br />The GoToSocial sloth mascot is licensed under a <a rel="license" href="http://creativecommons.org/licenses/by-sa/4.0/">Creative Commons Attribution-ShareAlike 4.0 International License</a>.

The Creative Commons Attribution-ShareAlike 4.0 International License license applies specifically to the following files and subdirectories of this repository:

- [sloth logo png](https://github.com/superseriousbusiness/gotosocial/blob/main/web/assets/logo.png)
- [sloth logo webp](https://github.com/superseriousbusiness/gotosocial/blob/main/web/assets/logo.webp)
- [sloth logo svg](https://github.com/superseriousbusiness/gotosocial/blob/main/web/assets/logo.svg)
- [all default avatars](https://github.com/superseriousbusiness/gotosocial/blob/main/web/assets/default_avatars)

Under the terms of the license, you are free to:

- Share — copy and redistribute the abovementioned material in any medium or format.
- Adapt — remix, transform, and build upon the abovementioned material for any purpose, even commercially.

### Team

In alphabetical order (... and order of smell):

- daenney
- f0x \[[donate with liberapay](https://liberapay.com/f0x)\]
- kim \[check out my code @ [codeberg](https://codeberg.org/gruf), or find me @ [@kim](https://k.iim.gay/@kim)\]
- tobi \[[donate with liberapay](https://liberapay.com/GoToSocial/)\]
- vyr

### Special Thanks

A huge thank you to CJ from [go-fed](https://github.com/go-fed/activity): without your work, GoToSocial would not have been possible.

Thanks to everyone who has used GtS, opened an issue, suggested something, given funding, and otherwise encouraged or supported the project!

---

## Sponsorship + Funding

**Update regarding corporate sponsors: we are open to sponsorship arrangements with organizations that align with our values; see the conditions below**

### Crowdfunding

![open collective Standard Sloth badge](https://opencollective.com/gotosocial/tiers/standard-sloth/badge.svg?label=Standard%20Sloth&color=brightgreen) ![open collective Stable Sloth badge](https://opencollective.com/gotosocial/tiers/stable-sloth/badge.svg?label=Stable%20Sloth&color=green) ![open collective Special Sloth badge](https://opencollective.com/gotosocial/tiers/special-sloth/badge.svg?label=Special%20Sloth&color=yellowgreen) ![open collective Sugar Sloth badge](https://opencollective.com/gotosocial/tiers/sugar-sloth/badge.svg?label=Sugar%20Sloth&color=blue)

If you would like to donate to GoToSocial to keep the lights on during development, [you can do so via our OpenCollective page](https://opencollective.com/gotosocial#support)!

![LiberaPay patrons](https://img.shields.io/liberapay/patrons/GoToSocial.svg?logo=liberapay) ![receives via LiberaPay](https://img.shields.io/liberapay/receives/GoToSocial.svg?logo=liberapay)

If you prefer, we also have an account on LiberaPay! You can find that [right here](https://liberapay.com/GoToSocial/).

Crowdfunded donations to our OpenCollective and Liberapay accounts go towards paying the core team, paying server costs, and paying for GtS art, design, and other bits and bobs.

💕 🦥 💕 Thank you!


### Corporate Sponsorship

GoToSocial is open to sponsorship arrangements with organizations that align with our values. To show our thanks for your support, we will display your logo, website, and a short tagline on the repository and documentation. The following caveats apply to sponsorships:

1. GoToSocial project direction will always remain 100% in the hands of the core team, and will never be dictated or influenced by corporate sponsorship. This is non-negotiable. Corporations are of course free of course to suggest / request features in the same manner as any other user, but they are not given special treatment.

2. Corporate sponsorship is dependent on your organization meeting our team's ethical guidelines. These are not a concrete set of rules but instead boil down to "is your company causing harm?". For example, those in the defense industry need not apply as the simple answer to that question is, "yes!".

If after reading this you are still interested in supporting us, that is wonderful! Please reach out to us at admin@gotosocial.org to further discuss :)

### NLnet

<img src="https://nlnet.nl/logo/NGI/NGIZero-green.hex.svg" width="75" alt="NGIZero logo"/>

Combined with the above crowdfunding sources, 2023 Alpha development of GoToSocial was funded by a 50,000 EUR grant from the [NGI0 Entrust Fund](https://nlnet.nl/entrust/), via [NLnet](https://nlnet.nl/). See [here](https://nlnet.nl/project/GoToSocial/#ack) for more details. The successful grant application is archived [here](https://github.com/superseriousbusiness/gotosocial/blob/main/archive/nlnet/2022-next-generation-internet-zero.md).

2024 Beta development of GoToSocial is being funded by an additional 50,000 EUR grant from the [NGI0 Entrust Fund](https://nlnet.nl/entrust/), via [NLnet](https://nlnet.nl/).

---

## License

![the gnu AGPL logo](https://www.gnu.org/graphics/agplv3-155x51.png)

GoToSocial is free software, licensed under the [GNU AGPL v3 LICENSE](https://github.com/superseriousbusiness/gotosocial/blob/main/LICENSE). We encourage forking and changing the code, hacking around with it, and experimenting.

See [here](https://www.gnu.org/licenses/why-affero-gpl.html) for the differences between AGPL versus GPL licensing, and [here](https://www.gnu.org/licenses/gpl-faq.html) for FAQ's about GPL licenses, including the AGPL.

If you modify the GoToSocial source code, and run that modified code in a way that's accessible over a network, you *must* make your modifications to the source code available following the guidelines of the license:

> \[I\]f you modify the Program, your modified version must prominently offer all users interacting with it remotely through a computer network (if your version supports such interaction) an opportunity to receive the Corresponding Source of your version by providing access to the Corresponding Source from a network server at no charge, through some standard or customary means of facilitating copying of software.

Copyright (C) GoToSocial Authors

<!--I'm adding this here to take the crown of having the 1000th commit ~ kim-->
<!--body-2-end-->
