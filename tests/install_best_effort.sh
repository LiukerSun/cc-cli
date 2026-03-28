#!/bin/bash

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
trap 'chmod -R u+w "$TMP_DIR" 2>/dev/null || true; rm -rf "$TMP_DIR"' EXIT

HOME_DIR="$TMP_DIR/home"
BIN_DIR="$HOME_DIR/.local/bin"
GO_BIN_DIR="$(dirname "$(command -v go)")"
mkdir -p "$HOME_DIR"

export HOME="$HOME_DIR"
export PATH="$GO_BIN_DIR:/usr/bin:/bin"
export CCC_INSTALL_BIN_DIR="$BIN_DIR"
export GOCACHE="$TMP_DIR/go-cache"
export GOMODCACHE="$TMP_DIR/go-mod-cache"
export XDG_CONFIG_HOME="$HOME_DIR/.config"
export XDG_DATA_HOME="$HOME_DIR/.local/share"
export XDG_CACHE_HOME="$HOME_DIR/.cache"
export XDG_STATE_HOME="$HOME_DIR/.local/state"

bash "$REPO_DIR/install.sh" > "$TMP_DIR/install-output.txt" 2>&1

if [ ! -x "$BIN_DIR/ccc" ]; then
    echo "expected ccc binary to be installed to $BIN_DIR/ccc" >&2
    cat "$TMP_DIR/install-output.txt" >&2
    exit 1
fi

bash "$REPO_DIR/install.sh" --uninstall > "$TMP_DIR/uninstall-output.txt" 2>&1

if [ -e "$BIN_DIR/ccc" ]; then
    echo "expected uninstall to remove $BIN_DIR/ccc" >&2
    cat "$TMP_DIR/uninstall-output.txt" >&2
    exit 1
fi

if ! grep -q "Config and data were left untouched." "$TMP_DIR/uninstall-output.txt"; then
    echo "expected uninstall output to mention preserved config/data" >&2
    cat "$TMP_DIR/uninstall-output.txt" >&2
    exit 1
fi

echo "install_best_effort.sh: ok"
