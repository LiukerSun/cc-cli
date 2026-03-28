#!/bin/bash

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

mkdir -p "$TMP_DIR/bin" "$TMP_DIR/home"

bash -lc '
command() {
    if [ "$1" = "-v" ] && [ "$2" = "node" ]; then
        return 1
    fi
    builtin command "$@"
}

HOME="$1" source "$2"
' _ "$TMP_DIR/home" "$REPO_DIR/install.sh" > "$TMP_DIR/output.txt" 2>&1 && {
    echo "expected installer to fail when node is missing" >&2
    cat "$TMP_DIR/output.txt" >&2
    exit 1
}

if ! grep -q "Node.js is required to install ccc" "$TMP_DIR/output.txt"; then
    echo "expected missing node error message" >&2
    cat "$TMP_DIR/output.txt" >&2
    exit 1
fi

if ! grep -q "Please install Node.js first, then rerun the installer." "$TMP_DIR/output.txt"; then
    echo "expected install guidance for missing node" >&2
    cat "$TMP_DIR/output.txt" >&2
    exit 1
fi

if ! grep -q "Recommended for macOS/Linux: install Node.js with nvm" "$TMP_DIR/output.txt"; then
    echo "expected nvm recommendation for missing node" >&2
    cat "$TMP_DIR/output.txt" >&2
    exit 1
fi

if ! grep -q "curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.4/install.sh | bash" "$TMP_DIR/output.txt"; then
    echo "expected nvm install command for missing node" >&2
    cat "$TMP_DIR/output.txt" >&2
    exit 1
fi

echo "install_requires_node.sh: ok"
