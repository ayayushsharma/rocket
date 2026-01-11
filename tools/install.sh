#!/usr/bin/env bash

OWNERNAME="ayayushsharma"
REPONAME="rocket"
APP_NAME="rocket"
APP_OS=""
APP_ARCH=""

abort() {
	printf "%s\n" "$@" >&2
	exit 1
}

# OS Checks
OS="$(uname)"
if [[ "${OS}" == "Linux" ]]; then
	APP_OS="linux"
elif [[ "${OS}" == "Darwin" ]]; then
	APP_OS="darwin"
else
	abort "Not configured to be used on OSs other than Mac and Linux. Download Binaries manually if available"
fi

# Architecture checks
ARCH="$(uname -m)"
if [[ "${ARCH}" == "x86_64" || "${ARCH}" == "amd64" ]]; then
	APP_ARCH="amd64"
elif [[ "${ARCH}" == "arm" || "${ARCH}" == "aarch64" || "${ARCH}" == "arm64" ]]; then
	APP_ARCH="arm64"
else
	abort "Not configured for architecture $ARCH. Download Binaries manually"
fi

APP_BINARY="${APP_NAME}-${APP_OS}-${APP_ARCH}"
APP_ROUTER_PACKAGE="${APP_NAME}-router-package"
APP_ROUTER_PACKAGE_TAR="${APP_ROUTER_PACKAGE}.tar.gz"
APP_CONFIG_DIR="${HOME}/.config/${APP_NAME}/"

# Removing preexiting files that collide with script downloads
if [ -f "./${APP_BINARY}" ]; then
	echo "=> Removing existing $APP_NAME binary > $APP_BINARY"
	rm -r "./${APP_BINARY}"
fi

if [ -f "./${APP_ROUTER_PACKAGE_TAR}" ]; then
	echo "=> Removing existing package for router"
	rm -r "./${APP_ROUTER_PACKAGE_TAR}"
fi

if [ -d "./${APP_ROUTER_PACKAGE}" ]; then
	echo "=> Removing existing extract for router"
	rm -r "./${APP_ROUTER_PACKAGE}"
fi

echo "==> Downloading Binary $APP_BINARY"
curl -LJO --progress-bar "https://github.com/$OWNERNAME/$REPONAME/releases/latest/download/${APP_BINARY}"

echo "==> Downloading Router $APP_ROUTER_PACKAGE"
curl -LJO --progress-bar "https://github.com/$OWNERNAME/$REPONAME/releases/latest/download/${APP_ROUTER_PACKAGE_TAR}"

echo "==> Extracting latest router package"
tar -xf "./${APP_ROUTER_PACKAGE_TAR}"

echo "==> Installing/Upgrading to latest router package"
cp -r ./${APP_ROUTER_PACKAGE}/resources/* $APP_CONFIG_DIR

echo "==> Installing/Upgrading $APP_NAME"
chmod +x ./${APP_BINARY}
cp ./${APP_BINARY} "${HOME}/.local/bin/${APP_NAME}"

echo "==> Syncing configurations"
./${APP_BINARY} sync

echo "==> Cleaning Up"
if [ -f "./${APP_BINARY}" ]; then
	echo "=> Removing existing $APP_NAME binary > $APP_BINARY"
	rm -r "./${APP_BINARY}"
fi

if [ -f "./${APP_ROUTER_PACKAGE_TAR}" ]; then
	echo "=> Removing existing package for router"
	rm -r "./${APP_ROUTER_PACKAGE_TAR}"
fi

if [ -d "./${APP_ROUTER_PACKAGE}" ]; then
	echo "=> Removing existing extract for router"
	rm -r "./${APP_ROUTER_PACKAGE}"
fi

echo

echo "Ensure you have podman installed and a podman machine running if you are not on linux"
echo

echo "Enjoy $APP_NAME!!!! Start with"
echo
echo "rocket register"
echo "rocket start"
echo "rocket launch --all"
echo 'open "http://app.localhost:32100"'
