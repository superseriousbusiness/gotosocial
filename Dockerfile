# syntax=docker/dockerfile:1.4
# Dockerfile reference: https://docs.docker.com/engine/reference/builder/

# stage 1: generate up-to-date swagger.yaml to put in the final container
FROM --platform=${BUILDPLATFORM} quay.io/goswagger/swagger:v0.30.0 AS swagger

WORKDIR /go/src/github.com/superseriousbusiness/gotosocial
COPY go.mod go.sum ./
COPY cmd ./
COPY internal ./
RUN swagger generate spec -o ./swagger.yaml --scan-models --exclude-deps


# stage 2: generate the web/assets/dist bundles
FROM --platform=${BUILDPLATFORM} node:16.15.1-alpine3.15 AS bundler

COPY web web
RUN yarn install --cwd web/source && \
    BUDO_BUILD=1 node web/source  && \
    rm -r web/source


# stage 3: build the application
FROM --platform=${BUILDPLATFORM} golang:1.19-alpine3.16 AS gobuild
# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
WORKDIR /go/src/github.com/superseriousbusiness/gotosocial
COPY go.mod go.sum vendor ./
RUN go mod download && go mod verify


COPY . .
# we need the built assets, so they can be bundled into the build.
COPY --from=bundler web ./
COPY --from=swagger /go/src/github.com/superseriousbusiness/gotosocial/swagger.yaml ./web/assets/

ARG VERSION=SNAPSHOT
RUN CGO_ENABLED=0 go build -v -tags=netgo,osusergo,static_build -ldflags='-s -w -extldflags -static -X main.Version=${VERSION}' -o gotosocial github.com/superseriousbusiness/gotosocial/cmd/gotosocial


# stage 4: use the executor image which will copy in the remaining assets
FROM --platform=${TARGETPLATFORM} alpine:3.15.4 as executor

# switch to non-root user:group for GtS
USER 1000:1000

# Because we're doing multi-arch builds we can't easily do `RUN mkdir [...]`
# but we can hack around that by having docker's WORKDIR make the dirs for
# us, as the user created above.
#
# See https://docs.docker.com/engine/reference/builder/#workdir
#
# First make sure storage exists + is owned by 1000:1000, then go back
# to just /gotosocial, where we'll run from
WORKDIR "/gotosocial/storage"
WORKDIR "/gotosocial/web/assets"
WORKDIR "/gotosocial/web/templates"
WORKDIR "/gotosocial"

# copy in the application. not now, but when this image is used in another Dockerfile's FROM statement.
# this makes it work both in the CI goreleaser buildcontext, and also when used in the standalone docker file
COPY --chown=1000:1000 --from=gobuild /go/src/github.com/superseriousbusiness/gotosocial/gotosocial /gotosocial/gotosocial

VOLUME [ "/gotosocial/storage", "/gotosocial/web" ]

# using CMD instead of ENTRYPOINT allows to start the container with sh for debugging work.
CMD [ "/gotosocial/gotosocial", "server", "start" ]
