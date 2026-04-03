#!/usr/bin/env bash

set -euo pipefail

REPO="bentos-lab/parley"
APP_NAME="parley"
INSTALL_DIR="/usr/local/bin"

# Determine the OS (linux or darwin) so we can download the matching artifact.
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
  linux) OS="linux" ;;
  darwin) OS="darwin" ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

# Determine the architecture so we download the correct binary variant.
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported ARCH: $ARCH"
    exit 1
    ;;
esac

# Fetch the latest release tag from GitHub.
RELEASE_JSON="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest")"
VERSION="$(echo "$RELEASE_JSON" | grep -m1 'tag_name' | cut -d '"' -f4)"

FILENAME="${APP_NAME}-${VERSION}-${OS}-${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

# Download and unpack the archive into a temporary directory.
TMP_DIR=$(mktemp -d)
trap 'rm -rf "${TMP_DIR}"' EXIT

echo "Detect: ${OS}/${ARCH}"
echo "Latest version: ${VERSION}"
echo "⬇Downloading ${FILENAME}..."
curl -fsSL "$URL" -o "${TMP_DIR}/${FILENAME}"

echo "Extracting..."
tar -xzf "${TMP_DIR}/${FILENAME}" -C "${TMP_DIR}"

# Move the binary into the install directory, using sudo if required.
if [ -w "${INSTALL_DIR}" ]; then
  mv "${TMP_DIR}/${APP_NAME}" "${INSTALL_DIR}/${APP_NAME}"
else
  echo "Installing to ${INSTALL_DIR} (may prompt for sudo)..."
  sudo mv "${TMP_DIR}/${APP_NAME}" "${INSTALL_DIR}/${APP_NAME}"
fi

chmod +x "${INSTALL_DIR}/${APP_NAME}"

echo ""
echo "Installed ${APP_NAME} ${VERSION}!"
echo "Run: ${APP_NAME} --version"
