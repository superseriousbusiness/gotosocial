# activity

**THIS IS FORKED FROM https://github.com/go-fed/activity**!

> Complete ActivityStreams-based ontologies plus middleware handlers implementing ActivityPub

[![Build Status][Build-Status-Image]][Build-Status-Url] [![Go Reference][Go-Reference-Image]][Go-Reference-Url]
[![Go Report Card][Go-Report-Card-Image]][Go-Report-Card-Url] [![License][License-Image]][License-Url]
[![Chat][Chat-Image]][Chat-Url] [![OpenCollective][OpenCollective-Image]][OpenCollective-Url]

`go get github.com/go-fed/activity`

This repository contains two libraries and a tool:

* `astool`: A linked-data aware tool to generate golang native types for any
ActivityStreams vocabulary.
* `streams`: The ActivityStreams native types generated with the `astool`.
* `pub`: ActivityPub Social Protocol (Client-to-Server or C2S) and Federating
Protocol (Server-to-Server or S2S)

Check out [go-fed.org](https://go-fed.org/) for tutorials and documentation.

## Status

**1.0.0** ([Semantic Versioning](https://semver.org/))

This library has been successfully used to
[federate since May 17, 2019](https://cjslep.com/c/blog/this-blog-is-federated).

An [official implementation report](https://activitypub.rocks/implementation-report/)
was last submitted for version **0.2.0** [here](https://github.com/w3c/activitypub/issues/318).
Unfortunately, the official implementation report tool is no longer maintained.
Previous unofficial implementation reports are available in [issue #46](https://github.com/go-fed/activity/issues/46).

Please see CHANGELOG for changes between versions.

## Getting Started

Check out [go-fed.org](https://go-fed.org/) for tutorials and documentation.

Also, see `astool`, `streams`, or `pub` for their own README.

## FAQ

### What vocabularies are supported?

* [ActivityStreams](https://www.w3.org/TR/activitystreams-vocabulary).
* A subset of the [toot](https://github.com/tootsuite/mastodon/blob/master/app/lib/activitypub/adapter.rb) vocabulary.
* A subset of the [security](https://w3c-ccg.github.io/security-vocab/) vocabulary.
* [ForgeFed](https://forgefed.peers.community/vocabulary.html).

### How well tested are these libraries?

I took great care to add numerous tests using examples directly from
specifications, official test repositories, and my own end-to-end tests.

**v1.0.0** has around 200 unit tests. The **federation** or **S2S** portion of
the library is very well tested. The **social** or **C2S** portion could use
additional unit tests, but is far less popular than federation. About 70% of the
lines are covered by unit tests.

### Who is using this library currently?

Note: This list only includes those who have reached out to me to explicitly be
included.

| Application | Description                                       | Repository                                                                 | Point Of Contact                                                                                                    | Homepage                             |
|:-----------:|:-------------------------------------------------:|:--------------------------------------------------------------------------:|:-------------------------------------------------------------------------------------------------------------------:|:------------------------------------:|
| Anancus       | Self-hosted and federated social link aggregation              | [https://gitlab.com/tuxether/anancus](https://gitlab.com/tuxether/anancus)       | [@tuxether@floss.social](https://floss.social/@tuxether) or [tuxether@protonmail.ch](mailto:tuxether@protonmail.ch) | N/A                                                |
| WriteFreely   | Simple, open-source, privacy-focused blogging platform         | [https://github.com/writeas/writefreely](https://github.com/writeas/writefreely) | [@write_as@writing.exchange](https://writing.exchange/@write_as) or [hello@write.as](mailto:hello@write.as)         | [https://writefreely.org](https://writefreely.org) |
| Read.as       | Long-form reader built on open protocols                       | [https://github.com/writeas/Read.as](https://github.com/writeas/Read.as)         | [@write_as@writing.exchange](https://writing.exchange/@write_as) or [hello@write.as](mailto:hello@write.as)         | [https://read.as](https://read.as)                 |
| go-fed/apcore | Generic ActivityPub server framework in Go                     | [https://github.com/go-fed/apcore](https://github.com/go-fed/apcore)             | [@cj@mastodon.technology](https://mastodon.technology/@cj) or [cjslep@gmail.com](mailto:cjslep@gmail.com)           | [https://go-fed.org](https://go-fed.org)           |

### How do I use these libraries?

Check out [go-fed.org](https://go-fed.org/) for tutorials and documentation.

Please see each subdirectory for its own README for further elaboration.

### How can I get help, file issues, or contribute?

Please see the CONTRIBUTING.md file!

### Useful References

* [ActivityPub Specification](https://www.w3.org/TR/activitypub)
* [ActivityPub GitHub Repo](https://github.com/w3c/activitypub)
* [ActivityStreams Core Specification](https://www.w3.org/TR/activitystreams-core)
* [ActivityStreams Vocabulary Specification](https://www.w3.org/TR/activitystreams-vocabulary)
* [ActivityStreams GitHub Repo](https://github.com/w3c/activitystreams)

## Thanks

I would like to thank those that have worked hard to create the technologies
and standards that created the opportunity to implement this suite of
libraries.

Thanks to those who have been early adopters with v0 and/or provided early
feedback.

[Build-Status-Image]: https://github.com/go-fed/activity/workflows/build/badge.svg
[Build-Status-Url]: https://github.com/go-fed/activity/actions
[Go-Reference-Image]: https://pkg.go.dev/badge/github.com/go-fed/activity
[Go-Reference-Url]: https://pkg.go.dev/github.com/go-fed/activity
[Go-Report-Card-Image]: https://goreportcard.com/badge/github.com/go-fed/activity
[Go-Report-Card-Url]: https://goreportcard.com/report/github.com/go-fed/activity
[License-Image]: https://img.shields.io/github/license/go-fed/activity?color=blue
[License-Url]: https://opensource.org/licenses/BSD-3-Clause
[Chat-Image]: https://img.shields.io/matrix/go-fed:feneas.org?server_fqdn=matrix.org
[Chat-Url]: https://matrix.to/#/!BLOSvIyKTDLIVjRKSc:feneas.org?via=feneas.org&via=matrix.org
[OpenCollective-Image]: https://img.shields.io/opencollective/backers/go-fed-activitypub-labs
[OpenCollective-Url]: https://opencollective.com/go-fed-activitypub-labs
