ARG MATRIX_ARCH

FROM ghcr.io/arhat-dev/builder-go:alpine as builder

ARG MATRIX_ARCH

COPY . /app
RUN renovate-server golang local build renovate-server -m kernel=linux -m arch=${MATRIX_ARCH}

# TODO: support multiarch build
FROM mcr.microsoft.com/windows/servercore:ltsc2019

LABEL org.opencontainers.image.source https://github.com/arhat-dev/renovate-server

ARG MATRIX_ARCH
COPY --from=builder /app/build/renovate-server.linux.${MATRIX_ARCH} /renovate-server

ENTRYPOINT [ "/renovate-server" ]
