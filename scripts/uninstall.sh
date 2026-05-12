#!/bin/sh
set -eu

BIN_NAME="vi-mongo"

err() { printf 'error: %s\n' "$*" >&2; exit 1; }
info() { printf '%s\n' "$*"; }

ask() {
	prompt="$1"
	printf '%s [y/N] ' "$prompt"
	read -r answer </dev/tty
	case "$answer" in
		y|Y|yes|YES) return 0 ;;
		*) return 1 ;;
	esac
}

BIN_PATH=$(command -v "$BIN_NAME" 2>/dev/null || true)

# Resolve actual paths from the binary so user config overrides (log path,
# XDG_CONFIG_HOME, macOS Library dir) are respected. Binary is removed last.
CONFIG_DIR=""
LOG_FILE=""
if [ -n "$BIN_PATH" ]; then
	PATHS_OUT=$("$BIN_PATH" --paths 2>/dev/null) || true
	CONFIG_FILE=$(printf '%s\n' "$PATHS_OUT" | grep '^Config:' | cut -d: -f2- | sed 's/^ *//')
	[ -n "$CONFIG_FILE" ] && CONFIG_DIR=$(dirname "$CONFIG_FILE")
	LOG_FILE=$(printf '%s\n' "$PATHS_OUT" | grep '^Log:' | cut -d: -f2- | sed 's/^ *//')
fi

# Fall back to OS-aware defaults when the binary is not available.
if [ -z "$CONFIG_DIR" ]; then
	case "$(uname -s)" in
		Darwin) CONFIG_DIR="$HOME/Library/Application Support/vi-mongo" ;;
		*)      CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/vi-mongo" ;;
	esac
fi
: "${LOG_FILE:=/tmp/vi-mongo.log}"

info "Vi-Mongo uninstall"
info ""

if [ -d "$CONFIG_DIR" ]; then
	if ask "Remove config directory ${CONFIG_DIR}?"; then
		rm -rf "$CONFIG_DIR"
		info "  removed ${CONFIG_DIR}"
	fi
fi

if [ -f "$LOG_FILE" ]; then
	if ask "Remove log file ${LOG_FILE}?"; then
		rm -f "$LOG_FILE"
		info "  removed ${LOG_FILE}"
	fi
fi

if [ -n "$BIN_PATH" ]; then
	if ask "Remove binary at ${BIN_PATH}?"; then
		if [ -w "$BIN_PATH" ]; then
			rm -f "$BIN_PATH"
		else
			sudo rm -f "$BIN_PATH"
		fi
		info "  removed ${BIN_PATH}"
	fi
else
	info "Binary not found on PATH (skipping)."
fi

info ""
info "Done."
