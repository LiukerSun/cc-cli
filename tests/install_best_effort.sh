#!/bin/bash

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

HOME_DIR="$TMP_DIR/home"
BIN_DIR="$HOME_DIR/.local/bin"
mkdir -p "$HOME_DIR"

HOME="$HOME_DIR" \
PATH="/usr/bin:/bin" \
CCC_INSTALL_BIN_DIR="$BIN_DIR" \
bash "$REPO_DIR/install.sh" > "$TMP_DIR/install-output.txt" 2>&1

if [ ! -x "$BIN_DIR/ccc" ]; then
    echo "expected ccc binary to be installed to $BIN_DIR/ccc" >&2
    cat "$TMP_DIR/install-output.txt" >&2
    exit 1
fi

HOME="$HOME_DIR" \
PATH="/usr/bin:/bin" \
CCC_INSTALL_BIN_DIR="$BIN_DIR" \
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
