#!/usr/bin/env bash
set -euo pipefail

# Usage:
#   build-all.sh --bin rocket --pkg ./cmd/rocket --tag v1.2.3
#
# Notes:
# - Pure-Go targets (CGO=0) build with tags: "remote,containers_image_openpgp"
#   to avoid GPGME/cgo entirely.
# - linux/amd64 can be built with CGO=1 to keep full signing/GPGME features.
# - If your code directly imports github.com/proglottis/gpgme,
#   guard those files with //go:build cgo and provide !cgo stubs.

BIN_NAME=""
PKG_PATH=""
TAG="dev"

while [[ $# -gt 0 ]]; do
	case "$1" in
	--bin)
		BIN_NAME="$2"
		shift 2
		;;
	--pkg)
		PKG_PATH="$2"
		shift 2
		;;
	--tag)
		TAG="$2"
		shift 2
		;;
	*)
		echo "Unknown arg: $1" >&2
		exit 2
		;;
	esac
done

: "${BIN_NAME:?ERROR: --bin is required (e.g., --bin rocket)}"
: "${PKG_PATH:?ERROR: --pkg is required (e.g., --pkg ./cmd/rocket)}"

# Prefer a mounted trust bundle; else use system bundle.
MOUNTED_CA="/etc/ssl/localcerts/trust-bundle.pem"
SYSTEM_CA="/etc/ssl/certs/ca-certificates.crt"
if [[ -f "${MOUNTED_CA}" ]]; then
	export SSL_CERT_FILE="${MOUNTED_CA}"
	export GIT_SSL_CAINFO="${MOUNTED_CA}"
	export REQUESTS_CA_BUNDLE="${MOUNTED_CA}"
else
	export SSL_CERT_FILE="${SSL_CERT_FILE:-${SYSTEM_CA}}"
	export GIT_SSL_CAINFO="${GIT_SSL_CAINFO:-${SYSTEM_CA}}"
	export REQUESTS_CA_BUNDLE="${REQUESTS_CA_BUNDLE:-${SYSTEM_CA}}"
fi

echo "==> go version"
go version
echo "==> go mod download"
go mod download

# Default build matrix.
# Override by exporting TARGETS (same "GOOS GOARCH CGO" per line) if needed.
read -r -d '' DEFAULT_TARGETS <<'EOS' || true
linux amd64 1
darwin arm64 0
windows amd64 0
EOS

# darwin amd64 0
# linux arm64 0
# windows arm64 0

IFS=$'\n' read -r -d '' -a TARGETS <<<"${TARGETS_OVERRIDE:-${DEFAULT_TARGETS}}" || true

mkdir -p /out /work/build

for row in "${TARGETS[@]}"; do
	[[ -z "${row// /}" ]] && continue
	IFS=' ' read -r GOOS GOARCH CGO <<<"${row}"
	: "${GOOS:?internal: GOOS empty}"
	: "${GOARCH:?internal: GOARCH empty}"
	: "${CGO:?internal: CGO empty}"

	echo "==> Building ${GOOS}/${GOARCH} (CGO=${CGO})"
	export GOOS GOARCH CGO_ENABLED="${CGO}"

	# Per-target build tags
	BUILD_TAGS="remote"
	if [[ "${CGO}" == "0" ]]; then
		# Pure-Go: force containers/image to use pure-Go OpenPGP verifier (no GPGME)
		BUILD_TAGS="${BUILD_TAGS},containers_image_openpgp"
	fi
	# Allow callers to add more tags via EXTRA_TAGS env (comma-separated)
	if [[ -n "${EXTRA_TAGS:-}" ]]; then
		BUILD_TAGS="${BUILD_TAGS},${EXTRA_TAGS}"
	fi

	# Cross C compiler & pkg-config paths only when CGO is enabled
	case "${GOOS}/${GOARCH}/${CGO_ENABLED}" in
	linux/amd64/1)
		export CC=x86_64-linux-gnu-gcc
		export PKG_CONFIG_LIBDIR="/usr/lib/x86_64-linux-gnu/pkgconfig:/usr/share/pkgconfig"
		;;
	linux/arm64/1)
		export CC=aarch64-linux-gnu-gcc
		export PKG_CONFIG_LIBDIR="/usr/lib/aarch64-linux-gnu/pkgconfig:/usr/share/pkgconfig"
		;;
	*)
		unset CC
		unset PKG_CONFIG_LIBDIR
		;;
	esac

	OUT_BIN="/work/build/${BIN_NAME}-${GOOS}-${GOARCH}"
	: "${OUT_BIN:?internal: OUT_BIN empty}"

	go build -trimpath -tags="${BUILD_TAGS}" \
		-ldflags "-s -w -X main.version=${TAG}" \
		-o "${OUT_BIN}" "${PKG_PATH}"

	OUTPUT_PATH="/out/${BIN_NAME}-${GOOS}-${GOARCH}"
	if [[ "${GOOS}" == "windows" ]]; then
		OUTPUT_PATH="${OUTPUT_PATH}.exe"
	fi

	cp "${OUT_BIN}" "${OUTPUT_PATH}"

	if [[ "${GOOS}" != "windows" ]]; then
		chmod +x "${OUTPUT_PATH}"
	fi

	sha256sum "${OUTPUT_PATH}" >"${OUTPUT_PATH}.sha256"
	echo "==> Wrote: ${OUTPUT_PATH}"
done

echo "==> Packaging Router Configs"
STAGE="/work/build/pack/rocket-router-data"

mkdir -p "$STAGE"
TAR="/out/${BIN_NAME}-router-package.tar.gz"
if [[ -d "/work/resources" ]]; then cp -r "/work/resources" "${STAGE}/"; fi
[[ -f "/work/LICENSE" ]] && cp "/work/LICENSE" "${STAGE}/"
[[ -f "/work/README.md" ]] && cp "/work/README.md" "${STAGE}/"
tar --sort=name --owner=0 --group=0 --numeric-owner --mtime='UTC 2020-01-01' \
	-czf "${TAR}" -C "/work/build/pack" "rocket-router-data"
sha256sum "${TAR}" >"${TAR}.sha256"

echo "==> All artifacts in /out:"
ls -lh /out
