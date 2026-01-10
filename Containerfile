# syntax=docker/dockerfile:1.8
ARG GO_VERSION=1.25
FROM golang:${GO_VERSION} AS builder

ENV DEBIAN_FRONTEND=noninteractive

RUN set -eux; \
  arch="$(dpkg --print-architecture)"; \
  # Enable foreign arch if you truly need cross CGO builds later
  if [ "$arch" = "arm64" ]; then dpkg --add-architecture amd64; fi; \
  # NOTE: On amd64 runners, do NOT add arm64 dev packages for -dev libs (they conflict).
  apt-get update; \
  apt-get install -y --no-install-recommends \
    ca-certificates \
    git \
    pkg-config \
    build-essential \
    gcc-aarch64-linux-gnu g++-aarch64-linux-gnu \
    gcc-x86-64-linux-gnu g++-x86-64-linux-gnu; \
  if [ "$arch" = "arm64" ]; then \
    # Only on arm64 base images: install amd64 dev packages for cross CGO to linux/amd64
    apt-get install -y --no-install-recommends \
      libgpgme-dev:amd64 \
      libassuan-dev:amd64 \
      libgpg-error-dev:amd64 \
      libbtrfs-dev:amd64; \
  else \
    # On amd64 runner: install ONLY native dev packages (no :arm64 to avoid Conflicts)
    apt-get install -y --no-install-recommends \
      libgpgme-dev \
      libassuan-dev \
      libgpg-error-dev \
      libbtrfs-dev; \
  fi; \
  update-ca-certificates; \
  rm -rf /var/lib/apt/lists/*

WORKDIR /work
COPY tools/build-all.sh /usr/local/bin/build-all.sh
RUN chmod +x /usr/local/bin/build-all.sh
ENTRYPOINT ["/usr/local/bin/build-all.sh"]
