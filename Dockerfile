# syntax=docker/dockerfile:1.3
FROM --platform=${TARGETPLATFORM} alpine:3.15.0

# copy the dist binary created by goreleaser or build.sh
COPY --chown=1000:1000 gotosocial /gotosocial/gotosocial

# copy over the web directories with templates, assets etc
COPY --chown=1000:1000 web/assets /gotosocial/web/assets
COPY --chown=1000:1000 web/template /gotosocial/web/template

WORKDIR "/gotosocial"
ENTRYPOINT [ "/gotosocial/gotosocial", "server", "start" ]
