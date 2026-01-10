# syntax=docker/dockerfile:1.8
ARG GO_VERSION=1.25
FROM golang:${GO_VERSION} AS builder

ENV DEBIAN_FRONTEND=noninteractive

WORKDIR /work
COPY tools/install-build-deps.sh /usr/local/bin/install-build-deps.sh
RUN chmod +x /usr/local/bin/install-build-deps.sh
RUN /usr/local/bin/install-build-deps.sh

COPY tools/build-binaries.sh /usr/local/bin/build-binaries.sh
RUN chmod +x /usr/local/bin/build-binaries.sh

ENTRYPOINT ["/usr/local/bin/build-binaries.sh"]
