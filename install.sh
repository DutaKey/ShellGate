#!/bin/sh
set -e

REPO="DutaKey/ShellGate"
BINARY="shellgate"
INSTALL_DIR="/usr/local/bin"

# detect OS and arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" && exit 1 ;;
esac

# fetch latest release tag
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
if [ -z "$LATEST" ]; then
  echo "Could not determine latest release."
  exit 1
fi

FILENAME="${BINARY}_${OS}_${ARCH}"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${FILENAME}"

echo "Installing ShellGate ${LATEST} (${OS}/${ARCH})..."
curl -fsSL "$URL" -o "/tmp/${BINARY}"
chmod +x "/tmp/${BINARY}"

if [ -w "$INSTALL_DIR" ]; then
  mv "/tmp/${BINARY}" "${INSTALL_DIR}/${BINARY}"
else
  sudo mv "/tmp/${BINARY}" "${INSTALL_DIR}/${BINARY}"
fi

echo "ShellGate installed to ${INSTALL_DIR}/${BINARY}"
echo ""
echo "Get started:"
echo "  shellgate init"
echo "  shellgate login codex"
echo "  shellgate serve"
