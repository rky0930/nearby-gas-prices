#!/usr/bin/env bash
set -euo pipefail

REPO="rky0930/nearby-gas-prices"
BIN="nearby-gas-prices"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported arch: $ARCH"; exit 1 ;;
esac

# Asset name convention (we'll match our GH Actions):
# nearby-gas-prices_<os>_<arch>.tar.gz
ASSET="${BIN}_${OS}_${ARCH}.tar.gz"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

LATEST_URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"

echo "Downloading: $LATEST_URL"

curl -fL "$LATEST_URL" -o "$TMPDIR/$ASSET"

tar -xzf "$TMPDIR/$ASSET" -C "$TMPDIR"

INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
  echo "Installing to $INSTALL_DIR (sudo required)"
  sudo install -m 0755 "$TMPDIR/$BIN" "$INSTALL_DIR/$BIN"
else
  echo "Installing to $INSTALL_DIR"
  install -m 0755 "$TMPDIR/$BIN" "$INSTALL_DIR/$BIN"
fi

echo "Done. Try: $BIN --help"
