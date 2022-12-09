# syntax=docker/dockerfile:1.3
# Dockerfile reference: https://docs.docker.com/engine/reference/builder/

# stage 1: generate up-to-date swagger.yaml to put in the final container
FROM --platform=${BUILDPLATFORM} quay.io/goswagger/swagger:v0.30.0 AS swagger

COPY go.mod /go/src/github.com/superseriousbusiness/gotosocial/go.mod
COPY go.sum /go/src/github.com/superseriousbusiness/gotosocial/go.sum
COPY cmd /go/src/github.com/superseriousbusiness/gotosocial/cmd
COPY internal /go/src/github.com/superseriousbusiness/gotosocial/internal
WORKDIR /go/src/github.com/superseriousbusiness/gotosocial
RUN swagger generate spec -o /go/src/github.com/superseriousbusiness/gotosocial/swagger.yaml --scan-models

# stage 2: generate the web/assets/dist bundles
FROM --platform=${BUILDPLATFORM} node:16.15.1-alpine3.15 AS bundler

COPY web web
RUN yarn install --cwd web/source && \
    BUDO_BUILD=1 node web/source  && \
    rm -r web/source

# stage 3: build the executor container
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
WORKDIR "/gotosocial"

# copy the dist binary created by goreleaser or build.sh
COPY --chown=1000:1000 gotosocial /gotosocial/gotosocial

# copy over the web directories with templates, assets etc
COPY --chown=1000:1000 --from=bundler web /gotosocial/web
COPY --chown=1000:1000 --from=swagger /go/src/github.com/superseriousbusiness/gotosocial/swagger.yaml web/assets/swagger.yaml

VOLUME [ "/gotosocial/storage" ]
ENTRYPOINT [ "/gotosocial/gotosocial", "server", "start" ]
