ARG ARCH=amd64

FROM arhatdev/builder-go:alpine as builder
FROM arhatdev/go:alpine-${ARCH}

RUN apk add --no-cache tzdata

ARG APP=renovate-server

ENTRYPOINT [ "/renovate-server" ]
