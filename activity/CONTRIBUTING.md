# Contributing

Whether you have a question, an issue, a feature request, or desire to help out
with the software engineering, `go-fed` welcomes you!

## Table of Contents

1. Help, I have a question!
2. Help, I found a bug!
3. Whoa, I have a great idea!
4. Beep boop, I want to contribute code!
5. FAQ
6. Contributors

## I have a question!

The issues section of the repositories is generally **not** the place to ask
questions. However, it is worth checking the
[existing issues](https://github.com/go-fed/activity/issues?q=is%3Aissue) to see
if an existing bug or feature request provides enough context to answer the
question.

For direct support, the best way to engage is to reach out on the Fediverse
(such as on [Mastodon](https://joinmastodon.org/)) to `@cj@mastodon.technology`.
That will be a direct communication to myself and will provide visibility to
others who are invested in the ActivityPub Fediverse.

In the future, there will be a website hosting better documentation and a
tutorial for this library. My apologies that it is not available at this time.

## I found a bug!

The issues section is made just for you! Please check the
[existing issues](https://github.com/go-fed/activity/issues?q=is%3Aissue) to see
if it has already been filed. If not, please file a new one with the
[`bug` issue label](https://github.com/go-fed/activity/issues/new?template=bug-report-template.md&labels=bug).

## I have a great idea!

The issues section is made just for you! Please check the 
[existing issues](https://github.com/go-fed/activity/issues?q=is%3Aissue) to see
if the idea has already been proposed. If not, please file a new one with the
[`feature request` issue label](https://github.com/go-fed/activity/issues/new?template=feature-request-template.md&labels=feature%20request).

## I want to contribute code!

Great! Please start participating in discussions on various bugs and feature
requests. For more casual discussions, reach out on the Fediverse at
`@cj@mastodon.technology`.

## FAQ

Here's a list of common or known issues.

### Do you accept contributors?

Yes!

### Compilation of `vocab` is tough on time and resources!

The `vocab` and `streams` packages are code generated on order of hundreds of
thousands to a million lines long. If using Go 1.9 or before, use `go install`
or `go build -i` to cache the build artifacts and do incremental builds.

Additionally, see [#42](https://github.com/go-fed/activity/issues/42).

### Can I financially support this effort?

Donations are strictly viewed as tips and not work-for-hire:

* [cjslep](https://liberapay.com/cj/)

## Contributors To This Repository

In order of first commit contribution.

* cjslep
* 21stio
