# syntax=docker/dockerfile:1.3
# Dockerfile reference: https://docs.docker.com/engine/reference/builder/

# stage 1: generate up-to-date swagger.yaml to put in the final container
FROM --platform=${BUILDPLATFORM} golang:1.22-alpine AS swagger

RUN \
    ### Installs goswagger for building swagger definitions inside this container
    go install "github.com/go-swagger/go-swagger/cmd/swagger@c46c303aaa02" && \
    # Makes swagger executable
    chmod +x /go/bin/swagger

COPY go.mod /go/src/github.com/superseriousbusiness/gotosocial/go.mod
COPY go.sum /go/src/github.com/superseriousbusiness/gotosocial/go.sum
COPY cmd /go/src/github.com/superseriousbusiness/gotosocial/cmd
COPY internal /go/src/github.com/superseriousbusiness/gotosocial/internal

WORKDIR /go/src/github.com/superseriousbusiness/gotosocial
RUN /go/bin/swagger generate spec -o /go/src/github.com/superseriousbusiness/gotosocial/swagger.yaml --scan-models

# stage 2: generate the web/assets/dist bundles
FROM --platform=${BUILDPLATFORM} node:18-alpine AS bundler

COPY web web
RUN yarn --cwd ./web/source install && \
    yarn --cwd ./web/source ts-patch install && \
    yarn --cwd ./web/source build   && \
    rm -rf ./web/source

# stage 3: build the executor container
FROM --platform=${TARGETPLATFORM} alpine:3.19.1 as executor

# switch to non-root user:group for GtS
USER 1000:1000

# Because we're doing multi-arch builds we can't easily do `RUN mkdir [...]`
# but we can hack around that by having docker's WORKDIR make the dirs for
# us, as the user created above.
#
# See https://docs.docker.com/engine/reference/builder/#workdir
#
# First make sure storage + cache exist and are owned by 1000:1000,
# then go back to just /gotosocial, where we'll actually run from.
WORKDIR "/gotosocial/storage"
WORKDIR "/gotosocial/.cache"
WORKDIR "/gotosocial"

# copy the dist binary created by goreleaser or build.sh
COPY --chown=1000:1000 gotosocial /gotosocial/gotosocial

# copy over the web directories with templates, assets etc
COPY --chown=1000:1000 --from=bundler web /gotosocial/web
COPY --chown=1000:1000 --from=swagger /go/src/github.com/superseriousbusiness/gotosocial/swagger.yaml web/assets/swagger.yaml

VOLUME [ "/gotosocial/storage", "/gotosocial/.cache" ]
ENTRYPOINT [ "/gotosocial/gotosocial", "server", "start" ]
