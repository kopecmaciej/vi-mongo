#!/bin/sh
set -eu

REPO="kopecmaciej/vi-mongo"
BIN_NAME="vi-mongo"

detect_os() {
	case "$(uname -s)" in
		Linux)  echo "Linux" ;;
		Darwin) echo "Darwin" ;;
		MINGW*|MSYS*|CYGWIN*) echo "Windows" ;;
		*) echo "unsupported" ;;
	esac
}

detect_arch() {
	case "$(uname -m)" in
		x86_64|amd64) echo "x86_64" ;;
		arm64|aarch64) echo "arm64" ;;
		i386|i686) echo "i386" ;;
		*) echo "unsupported" ;;
	esac
}

err() { printf 'error: %s\n' "$*" >&2; exit 1; }
info() { printf '%s\n' "$*"; }

OS=$(detect_os)
ARCH=$(detect_arch)
[ "$OS" = "unsupported" ] && err "unsupported OS: $(uname -s)"
[ "$ARCH" = "unsupported" ] && err "unsupported arch: $(uname -m)"

if [ "$OS" = "Windows" ]; then
	err "use a release artifact directly on Windows: https://github.com/${REPO}/releases"
fi

VERSION="${VI_MONGO_VERSION:-}"
if [ -z "$VERSION" ]; then
	info "fetching latest release tag…"
	VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
		| grep '"tag_name"' | head -n1 | cut -d'"' -f4)
	[ -n "$VERSION" ] || err "could not determine latest version"
fi

ASSET="vi-mongo_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"

if [ -n "${VI_MONGO_INSTALL_DIR:-}" ]; then
	INSTALL_DIR="$VI_MONGO_INSTALL_DIR"
elif [ -w "/usr/local/bin" ] 2>/dev/null; then
	INSTALL_DIR="/usr/local/bin"
elif [ -d "$HOME/.local/bin" ]; then
	INSTALL_DIR="$HOME/.local/bin"
else
	INSTALL_DIR="$HOME/.local/bin"
	mkdir -p "$INSTALL_DIR"
fi

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

info "downloading ${ASSET} (${VERSION})…"
curl -fsSL "$URL" -o "${TMP}/${ASSET}" || err "download failed: $URL"

info "extracting…"
tar -xzf "${TMP}/${ASSET}" -C "$TMP"

# goreleaser wraps the binary in a subdirectory named after the archive
BIN_PATH=$(find "$TMP" -name "$BIN_NAME" -type f | head -n1)
[ -n "$BIN_PATH" ] || err "binary not found in archive"
chmod +x "$BIN_PATH"

info "installing to ${INSTALL_DIR}…"
if [ -w "$INSTALL_DIR" ]; then
	mv "$BIN_PATH" "${INSTALL_DIR}/${BIN_NAME}"
else
	sudo mv "$BIN_PATH" "${INSTALL_DIR}/${BIN_NAME}"
fi

info ""
info "installed: ${INSTALL_DIR}/${BIN_NAME} (${VERSION})"
case ":$PATH:" in
	*":${INSTALL_DIR}:"*) ;;
	*) info "note: ${INSTALL_DIR} is not in your PATH" ;;
esac
