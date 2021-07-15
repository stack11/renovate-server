ARG MATRIX_ARCH

FROM ghcr.io/arhat-dev/builder-golang:1.16-alpine as builder

ARG MATRIX_ARCH

COPY . /app
RUN dukkha golang local build renovate-server \
    -m kernel=linux -m arch=${MATRIX_ARCH}

FROM scratch

LABEL org.opencontainers.image.source https://github.com/arhat-dev/renovate-server

ARG MATRIX_ARCH
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder \
    "/app/build/renovate-server.linux.${MATRIX_ARCH}" \
    /renovate-server

ENTRYPOINT [ "/renovate-server" ]
