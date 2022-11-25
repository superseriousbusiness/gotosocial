# syntax=docker/dockerfile:1.4
# Dockerfile reference: https://docs.docker.com/engine/reference/builder/

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

# copy in the application, containing all the assets in the binary.
COPY --chown=1000:1000 gotosocial /gotosocial/gotosocial

VOLUME [ "/gotosocial/storage", "/gotosocial/web" ]

# using CMD instead of ENTRYPOINT allows to start the container with sh for debugging work.
CMD [ "/gotosocial/gotosocial", "server", "start" ]
