FROM golang:1.16.4-alpine3.13 AS builder
RUN apk update && apk upgrade --no-cache
RUN apk add git

# create build dir
RUN mkdir -p /go/src/github.com/superseriousbusiness/gotosocial
WORKDIR /go/src/github.com/superseriousbusiness/gotosocial

# move source files
ADD cmd /go/src/github.com/superseriousbusiness/gotosocial/cmd
ADD internal /go/src/github.com/superseriousbusiness/gotosocial/internal
ADD testrig /go/src/github.com/superseriousbusiness/gotosocial/testrig
ADD go.mod /go/src/github.com/superseriousbusiness/gotosocial/go.mod
ADD go.sum /go/src/github.com/superseriousbusiness/gotosocial/go.sum

# move .git dir and version for versioning
ADD .git /go/src/github.com/superseriousbusiness/gotosocial/.git
ADD version /go/src/github.com/superseriousbusiness/gotosocial/version

# move the build script
ADD build.sh /go/src/github.com/superseriousbusiness/gotosocial/build.sh

# do the build step
RUN ./build.sh

FROM alpine:3.13 AS executor
RUN apk update && apk upgrade --no-cache

# copy over the binary from the first stage
RUN mkdir -p /gotosocial/storage
COPY --from=builder /go/src/github.com/superseriousbusiness/gotosocial/gotosocial /gotosocial/gotosocial

# copy over the web directory with templates etc
COPY web /gotosocial/web

# make the gotosocial group and user
RUN addgroup -g 1000 gotosocial
RUN adduser -HD -u 1000 -G gotosocial gotosocial

# give ownership of the gotosocial dir to the new user
RUN chown -R gotosocial gotosocial /gotosocial

# become the user
USER gotosocial

WORKDIR /gotosocial
ENTRYPOINT [ "/gotosocial/gotosocial", "server", "start" ]
