# Contributing

Hey! Welcome to the CONTRIBUTING.md for GoToSocial :) Thanks for taking a look, that kicks ass.

This document will expand as the project expands, so for now this is basically a stub.

Contributions are welcome at this point, since the API is fairly stable now and the structure is at least vaguely coherent.

Check the [issues](https://github.com/superseriousbusiness/gotosocial/issues) to see if there's anything you fancy jumping in on.

Before starting on something, please comment on an issue to say that you're working on it, and send a message to `@dumpsterqueer@ondergrond.org` (Mastodon) to let them know.

## Setting up your development environment

To get started, you first need to have Go installed. GTS was developed with Go 1.16.4, so you should take that too. See [here](https://golang.org/doc/install).

Once you've got go installed, clone this repository into your Go path. Normally, this should be `~/go/src/github.com/superseriousbusiness/gotosocial`.

Once that's done, you can try building the project: `go build ./cmd/gotosocial`. This will build the `gotosocial` binary.

If there are no errors, great, you're good to go!

## Setting up your test environment

GoToSocial provides a [testrig](https://github.com/superseriousbusiness/gotosocial/tree/main/testrig) with a bunch of mock packages you can use in integration tests.

One thing that *isn't* mocked is the Database interface, because it's just easier to use a real Postgres database running on localhost.

You can spin up a Postgres easily using Docker:

```bash
docker run -d --user postgres --network host -e POSTGRES_PASSWORD=postgres postgres
```

If you want a nice interface for peeking at what's in the Postgres database during tests, use PGWeb:

```bash
docker run -d --user postgres --network host sosedoff/pgweb
```

This will launch a pgweb at `http://localhost:8081`.

## Running tests

Because the tests use a real Postgres under the hood, you can't run them in parallel, so you need to run tests with the following command:

```bash
go test -count 1 -p 1 ./...
```

The `count` flag means that tests will be run at least once, and the `-p 1` flag means that only 1 test will be run at a time.

## Linting

We use [golangci-lint](https://golangci-lint.run/) for linting. To run this locally, first install the linter following the instructions [here](https://golangci-lint.run/usage/install/#local-installation).

Then, you can run the linter with:

```bash
golangci-lint run
```

Note that this linter also runs as a job on the Github repo, so if you make a PR that doesn't pass the linter, it will be rejected. As such, it's good practice to run the linter locally before pushing or opening a PR.

Another useful linter is [golint](https://pkg.go.dev/github.com/360EntSecGroup-Skylar/goreporter/linters/golint), which catches some style issues that golangci-lint does not.

To install golint, run:

```bash
go get -u github.com/golang/lint/golint
```

To run the linter, use:

```bash
golint ./...
```

## Financial Compensation

Right now there's no structure in place for financial compensation for pull requests and code. This is simply because there's no money being made on the project apart from the very small weekly Liberapay donations.

If money starts coming in, I'll start looking at proper financial structures, but for now code is considered to be a donation in itself.
