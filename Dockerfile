# syntax=docker/dockerfile:1.3
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

RUN mkdir -p /gotosocial/data & mkdir -p /gotosocial/storage \
    & chown 1000:1000 -R /gotosocial

# copy the dist binary created by goreleaser or build.sh
COPY --chown=1000:1000 gotosocial /gotosocial/gotosocial
COPY --chown=1000:1000 scripts/quick-start.sh /gotosocial/quick-start.sh

# copy over the web directories with templates, assets etc
COPY --chown=1000:1000 --from=bundler web /gotosocial/web
COPY --chown=1000:1000 --from=swagger /go/src/github.com/superseriousbusiness/gotosocial/swagger.yaml web/assets/swagger.yaml

WORKDIR /gotosocial
VOLUME ["/gotosocial/data", "/gotosocial/storage"]
EXPOSE 8080

ENTRYPOINT [ "/gotosocial/gotosocial", "server", "start" ]
