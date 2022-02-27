# syntax=docker/dockerfile:1.3

# bundle the admin webapp
FROM --platform=${BUILDPLATFORM} node:17.6.0-alpine3.15 AS admin_builder
RUN apk update && apk upgrade --no-cache
RUN apk add git

RUN git clone https://github.com/superseriousbusiness/gotosocial-admin
WORKDIR /gotosocial-admin

RUN npm install
RUN node index.js

FROM --platform=${TARGETPLATFORM} alpine:3.15.0 AS executor

# copy over the binary from the first stage
COPY --chown=1000:1000 gotosocial /gotosocial/gotosocial

# copy over the web directory with templates etc
COPY --chown=1000:1000 web /gotosocial/web

# copy over the admin directory
COPY --chown=1000:1000 --from=admin_builder /gotosocial-admin/public /gotosocial/web/assets/admin

WORKDIR "/gotosocial"
ENTRYPOINT [ "/gotosocial/gotosocial", "server", "start" ]
