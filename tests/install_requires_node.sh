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
bash "$REPO_DIR/install.sh" > "$TMP_DIR/output.txt" 2>&1

if ! grep -q "Building ccc from local checkout" "$TMP_DIR/output.txt"; then
    echo "expected installer to build from local checkout" >&2
    cat "$TMP_DIR/output.txt" >&2
    exit 1
fi

if grep -q "Node.js is required to install ccc" "$TMP_DIR/output.txt"; then
    echo "installer should no longer require Node.js" >&2
    cat "$TMP_DIR/output.txt" >&2
    exit 1
fi

if [ ! -x "$BIN_DIR/ccc" ]; then
    echo "expected ccc binary to be installed to $BIN_DIR/ccc" >&2
    cat "$TMP_DIR/output.txt" >&2
    exit 1
fi

if ! "$BIN_DIR/ccc" version | grep -q "ccc version"; then
    echo "expected installed ccc binary to run" >&2
    exit 1
fi

echo "install_requires_node.sh: ok"
