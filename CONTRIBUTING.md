# Contribution Guidelines <!-- omit in toc -->

Hey! Welcome to the CONTRIBUTING.md for GoToSocial :) Thanks for taking a look, that kicks ass.

These contribution guidelines were adapted from / inspired by those of Gitea (https://github.com/go-gitea/gitea/blob/main/CONTRIBUTING.md). Thanks Gitea!

## Table of Contents  <!-- omit in toc -->

- [Introduction](#introduction)
- [Bug reports and feature requests](#bug-reports-and-feature-requests)
- [Pull requests](#pull-requests)
  - [Code](#code)
  - [Documentation](#documentation)
- [Development](#development)
  - [Golang forking quirks](#golang-forking-quirks)
  - [Building GoToSocial](#building-gotosocial)
    - [Binary](#binary)
    - [Docker](#docker)
      - [With GoReleaser](#with-goreleaser)
      - [Manually](#manually)
  - [Stylesheet / Web dev](#stylesheet--web-dev)
    - [Live Reloading](#live-reloading)
  - [Project Structure](#project-structure)
    - [Finding your way around the code](#finding-your-way-around-the-code)
  - [Style / Linting / Formatting](#style--linting--formatting)
  - [Testing](#testing)
    - [Standalone Testrig with Pinafore](#standalone-testrig-with-pinafore)
    - [Running automated tests](#running-automated-tests)
      - [SQLite](#sqlite)
      - [Postgres](#postgres)
    - [CLI Tests](#cli-tests)
  - [Updating Swagger docs](#updating-swagger-docs)
  - [CI/CD configuration](#cicd-configuration)
  - [Release Checklist](#release-checklist)
    - [What if something goes wrong?](#what-if-something-goes-wrong)

## Introduction

This document contains important information that will help you to write a successful contribution to GoToSocial. Please read it carefully before opening a pull request!

## Bug reports and feature requests

Currently, we use Github's issue system for tracking bug reports and feature requests.

You can view all open issues [here](https://github.com/superseriousbusiness/gotosocial/issues "The Github Issues page for GoToSocial").

Before opening a new issue, whether bug or feature request, **please search carefully through both open and closed issues to make sure it hasn't been addressed already**. You can use Github's keyword issue search for this. If your issue is a duplicate of an existing issue, it will be closed.

Before you open a feature request issue, please consider the following:

- Does this feature fit within the scope of GoToSocial? Since we are a small team, we are wary of [feature creep](https://en.wikipedia.org/wiki/Feature_creep "Wikipedia article on Feature Creep"), which can cause maintenance issues.
- Will this feature be generally useful for many users of the software, or is it handy for only a very specific use case?
- Will this feature have a negative impact on the performance of the software? If so, is the tradeoff worth it?
- Does this feature require loosening API security restrictions in some way? If so, it will need a good justification.
- Does this feature belong in GoToSocial's server backend, or is it something that a client could implement instead?

We tend to prioritize feature requests related to accessibility, fedi interoperability, and client compatibility.

## Pull requests

We welcome pull requests from new and existing contributors, with the following provisos:

- You have read and agree to our Code of Conduct.
- The pull request addresses an existing issue or bug (please link to the relevant issue in your pull request), or is related to documentation.
- If your pull request introduces significant new code paths, you are willing to do some maintenance work on those code paths, and address bugs. We do not appreciate drive-by pull requests that introduce a significant maintenance burden!
- The pull request is of decent quality. We are a small team and unfortunately we don't have a lot of time to help shepherd pull requests, or help with basic coding questions. If you're unsure, don't bite off more than you can chew: start with a small feature or bugfix for your first PR, and work your way up.

If you have small questions or comments before/during the pull request process, you can [join our Matrix space](https://matrix.to/#/#gotosocial-space:superseriousbusiness.org "GoToSocial Matrix space") at `#gotosocial-space:superseriousbusiness.org`.

Please read the appropriate section below for the kind of pull request you plan to open.

### Code

To keep things manageable for maintainers, the process for opening pull requests against the GoToSocial repository works roughly as follows:

1. Open an issue for the feature, bug, or issue your pull request will address, or comment on an existing issue to let everyone know you want to work on it.
2. Use the open issue to discuss your design with us, gather feedback, and resolve any concerns about the implementation.
3. Write your code! Make sure all existing tests pass. Add tests where appropriate. Run linters and formatters. Update documentation.
4. Open your pull request. You can do this as a draft, if you want to gather more feedback on code-in-progress.
5. Let us know that your pull request is ready to be reviewed.
6. Wait for review.
7. Address review comments, make changes to the code where appropriate. It's OK to push back on review comments if you have a sensible reason--we're all learning, after all--but please do so with patience and grace.
8. Get your code merged, rejoice!

To make your code easier to review, try to split your pull request into sensibly-sized commits, but don't worry too much about making it totally perfect: we always squash merge anyways.

If your pull request ends up being massive, consider splitting it into smaller discrete pull requests to make it easier to review and reason about.

Make sure your pull request only contains code that's relevant to the feature you're trying to implement, or the bug you're trying to address. Don't include refactors of unrelated parts of the code in your pull request: make a separate PR for that!

If you open a code pull request without following the above process, we may close it and ask you to follow the process instead.

### Documentation

The process for documentation pull requests is a bit looser than the process for code.

If you see something in the documentation that's missing, wrong, or unclear, feel free to open a pull request addressing it; you don't necessarily need to open an issue first, but please explain why you're opening the PR in the PR comment.

## Development

### Golang forking quirks

One of the quirks of Golang is that it relies on the source management path being the same as the one used within `go.mod` and in package imports within individual Go files. This makes working with forks a bit awkward.

Let's say you fork GoToSocial to `github.com/yourgithubname/gotosocial`, and then clone that repository to `~/go/src/github.com/yourgithubname/gotosocial`. You will probably run into errors trying to run tests or build, so you might change your `go.mod` file so that the module is called `github.com/yourgithubname/gotosocial` instead of `github.com/superseriousbusiness/gotosocial`. But then this breaks all the imports within the project. Nightmare! So now you have to go through the source files and painstakingly replace `github.com/superseriousbusiness/gotosocial` with `github.com/yourgithubname/gotosocial`. This works OK, but when you decide to make a pull request against the original repo, all the changed paths are included! Argh!

The correct solution to this is to fork, then clone the upstream repository, then set `origin` of the upstream repository to that of your fork.

See [this blog post](https://blog.sgmansfield.com/2016/06/working-with-forks-in-go/) for more details.

In case this post disappears, here are the steps (slightly modified):

>
> Fork the repository on GitHub or set up whatever other remote git repo you will be using. In this case, I would go to GitHub and fork the repository.
>
> Now clone the upstream repo (not the fork):
>
> `mkdir -p ~/go/src/github.com/superseriousbusiness && git clone git@github.com:superseriousbusiness/gotosocial ~/go/src/github.com/superseriousbusiness/gotosocial`
>
> Navigate to the top level of the upstream repository on your computer:
>
> `cd ~/go/src/github.com/superseriousbusiness/gotosocial`
>
> Rename the current origin remote to upstream:
>
> `git remote rename origin upstream`
>
> Add your fork as origin:
>
> `git remote add origin git@github.com/yourgithubname/gotosocial`
>

### Building GoToSocial

#### Binary

To get started, you first need to have Go installed. GtS is currently using Go 1.19, so you should take that too. See [here](https://golang.org/doc/install) for installation instructions. **WARNING: Go version 1.19.4 does not work due to a bug in Go. Use 1.19.3 or 1.19.5+** instead.

Once you've got go installed, clone this repository into your Go path. Normally, this should be `~/go/src/github.com/superseriousbusiness/gotosocial`.

Once you've installed the prerequisites, you can try building the project: `./scripts/build.sh`. This will build the `gotosocial` binary.

If there are no errors, great, you're good to go!

For automatic re-compiling during development, you can use [nodemon](https://www.npmjs.com/package/nodemon):

```bash
nodemon -e go --signal SIGTERM --exec "go run ./cmd/gotosocial --host localhost testrig start || exit 1"
```

#### Docker

For both of the below methods, you need to have [Docker buildx](https://docs.docker.com/buildx/working-with-buildx/) installed.

##### With GoReleaser

GoToSocial uses the release tooling [GoReleaser](https://goreleaser.com/intro/) to make multiple-architecture + Docker builds simple.

GoReleaser is also used by GoToSocial for building and pushing Docker containers.

Normally, these processes are handled by Drone (see CI/CD above). However, you can also invoke GoReleaser manually for things like building snapshots.

To do this, first [install GoReleaser](https://goreleaser.com/install/).

Then install GoSwagger as described in [the Swagger section](#updating-swagger-docs).

Then install Node and Yarn as described in [Stylesheet / Web dev](#stylesheet--web-dev).

Finally, to create a snapshot build, do:

```bash
goreleaser --rm-dist --snapshot
```

If all goes according to plan, you should now have a number of multiple-architecture binaries and tars inside the `./dist` folder, and snapshot Docker images should be built (check your terminal output for version).

##### Manually

If you prefer a simple approach to building a Docker container, with fewer dependencies (go-swagger, Node, Yarn), you can also just build in the following way:

```bash
./scripts/build.sh && docker buildx build -t superseriousbusiness/gotosocial:latest .
```

The above command first builds the `gotosocial` binary, then invokes Docker buildx to build the container image.

If you want to build a docker image for a different CPU architecture without setting up buildx (for example for ARMv7 aka 32-bit ARM), first modify the Dockerfile by adding the following lines to the top (but don't commit this!):

```dockerfile
# When using buildx, these variables will be set by the tool:
# https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope
# However, declaring them as global build arguments like this allows them to be set manually with `--build-arg` instead. 
ARG BUILDPLATFORM
ARG TARGETPLATFORM
```

Then, you can use the following command:

```bash
GOOS=linux GOARCH=arm ./scripts/build.sh && docker build --build-arg BUILDPLATFORM=linux/amd64 --build-arg TARGETPLATFORM=linux/arm/v7 -t superseriousbusiness/gotosocial:latest .
```

See also: [exhaustive list of GOOS and GOARCH values](https://gist.github.com/lizkes/975ab2d1b5f9d5fdee5d3fa665bcfde6)

And: [exhaustive list of possible values for docker's `--platform`](https://github.com/tonistiigi/binfmt/#build-test-image)

### Stylesheet / Web dev

GoToSocial uses Gin templates in the `web/template` folder. Static assets are stored in `web/assets`. Source files for stylesheets and JS bundles (for frontend enhancement, and the settings interface) are stored in `web/source`, and bundled from there to the `web/assets/dist` folder (gitignored).

To bundle changes, you need [Node.js](https://nodejs.org/en/download/) and [Yarn](https://classic.yarnpkg.com/en/docs/install).

To install Yarn dependencies:

```bash
yarn install --cwd web/source
```

To recompile bundles:

```bash
node web/source
```

#### Live Reloading

For a more convenient development environment, you can run a livereloading version of the bundler alongside the [testrig](#testing).

Open two terminals, first start the testrig on port 8081:
``` bash
GTS_PORT=8081 go run ./cmd/gotosocial testrig start
```
Then start the bundler, which will run on port 8080, and proxy requests to the testrig instance where needed.
``` bash
NODE_ENV=development node web/source
```

The livereloading bundler *will not* change the bundled assets in `dist/`, so once you are finished with changes and want to deploy it somewhere, you have to run `node web/source` to generate production-ready bundles.

### Project Structure

For project structure, GoToSocial follows a standard and widely accepted project layout [defined here](https://github.com/golang-standards/project-layout). As the author writes:

> This is a basic layout for Go application projects. It's not an official standard defined by the core Go dev team; however, it is a set of common historical and emerging project layout patterns in the Go ecosystem.

Where possible, we prefer more files and packages of shorter length that very clearly pertain to definable chunks of application logic, rather than fewer but longer files: if one `.go` file is pushing 1,000 lines of code, it's probably too long.

#### Finding your way around the code

Most of the crucial business logic of the application is found inside the various packages and subpackages of the `internal` directory. Here's a brief summary of each of these:

`internal/ap` - ActivityPub utility functions and interfaces.

`internal/api` - Models, routers, and utilities for the client and federated (ActivityPub) APIs. This is where routes are attached to the router, and where you want to be if you're adding a route.

`internal/concurrency` - Worker models used by the processor and other queues.

`internal/config` - Code for configuration flags, CLI flag parsing, and getting/setting config.

`internal/db` - DB interfaces for interacting with sqlite/postgres databases. Database migration code is in `internal/db/bundb/migrations`.

`internal/email` - Email functionality, email sending via SMTP.

`internal/federation` - ActivityPub federation code; implements `go-fed` interfaces.

`internal/federation/federatingdb` - Implementation of `go-fed`'s Database interface.

`internal/federation/dereferencing` - Code for making http calls to fetch resources from remote instances.

`internal/gotosocial` - GoToSocial server startup/shutdown logic.

`internal/gtserror` - Error models.

`internal/gtsmodel` - Database and internal models. This is where `bundb` annotations live.

`internal/httpclient` - The HTTP client used by GoToSocial for making requests to remote resources.

`internal/id` - Code for generating IDs (ULIDs) for database models.

`internal/log` - Our logging implementation.

`internal/media` - Code related to managing + processing media attachments; images, video, emojis, etc.

`internal/messages` - Models for wrapping worker messages.

`internal/middleware` - Gin Gonic router middlewares: http signature checking, cache control, token checks, etc.

`internal/netutil` - HTTP / net request validation code.

`internal/oauth` - Wrapper code/interfaces for OAuth server implementation.

`internal/oidc` - Wrapper code/interfaces for OIDC claims and callbacks.

`internal/processing` - Logic for processing messages produced by the federation or client APIs. Much of the core business logic of GoToSocial is contained here.

`internal/regexes` - Regular expressions used for text parsing and matching of URLs, hashtags, mentions, etc.

`internal/router` - Wrapper for Gin HTTP router. Core HTTP logic is contained here. The router exposes functions for attaching routes, which are used by the code in `internal/api` handlers.

`internal/storage` - Wrapper for `codeberg.org/gruf/go-store` implementations. Local file storage and s3 logic goes here.

`internal/stream` - Websockets streaming logic.

`internal/text` - Text parsing and transformation. Status parsing logic is contained here -- both plain and markdown.

`internal/timeline` - Status timeline management code.

`internal/trans` - Code for exporting models to json backup files from the database, and importing backup json files into the database.

`internal/transport` - HTTP transport code and utilities.

`internal/typeutils` - Code for converting from internal database models to json, and back, or from ActivityPub format to internal database model format and vice versa. Basically, serdes.

`internal/uris` - Utilities for generating URIs used throughout GoToSocial.

`internal/util` - Odds and ends; small utility functions used by more than one package.

`internal/validate` - Model validation code -- currently not really used.

`internal/visibility` - Status visibility checking and filtering.

`internal/web` - Web UI handlers, specifically for serving web pages, the landing page, settings panels.

### Style / Linting / Formatting

It is a good idea to read the short official [Effective Go](https://golang.org/doc/effective_go) page before submitting code: this document is the foundation of many a style guide, for good reason, and GoToSocial more or less follows its advice.

Another useful style guide that we try to follow: [this one](https://github.com/bahlo/go-styleguide).

In addition, here are some specific highlights from Uber's Go style guide which agree with what we try to do in GtS:

- [Group Similar Declarations](https://github.com/uber-go/guide/blob/master/style.md#group-similar-declarations).
- [Reduce Nesting](https://github.com/uber-go/guide/blob/master/style.md#reduce-nesting).
- [Unnecessary Else](https://github.com/uber-go/guide/blob/master/style.md#unnecessary-else).
- [Local Variable Declarations](https://github.com/uber-go/guide/blob/master/style.md#local-variable-declarations).
- [Reduce Scope of Variables](https://github.com/uber-go/guide/blob/master/style.md#reduce-scope-of-variables).
- [Initializing Structs](https://github.com/uber-go/guide/blob/master/style.md#initializing-structs).

Before you submit any code, make sure to run `go fmt ./...` to update whitespace and other opinionated formatting.

We use [golangci-lint](https://golangci-lint.run/) for linting, which allows us to catch style inconsistencies and potential bugs or security issues, using static code analysis.

If you make a PR that doesn't pass the linter, it will be rejected. As such, it's good practice to run the linter locally before pushing or opening a PR.

To do this, first install the linter following the instructions [here](https://golangci-lint.run/usage/install/#local-installation).

Then, you can run the linter with:

```bash
golangci-lint run
```

If there's no output, great! It passed :)

### Testing

GoToSocial provides a [testrig](https://github.com/superseriousbusiness/gotosocial/tree/main/testrig) with a number of mock packages you can use in integration tests.

One thing that *isn't* mocked is the Database interface because it's just easier to use an in-memory SQLite database than to mock everything out.

#### Standalone Testrig with Pinafore

You can launch a testrig as a standalone server running at localhost, which you can connect to using something like [Pinafore](https://github.com/nolanlawson/pinafore).

To do this, first build the gotosocial binary with `./scripts/build.sh`.

Then, launch the testrig by invoking the binary as follows:

```bash
./gotosocial testrig start
```

To run Pinafore locally in dev mode, first clone the [Pinafore](https://github.com/nolanlawson/pinafore) repository, and then run the following command in the cloned directory:

```bash
yarn run dev
```

The Pinafore instance will start running on `localhost:4002`.

To connect to the testrig, navigate to `http://localhost:4002` and enter your instance name as `localhost:8080`.

At the login screen, enter the email address `zork@example.org` and password `password`. You will get a confirmation prompt. Accept, and you are logged in as Zork.

Note the following constraints:

- Since the testrig uses an in-memory database, the database will be destroyed when the testrig is stopped.
- If you stop the testrig and start it again, any tokens or applications you created during your tests will also be removed. As such, you need to log out and in again every time you stop/start the rig.
- The testrig does not make any actual external HTTP calls, so federation will not work from a testrig.

#### Running automated tests

Tests can be run against both SQLite and Postgres.

##### SQLite

If you would like to run tests as quickly as possible, using an SQLite in-memory database, use:

```bash
go test ./...
```

##### Postgres

If you want to run tests against a Postgres database on localhost, run:

```bash
GTS_DB_TYPE="postgres" GTS_DB_ADDRESS="localhost" go test -p 1 ./...
```

In the above command, it is assumed you are using the default Postgres password of `postgres`.

We set `-p 1` when running against Postgres because it requires tests to run in serial, not in parallel.

#### CLI Tests

In [./test/envparsing.sh](./test/envparsing.sh) there's a test for making sure that CLI flags, config, and environment variables get parsed as expected.

Although this test *is* part of the CI/CD testing process, you probably won't need to worry too much about running it yourself. That is, unless you're messing about with code inside the `main` package in `cmd/gotosocial`, or inside the `config` package in `internal/config`.

### Updating Swagger docs

GoToSocial uses [go-swagger](https://goswagger.io) to generate Swagger API documentation from code annotations.

You can install go-swagger following the instructions [here](https://goswagger.io/install.html).

If you change Swagger annotations on any of the API paths, you can generate a new Swagger file at `./docs/api/swagger.yaml` by running:

```bash
swagger generate spec --scan-models --exclude-deps -o docs/api/swagger.yaml
```

### CI/CD configuration

GoToSocial uses [Drone](https://www.drone.io/) for CI/CD tasks like running tests, linting, and building Docker containers.

These runs are integrated with GitHub, and will be run on opening a pull request or merging into main.

The Drone instance for GoToSocial is [here](https://drone.superseriousbusiness.org/superseriousbusiness/gotosocial).

The `drone.yml` file is [here](./.drone.yml) — this defines how and when Drone should run. Documentation for Drone is [here](https://docs.drone.io/).

It is worth noting that the `drone.yml` file must be signed by the Drone admin account to be considered valid. This must be done every time the file is changed. This is to prevent tampering and hijacking of the Drone instance. See [here](https://docs.drone.io/signature/).

To sign the file, first install and setup the [drone cli tool](https://docs.drone.io/cli/install/). Then, run:

```bash
drone -t PUT_YOUR_DRONE_ADMIN_TOKEN_HERE -s https://drone.superseriousbusiness.org sign superseriousbusiness/gotosocial --save
```

### Release Checklist

First things first: If this is a security hot-fix, we'll probably rush through this list, and make a prettier release a few days later.

Now, with that out of the way, here's our Checklist.

GoToSocial follows [Semantic Versioning](https://semver.org/).
So our first concern on the Checklist is:

- What version are we releasing?

Next we need to check:

- Do the assets have to be rebuilt and committed to the repository.
- Do the swagger docs have to be rebuilt?

On the project management side:

- Are there any issues that have to be moved to a different milestone?
- Are there any things on the [Roadmap](./ROADMAP.md) that can be ticked off?

Once we're happy with our Checklist, we can create the tag, and push it.
And the rest [is automation](./.drone.yml).

We can now head to GitHub, and add some personality to the release notes.
Finally, we make announcements on the all our channels that the release is out!

#### What if something goes wrong?

Sometimes things go awry.
We release a buggy release, we forgot something ­ something important.

If the release is so bad that it's unusable ­ or dangerous! ­ to a great part of our user-base, we can pull.
That is: Delete the tag.

Either way, once we've resolved the issue, we just start from the top of this list again. Version numbers are cheap. It's cheap to burn them.
