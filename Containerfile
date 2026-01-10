# syntax=docker/dockerfile:1.8
ARG GO_VERSION=1.25
FROM golang:${GO_VERSION} AS builder

ENV DEBIAN_FRONTEND=noninteractive

# Install toolchains. We only need GPGME devs if we build CGO=1 Linux targets.
RUN set -eux; \
    arch="$(dpkg --print-architecture)"; \
    # enable cross-arch for optional cross-CGO builds
    if [ "$arch" = "arm64" ]; then dpkg --add-architecture amd64; fi; \
    if [ "$arch" = "amd64" ]; then dpkg --add-architecture arm64; fi; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
      ca-certificates \
      git \
      pkg-config \
      build-essential \
      gcc-aarch64-linux-gnu g++-aarch64-linux-gnu \
      gcc-x86-64-linux-gnu g++-x86-64-linux-gnu; \
    if [ "$arch" = "arm64" ]; then \
      # cross-build to linux/amd64 with CGO=1 (optional)
      apt-get install -y --no-install-recommends \
        libgpgme-dev:amd64 \
        libassuan-dev:amd64 \
        libgpg-error-dev:amd64 \
        libbtrfs-dev:amd64; \
    else \
      # native amd64 devs + optional arm64 devs for cross CGO builds
      apt-get install -y --no-install-recommends \
        libgpgme-dev \
        libassuan-dev \
        libgpg-error-dev \
        libbtrfs-dev \
        libgpgme-dev:arm64 \
        libassuan-dev:arm64 \
        libgpg-error-dev:arm64 \
        libbtrfs-dev:arm64; \
    fi; \
    update-ca-certificates; \
    rm -rf /var/lib/apt/lists/*

# Go proxy configuration (fallback to direct if proxy blocked)
ENV GOPROXY="https://proxy.golang.org,direct" \
    GOSUMDB="sum.golang.org"

WORKDIR /work

# Build script
COPY tools/build-all.sh /usr/local/bin/build-all.sh
RUN chmod +x /usr/local/bin/build-all.sh

# If a trust bundle is mounted at /etc/ssl/localcerts/trust-bundle.pem,
# build-all.sh will use it via SSL_CERT_FILE.
ENTRYPOINT ["/usr/local/bin/build-all.sh"]
