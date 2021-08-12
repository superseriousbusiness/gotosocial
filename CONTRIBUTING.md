# Contributing

Hey! Welcome to the CONTRIBUTING.md for GoToSocial :) Thanks for taking a look, that kicks ass.

This document will expand as the project expands, so for now this is basically a stub.

Contributions are welcome at this point, since the API is fairly stable now and the structure is at least vaguely coherent.

Check the [issues](https://github.com/superseriousbusiness/gotosocial/issues) to see if there's anything you fancy jumping in on.

## Communications

Before starting on something, please comment on an issue to say that you're working on it, and/or drop into the GoToSocial Matrix room [here](https://matrix.to/#/#gotosocial:superseriousbusiness.org).

This is the recommended way of keeping in touch with other developers, asking direct questions about code, and letting everyone know what you're up to.

## Code of Conduct

In lieu of a fuller code of conduct, here are a few ground rules.

1. We *DO NOT ACCEPT* PRs from right-wingers, Nazis, transphobes, homophobes, racists, harassers, abusers, white-supremacists, misogynists, tech-bros of questionable ethics. If that's you, politely fuck off somewhere else.
2. Any PR that moves GoToSocial in the direction of surveillance capitalism or other bad fediverse behavior will be rejected.
3. Don't spam the chat too hard.

## Setting up your development environment

To get started, you first need to have Go installed. GTS was developed with Go 1.16.4, so you should take that too. See [here](https://golang.org/doc/install).

Once you've got go installed, clone this repository into your Go path. Normally, this should be `~/go/src/github.com/superseriousbusiness/gotosocial`.

Once that's done, you can try building the project: `./build.sh`. This will build the `gotosocial` binary.

If there are no errors, great, you're good to go!

To work with the stylesheet for templates, you need [Node.js](https://nodejs.org/en/download/), then run `yarn install` in `web/source/`. Recompiling the bundle.css is `node build.js` but can be automated with [nodemon](https://www.npmjs.com/package/nodemon) on file change: `nodemon -w style.css build.js`.

### Golang forking quirks

One of the quirks of Golang is that it relies on the source management path being the same as the one used within `go.mod` and in package imports within individual Go files. This makes working with forks a bit awkward.

Let's say you fork GoToSocial to `github.com/yourgithubname/gotosocial`, and then clone that repository to `~/go/src/github.com/yourgithubname/gotosocial`. You will probably run into errors trying to run tests or build, so you might change your `go.mod` file so that the module is called `github.com/yourgithubname/gotosocial` instead of `github.com/superseriousbusiness/gotosocial`. But then this breaks all the imports within the project. Nightmare! So now you have to go through the source files and painstakingly replace `github.com/superseriousbusiness/gotosocial` with `github.com/yourgithubname/gotosocial`. This works OK, but when you decide to make a pull request against the original repo, all the changed paths are included! Argh!

The correct solution to this is to fork, then clone the upstream repository, then set `origin` of the upstream repository to that of your fork.

See [this blogpost](https://blog.sgmansfield.com/2016/06/working-with-forks-in-go/) for more details.

In case this post disappears, here are the steps (slightly modified):

>
> Pull the original package from the canonical place with the standard go get command:
> 
> `go get github.com/superseriousbusiness/gotosocial`
> 
> Fork the repository on Github or set up whatever other remote git repo you will be using. In this case, I would go to Github and fork the repository.
> 
> Navigate to the top level of the repository on your computer. Note that this might not be the specific package you’re using:
> 
> `cd $GOPATH/src/github.com/superseriousbusiness/gotosocial`
> 
> Rename the current origin remote to upstream:
> 
> `git remote rename origin upstream`
> 
> Add your fork as origin:
> 
> `git remote add origin git@github.com/yourgithubname/gotosocial`
>

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

### Standalone Testrig

You can also launch a testrig as a standalone server running at localhost, which you can connect to using something like [Pinafore](https://github.com/nolanlawson/pinafore).

To do this, first build the gotosocial binary with `go build ./cmd/gotosocial`.

Then launch a clean Postgres container on localhost:

```bash
docker run -d --user postgres --network host -e POSTGRES_PASSWORD=postgres postgres
```

Then, launch the testrig by invoking the binary as follows:

```bash
./gotosocial --host localhost:8080 testrig start
```

To run Pinafore locally in dev mode, first clone the Pinafore repository, and run the following command in the cloned directory:

```bash
yarn run dev
```

The Pinafore instance will start running on `localhost:4002`.

To connect to the testrig, navigate to `https://localhost:4002` and enter your instance name as `localhost:8080`.

At the login screen, enter the email address `zork@example.org` and password `password`.

Note the following constraints:

- The testrig data will be destroyed when the testrig is destroyed. It does this by dropping all tables in Postgres on shutdown. As such, you should **NEVER RUN THE TESTRIG AGAINST A DATABASE WITH REAL DATA IN IT** because it will be destroyed. Be especially careful if you're forwarding database ports from a remote instance to your local machine, because you can easily and irreversibly nuke that data if you run the testrig against it.
- If you stop the testrig and start it again, any tokens or applications you created during your tests will also be removed. As such, you need to log out and in again every time you stop/start the rig.
- The testrig does not make any actual external http calls, so federation will (obviously) not work from a testrig.

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

Then make sure to run `go fmt ./...` to update whitespace and other opinionated formatting.

## Financial Compensation

Right now there's no structure in place for financial compensation for pull requests and code. This is simply because there's no money being made on the project apart from the very small weekly Liberapay donations.

If money starts coming in, I'll start looking at proper financial structures, but for now code is considered to be a donation in itself.
