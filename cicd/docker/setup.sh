#!/bin/sh

case "${MATRIX_ROOTFS}" in
debian)
  apt-get update
  DEBIAN_FRONTEND=noninteractive \
    apt-get install -y --no-install-recommends \
    tzdata

  rm -rf /var/lib/apt/lists/*
  ;;
alpine)
  apk add --no-cache tzdata
  ;;
esac
